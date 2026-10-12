package push

import (
	"context"
	"fmt"

	"firebase.google.com/go/v4/messaging"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/devicetokens"
)

// Report is the outcome of SendToAccountsReport -- unlike SendToAccounts (fire-and-forget by
// design), the admin console needs to tell the operator what actually happened.
type Report struct {
	Enabled bool     `json:"enabled"`
	Devices int      `json:"devices"`
	Success int      `json:"success"`
	Failure int      `json:"failure"`
	Pruned  int      `json:"pruned"`
	Errors  []string `json:"errors,omitempty"`
}

// Enabled reports whether FCM credentials are configured.
func Enabled(ctx context.Context) bool { return getClient(ctx) != nil }

// SendToAccountsReport is SendToAccounts with counts, used by the local admin panel.
func SendToAccountsReport(ctx context.Context, accountIds []string, title, body string, data map[string]string) Report {
	rep := Report{}
	c := getClient(ctx)
	if c == nil {
		return rep
	}
	rep.Enabled = true
	tokens, err := devicetokens.ForAccounts(ctx, accountIds)
	if err != nil {
		rep.Errors = append(rep.Errors, err.Error())
		return rep
	}
	rep.Devices = len(tokens)
	for start := 0; start < len(tokens); start += maxTokensPerBatch {
		end := min(start+maxTokensPerBatch, len(tokens))
		batch := tokens[start:end]
		strs := make([]string, len(batch))
		for i, t := range batch {
			strs[i] = t.Token
		}
		resp, err := c.SendEachForMulticast(ctx, &messaging.MulticastMessage{
			Tokens:       strs,
			Notification: &messaging.Notification{Title: title, Body: body},
			Data:         data,
			Android:      &messaging.AndroidConfig{Priority: "high"},
			APNS:         &messaging.APNSConfig{Payload: &messaging.APNSPayload{Aps: &messaging.Aps{Sound: "default"}}},
		})
		if err != nil {
			rep.Errors = append(rep.Errors, err.Error())
			rep.Failure += len(batch)
			continue
		}
		rep.Success += resp.SuccessCount
		rep.Failure += resp.FailureCount
		var dead []string
		for i, s := range resp.Responses {
			if s.Success || s.Error == nil {
				continue
			}
			if messaging.IsUnregistered(s.Error) || messaging.IsInvalidArgument(s.Error) {
				dead = append(dead, batch[i].Id)
			} else if len(rep.Errors) < 5 {
				rep.Errors = append(rep.Errors, fmt.Sprint(s.Error))
			}
		}
		if len(dead) > 0 {
			if devicetokens.DeleteTokens(ctx, dead) == nil {
				rep.Pruned += len(dead)
			}
		}
	}
	return rep
}
