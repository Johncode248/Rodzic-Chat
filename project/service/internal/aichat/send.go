package aichat

import (
	"context"
	"fmt"
	"time"

	aichatpg "github.com/Johncode248/Rodzic-Chat/project/common/aichat"
	parentsubscriptionspg "github.com/Johncode248/Rodzic-Chat/project/common/parentsubscriptions"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/nfz"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentsubscriptions"
)

const baseSystemPrompt = "Jesteś pomocnym asystentem AI w aplikacji Rodzic Chat, wspierającym rodziców w opiece nad dziećmi: przypomnienia, rekomendacje (lekarze, zajęcia, place zabaw), proste planowanie. Odpowiadaj po polsku, zwięźle i konkretnie. Nie udzielaj porad medycznych ani prawnych -- w takich sprawach zawsze odsyłaj do specjalisty.\n\n" +
	"Gdy rodzic szuka lekarza/specjalisty dla dziecka, użyj narzędzia find_nfz_benefit, a potem search_nfz_queues (oficjalne dane NFZ o kolejkach) -- dobierz najlepszą placówkę biorąc pod uwagę prognozowany czas oczekiwania, odległość/miejscowość i czy przyjmuje dzieci w odpowiednim wieku. Do zwykłych miejsc (przedszkola, place zabaw, zwykły pediatra/POZ) użyj search_places. Prezentuj wyniki czytelnie: nazwa placówki, adres/miejscowość, telefon, prognozowany czas oczekiwania -- w krótkiej liście, nie jako surowe dane."

// Send is the parent AI assistant's whole request pipeline: check the
// caller's plan has room this period, ask Claude, and (only on a
// successful reply -- a failed call never counts against the parent's
// quota) record the usage. See project/common/parentsubscriptions for the
// plan/limit model and project/service/internal/aichat/client.go for the
// actual Anthropic call.
func Send(ctx context.Context, accountId string, argument aichatpg.SendRequest) (res aichatpg.SendResponse, err error) {
	subscription, err := parentsubscriptions.GetOrCreate(ctx, accountId)
	if err != nil {
		return res, err
	}

	tokenLimit := parentsubscriptionspg.MonthlyAiTokenLimit(subscription.Plan)
	if tokenLimit >= 0 && subscription.AiTokensUsed >= tokenLimit {
		return res, parentsubscriptions.ErrAiLimitReached
	}

	systemPrompt := baseSystemPrompt
	if argument.ChildId != "" {
		if childContext, ok := childContextFor(ctx, accountId, argument.ChildId); ok {
			systemPrompt = fmt.Sprintf("%s\n\nRozmowa dotyczy konkretnego dziecka: %s.", baseSystemPrompt, childContext)
		}
	}
	if argument.Latitude != nil && argument.Longitude != nil {
		code, name := nfz.ProvinceForLocation(*argument.Latitude, *argument.Longitude)
		systemPrompt = fmt.Sprintf(
			"%s\n\nDomyślne województwo rodzica na podstawie jego lokalizacji: %s (kod NFZ %s). Użyj go w search_nfz_queues, chyba że rodzic sam wskaże inne miasto/województwo.",
			systemPrompt, name, code,
		)
	}

	reply, tokensUsed, err := converseWithTools(ctx, systemPrompt, argument.Messages)
	if err != nil {
		return res, err
	}

	updated, err := parentsubscriptions.IncrementUsageWithTokens(ctx, accountId, tokensUsed)
	if err != nil {
		// The reply was already generated -- still return it to the parent
		// even if the usage bookkeeping itself failed, rather than
		// discarding a real (costly) Anthropic response over a DB hiccup.
		return aichatpg.SendResponse{
			Reply:           reply,
			AiRequestsUsed:  subscription.AiRequestsUsed + 1,
			AiRequestsLimit: parentsubscriptionspg.MonthlyAiRequestLimit(subscription.Plan),
			AiTokensUsed:    subscription.AiTokensUsed + tokensUsed,
			AiTokensLimit:   tokenLimit,
		}, nil
	}

	return aichatpg.SendResponse{
		Reply:           reply,
		AiTokensUsed:    updated.AiTokensUsed,
		AiTokensLimit:   parentsubscriptionspg.MonthlyAiTokenLimit(updated.Plan),
		AiRequestsUsed:  updated.AiRequestsUsed,
		AiRequestsLimit: parentsubscriptionspg.MonthlyAiRequestLimit(updated.Plan),
	}, nil
}

// childContextFor returns a short "Name, N lat" description for childId if
// accountId is genuinely that child's parent (via children.SelectMine, the
// same unified personal+placowka list the "Moje dzieci" tab uses) -- ok is
// false if the child isn't the caller's, in which case Send falls back to
// the generic system prompt rather than erroring the whole request over an
// optional context hint.
func childContextFor(ctx context.Context, accountId, childId string) (description string, ok bool) {
	mine, err := children.SelectMine(ctx, accountId)
	if err != nil {
		return "", false
	}

	for _, child := range mine.Children {
		if child.Id != childId {
			continue
		}

		name := child.FirstName
		if child.BirthDate.IsZero() {
			return name, true
		}

		age := int(time.Since(child.BirthDate).Hours() / 24 / 365.25)
		return fmt.Sprintf("%s, %d lat", name, age), true
	}

	return "", false
}
