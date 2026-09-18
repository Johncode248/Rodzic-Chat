package aichat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	aichatpg "github.com/Johncode248/Rodzic-Chat/project/common/aichat"
	nfzpg "github.com/Johncode248/Rodzic-Chat/project/common/nfz"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/nfz"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/places"
)

// maxToolRounds caps how many times the assistant may call a tool in one Send -- a hard ceiling
// against a runaway loop (e.g. the model repeatedly trying slightly different search terms).
// Four rounds comfortably covers the expected chain (find_nfz_benefit -> search_nfz_queues,
// maybe once more with a different city/benefit) while keeping worst-case latency and Anthropic
// spend bounded.
const maxToolRounds = 4

var chatTools = []toolDef{
	{
		Name: "find_nfz_benefit",
		Description: "Wyszukuje w oficjalnym słowniku NFZ dokładne nazwy świadczeń (\"benefit\") pasujące do zapytania, np. \"alergolog dla dziecka\" -> \"PORADNIA ALERGOLOGICZNA DLA DZIECI\". " +
			"MUSISZ użyć tego narzędzia i wybrać dokładną nazwę z wyniku PRZED wywołaniem search_nfz_queues -- endpoint kolejek wymaga dokładnego dopasowania nazwy, przybliżona nazwa zwróci zero wyników. " +
			"Jeśli wynik jest pusty lub żadna pozycja nie pasuje do potrzeby rodzica, spróbuj innego, prostszego lub bardziej ogólnego zapytania (np. samo \"alergologiczna\" zamiast pełnego zdania).",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"query": {"type": "string", "description": "Fragment nazwy świadczenia po polsku, np. \"alergologiczna dla dzieci\", \"logopedyczna\", \"okulistyka\""}
			},
			"required": ["query"]
		}`),
	},
	{
		Name: "search_nfz_queues",
		Description: "Wyszukuje w oficjalnym API NFZ \"Terminy leczenia\" prognozowane czasy oczekiwania (PCUS) i placówki dla DOKŁADNEJ nazwy świadczenia zwróconej wcześniej przez find_nfz_benefit. " +
			"WAŻNE, wyjaśnij to rodzicowi jeśli o to zapyta: to NIE jest system rezerwacji wizyt -- nie ma tu konkretnych wolnych terminów/godzin, tylko prognoza NFZ w dniach/tygodniach oraz dane kontaktowe placówki (telefon, adres), aktualizowane raz w miesiącu. " +
			"To API NIE obejmuje zwykłych, rutynowych wizyt u pediatry w ramach POZ (podstawowej opieki zdrowotnej) -- tylko poradnie specjalistyczne, oddziały szpitalne i pracownie diagnostyczne. Jeśli rodzic pyta o zwykłego pediatrę/lekarza rodzinnego na już, użyj zamiast tego search_places (OpenStreetMap) albo wyjaśnij że taka wizyta zwykle nie wymaga kolejki NFZ.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"benefit": {"type": "string", "description": "Dokładna nazwa świadczenia NFZ, jedna z wyników find_nfz_benefit"},
				"province_code": {"type": "string", "description": "Dwucyfrowy kod województwa NFZ (01-16). Jeśli rodzic nie wskazał miasta/województwa, użyj domyślnego z kontekstu rozmowy podanego w prompcie systemowym."},
				"locality": {"type": "string", "description": "Nazwa miejscowości, jeśli rodzic ją podał (opcjonalnie, zawęża wyniki)"},
				"urgent": {"type": "boolean", "description": "true tylko jeśli sprawa jest pilna (przypadek pilny), w innym wypadku false/pomiń"},
				"child_age_band": {"type": "string", "enum": ["0_3", "4_9", "10_15", "16_18"], "description": "Przedział wiekowy dziecka, jeśli znany -- zawęża do placówek deklarujących obsługę tej grupy wiekowej"}
			},
			"required": ["benefit"]
		}`),
	},
	{
		Name:        "search_places",
		Description: "Wyszukuje zwykłe miejsca (nie NFZ) w OpenStreetMap -- przedszkola, place zabaw, zwykłych pediatrów/lekarzy rodzinnych przyjmujących prywatnie lub w ramach POZ, atrakcje dla dzieci itp. Zwraca nazwę, adres i współrzędne. Użyj tego, gdy pytanie nie dotyczy specjalistycznej kolejki NFZ.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"query": {"type": "string", "description": "Czego szukamy, np. \"pediatra Rybnik\" albo \"plac zabaw Katowice\""}
			},
			"required": ["query"]
		}`),
	},
}

// executeTool runs one tool_use block and returns its result as a JSON string for a tool_result
// block, plus whether it represents an error (Anthropic renders is_error results differently,
// nudging the model to try something else rather than treating the content as a real answer).
func executeTool(ctx context.Context, name string, input json.RawMessage) (result string, isError bool) {
	switch name {
	case "find_nfz_benefit":
		var args struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(input, &args); err != nil {
			return err.Error(), true
		}
		benefits, err := nfz.FindBenefit(ctx, args.Query)
		if err != nil {
			return err.Error(), true
		}
		if len(benefits) == 0 {
			return `{"benefits":[],"hint":"Brak dopasowań w słowniku NFZ -- spróbuj krótszego lub bardziej ogólnego zapytania."}`, false
		}
		return marshalOrError(map[string]any{"benefits": benefits})

	case "search_nfz_queues":
		var args struct {
			Benefit      string `json:"benefit"`
			ProvinceCode string `json:"province_code"`
			Locality     string `json:"locality"`
			Urgent       bool   `json:"urgent"`
			ChildAgeBand string `json:"child_age_band"`
		}
		if err := json.Unmarshal(input, &args); err != nil {
			return err.Error(), true
		}
		if args.Benefit == "" {
			return "benefit is required -- call find_nfz_benefit first", true
		}
		req := nfzpg.SearchQueuesRequest{
			Benefit:      args.Benefit,
			Province:     args.ProvinceCode,
			Locality:     args.Locality,
			ChildAgeBand: args.ChildAgeBand,
			Case:         nfzpg.CaseStable,
			Limit:        10,
		}
		if args.Urgent {
			req.Case = nfzpg.CaseUrgent
		}
		res, err := nfz.SearchQueues(ctx, req)
		if err != nil {
			return err.Error(), true
		}
		if len(res.Queues) == 0 {
			return `{"queues":[],"hint":"Brak wyników -- spróbuj innego województwa/miejscowości albo nieco innej nazwy świadczenia z find_nfz_benefit."}`, false
		}
		return marshalOrError(res)

	case "search_places":
		var args struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(input, &args); err != nil {
			return err.Error(), true
		}
		candidates, err := places.Search(ctx, args.Query)
		if err != nil {
			return err.Error(), true
		}
		return marshalOrError(map[string]any{"candidates": candidates})

	default:
		return fmt.Sprintf("unknown tool %q", name), true
	}
}

func marshalOrError(v any) (string, bool) {
	b, err := json.Marshal(v)
	if err != nil {
		return err.Error(), true
	}
	return string(b), false
}

// converseWithTools drives the whole tool-use loop for one Send: ask Claude, and as long as it
// keeps asking to call a tool (up to maxToolRounds), run that tool and feed the result back,
// until it produces a final text answer. history is the client-managed conversation so far,
// ending with the parent's new turn. tokensUsed sums every round's input+output tokens (see
// anthropicUsage.TotalTokens); this endpoint is hard-disabled server-side for MVP (see
// project/service/transport/aichat.Send), so nothing currently reads tokensUsed for enforcement.
func converseWithTools(ctx context.Context, systemPrompt string, history []aichatpg.ChatMessage) (reply string, tokensUsed int64, err error) {
	messages := make([]anthropicMessage, 0, len(history))
	for _, m := range history {
		messages = append(messages, anthropicMessage{Role: m.Role, Content: []contentBlock{{Type: "text", Text: m.Content}}})
	}

	for round := 0; round < maxToolRounds; round++ {
		resp, err := callAnthropicOnce(ctx, systemPrompt, messages, chatTools)
		if err != nil {
			return "", tokensUsed, err
		}
		tokensUsed += resp.Usage.TotalTokens()

		var textReply string
		var toolUses []contentBlock
		for _, block := range resp.Content {
			switch block.Type {
			case "text":
				textReply += block.Text
			case "tool_use":
				toolUses = append(toolUses, block)
			}
		}

		if resp.StopReason != "tool_use" || len(toolUses) == 0 {
			if textReply == "" {
				return "", tokensUsed, errors.New("aichat: Anthropic response had no text content")
			}
			return textReply, tokensUsed, nil
		}

		messages = append(messages, anthropicMessage{Role: "assistant", Content: resp.Content})

		results := make([]contentBlock, 0, len(toolUses))
		for _, tu := range toolUses {
			result, isErr := executeTool(ctx, tu.Name, tu.Input)
			results = append(results, contentBlock{Type: "tool_result", ToolUseId: tu.Id, Content: result, IsError: isErr})
		}
		messages = append(messages, anthropicMessage{Role: "user", Content: results})
	}

	return "", tokensUsed, fmt.Errorf("aichat: gave up after %d tool-use rounds without a final answer", maxToolRounds)
}
