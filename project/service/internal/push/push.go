// Package push sends mobile push notifications via Firebase Cloud
// Messaging, which fans out to both Android (direct) and iOS (via APNs
// under the hood) from one API, so the backend only needs one integration.
package push

import (
	"context"
	"log"
	"os"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/Johncode248/Rodzic-Chat/logger"
	devicetokenspg "github.com/Johncode248/Rodzic-Chat/project/common/devicetokens"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/devicetokens"
	"google.golang.org/api/option"
)

var (
	initOnce sync.Once
	client   *messaging.Client
)

// FCMCredentialsFile is the path to the Firebase service-account JSON
// (Firebase Console -> Project Settings -> Service Accounts -> Generate
// new private key). Deliberately not required at boot (unlike Stripe's
// keys in variables/secrets.go): push notifications are additive to a
// running app, so a missing/misconfigured credentials file logs a warning
// and every send becomes a no-op instead of crash-looping the whole
// backend over a feature that isn't chat itself.
func getClient(ctx context.Context) *messaging.Client {
	initOnce.Do(func() {
		path := os.Getenv("FCM_CREDENTIALS_FILE")
		if path == "" {
			log.Println("push: FCM_CREDENTIALS_FILE not set, push notifications are disabled")
			return
		}

		app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(path))
		if err != nil {
			log.Printf("push: failed to init Firebase app: %v", err)
			return
		}

		c, err := app.Messaging(ctx)
		if err != nil {
			log.Printf("push: failed to init FCM client: %v", err)
			return
		}

		client = c
	})
	return client
}

// maxTokensPerBatch is FCM's own cap on a single SendEachForMulticast call.
const maxTokensPerBatch = 500

// SendToAccounts pushes the same notification to every device registered
// to any of accountIds, skipping silently (not erroring) if push isn't
// configured or nobody has a registered device -- a failed/absent push
// should never block or fail the action that triggered it (e.g. sending a
// chat message). Any token FCM reports as no-longer-valid is pruned from
// the device_tokens table so future sends don't keep retrying it.
func SendToAccounts(ctx context.Context, accountIds []string, title, body string, data map[string]string) {
	ctx = logger.Enter(ctx, "push.SendToAccounts", accountIds, title)
	var err error
	defer logger.Exit(ctx, "push.SendToAccounts", &err)

	c := getClient(ctx)
	if c == nil {
		return
	}

	tokens, selectErr := devicetokens.ForAccounts(ctx, accountIds)
	if selectErr != nil {
		err = selectErr
		log.Printf("push: failed to load device tokens: %v", err)
		return
	}
	if len(tokens) == 0 {
		return
	}

	for start := 0; start < len(tokens); start += maxTokensPerBatch {
		end := start + maxTokensPerBatch
		if end > len(tokens) {
			end = len(tokens)
		}
		sendBatch(ctx, c, tokens[start:end], title, body, data)
	}
}

func sendBatch(ctx context.Context, c *messaging.Client, batch []devicetokenspg.DeviceToken, title, body string, data map[string]string) {
	tokenStrings := make([]string, len(batch))
	for i, t := range batch {
		tokenStrings[i] = t.Token
	}

	response, err := c.SendEachForMulticast(ctx, &messaging.MulticastMessage{
		Tokens: tokenStrings,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	})
	if err != nil {
		log.Printf("push: batch send failed: %v", err)
		return
	}

	if response.FailureCount == 0 {
		return
	}

	var deadIds []string
	for i, single := range response.Responses {
		if single.Success || single.Error == nil {
			continue
		}
		if messaging.IsUnregistered(single.Error) || messaging.IsInvalidArgument(single.Error) {
			deadIds = append(deadIds, batch[i].Id)
		}
	}
	if len(deadIds) > 0 {
		if delErr := devicetokens.DeleteTokens(ctx, deadIds); delErr != nil {
			log.Printf("push: failed to prune %d dead device token(s): %v", len(deadIds), delErr)
		}
	}
}
