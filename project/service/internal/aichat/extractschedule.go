package aichat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ScheduleEntry is one lesson slot extracted from a photographed plan zajęć -- see
// ExtractScheduleFromImage. Day is an English weekday name ("monday".."sunday") so the caller can
// map it straight onto RRULE BYDAY values without a Polish-name lookup table.
type ScheduleEntry struct {
	Day       string `json:"day"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Subject   string `json:"subject"`
}

type ScheduleExtraction struct {
	Entries []ScheduleEntry `json:"entries"`
}

const extractScheduleSystemPrompt = "Jesteś asystentem, który odczytuje zdjęcia planów zajęć (szkolnych/przedszkolnych/zajęć dodatkowych) i zwraca WYŁĄCZNIE surowy JSON (bez markdown, bez ```), dokładnie w tym kształcie:\n" +
	`{"entries":[{"day":"monday","start_time":"08:00","end_time":"08:45","subject":"Matematyka"}]}` + "\n\n" +
	"Zasady: \"day\" to zawsze jedna z angielskich nazw dni tygodnia pisana małymi literami (monday, tuesday, wednesday, thursday, friday, saturday, sunday), niezależnie od języka zdjęcia. \"start_time\"/\"end_time\" to czas w formacie 24-godzinnym HH:MM. \"subject\" to nazwa przedmiotu/zajęć dokładnie tak, jak jest napisana na zdjęciu (po polsku, jeśli zdjęcie jest po polsku). Jeśli jakiejś godziny końcowej brak, oszacuj rozsądnie 45 minut od startu. Pomiń puste/nieczytelne komórki zamiast zgadywać. Jeśli na zdjęciu w ogóle nie ma planu zajęć, zwróć {\"entries\":[]}."

// ExtractScheduleFromImage sends one photo to Claude (vision) and asks for a structured weekly
// schedule -- see project/service/transport/lessonplan for the HTTP handler this backs. A single,
// tool-free request: this only ever needs one round trip, unlike the assistant chat's
// converseWithTools (no NFZ/places lookups here).
func ExtractScheduleFromImage(ctx context.Context, mimeType string, base64Data string) (ScheduleExtraction, error) {
	messages := []anthropicMessage{
		{
			Role: "user",
			Content: []contentBlock{
				{Type: "image", Source: &imageSource{Type: "base64", MediaType: mimeType, Data: base64Data}},
				{Type: "text", Text: "Odczytaj plan zajęć z tego zdjęcia i zwróć JSON zgodnie z instrukcją."},
			},
		},
	}

	resp, err := callAnthropicOnce(ctx, extractScheduleSystemPrompt, messages, nil)
	if err != nil {
		return ScheduleExtraction{}, err
	}

	var text string
	for _, block := range resp.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}
	text = strings.TrimSpace(text)
	// Claude occasionally wraps JSON in a code fence despite the instruction not to -- strip it
	// rather than fail the whole extraction over formatting.
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var extraction ScheduleExtraction
	if err := json.Unmarshal([]byte(text), &extraction); err != nil {
		return ScheduleExtraction{}, fmt.Errorf("aichat: could not parse schedule extraction: %s", text)
	}
	return extraction, nil
}
