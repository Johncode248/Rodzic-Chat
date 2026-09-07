package authorization

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/library/random"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	passwordresetspg "github.com/Johncode248/Rodzic-Chat/project/common/passwordresets"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/passwordresets"
	"github.com/Johncode248/Rodzic-Chat/variables"
)

// passwordResetTokenTTL matches the "ważny przez 1 godzinę" promised in the
// email body below -- keep both in sync if either changes.
const passwordResetTokenTTL = 1 * time.Hour

func SendPasswordResetEmail(ctx context.Context, toEmail string) (err error) {
	ctx = logger.Enter(ctx, "authorization.SendPasswordResetEmail", toEmail)
	defer logger.Exit(ctx, "authorization.SendPasswordResetEmail", &err)

	host, port, login, password, from := variables.SMTPHost, variables.SMTPPort, variables.SMTPLogin, variables.SMTPPassword, variables.SMTPFrom

	accountlist, err := accounts.Select(ctx, accountspg.SelectArgument{Query: databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{
						Key:      "email",
						Operator: "=",
						Value:    toEmail,
					},
				},
			},
		},
	}})
	if err != nil {
		return err
	}

	if len(accountlist.Accounts) == 0 {
		return errors.New("no accounts found")
	}

	rawToken := random.GenerateToken()

	if _, err = passwordresets.Create(ctx, passwordresetspg.CreateArgument{
		AccountId: accountlist.Accounts[0].Id,
		TokenHash: credentials.HashToken(rawToken),
		ExpiresAt: time.Now().Add(passwordResetTokenTTL),
	}); err != nil {
		return err
	}

	resetLink := fmt.Sprintf("%s/password-reset?token=%s", variables.DOMEIN_F, rawToken)

	subject := "Resetowanie hasła 🔒"
	body := fmt.Sprintf(`<div style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="padding: 30px; background-color: #f9f9f9; border-radius: 10px;">
        <h2 style="color: #1f2937; text-align: center; margin-bottom: 25px;">Resetowanie hasła</h2>

        <p style="margin-bottom: 20px;">Otrzymaliśmy prośbę o zresetowanie hasła do Twojego konta w serwisie Rodzic Chat. Jeśli to nie Ty wysłałeś/aś tę prośbę, zignoruj tę wiadomość — Twoje hasło pozostanie bez zmian.</p>

        <p style="margin-bottom: 30px;">Aby ustawić nowe hasło, kliknij w poniższy przycisk (link jest ważny przez 1 godzinę):</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="display: inline-block; padding: 14px 28px; background-color: #0ea5e9; color: #ffffff; text-align: center; text-decoration: none; border-radius: 50px; font-weight: bold; font-size: 16px;">
                Zresetuj hasło 🔑
            </a>
        </div>

        <p style="margin-bottom: 20px; font-size: 13px; color: #6b7280;">
            Jeśli przycisk nie działa, skopiuj i wklej poniższy link do przeglądarki:<br>
            <a href="%s" style="color: #0ea5e9; word-break: break-all;">%s</a>
        </p>

        <p style="margin-bottom: 20px;">Jeśli masz problemy z resetem hasła, skontaktuj się z naszym zespołem wsparcia.</p>

        <p>
            Pozdrawiamy,<br>
            <strong>Zespół Rodzic Chat</strong>
        </p>

        <div style="text-align: center; margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; color: #7f8c8d; font-size: 14px;">
            <p style="margin: 5px 0;">© 2026 Rodzic Chat. Wszystkie prawa zastrzeżone.</p>
            <p style="margin: 5px 0;">Bezpieczeństwo Twojego konta jest dla nas priorytetem. Ten link wygaśnie za 1 godzinę.</p>
            <p style="margin: 5px 0; font-size: 12px;">Jeśli nie rozpoznajesz tej aktywności, <a href="mailto:kontakt@rodzicchat.pl" style="color: #0ea5e9;">skontaktuj się z nami</a>.</p>
        </div>
    </div>
</div>`, resetLink, resetLink, resetLink)

	msg := "MIME-Version: 1.0;\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\n" +
		"From: " + from + "\n" +
		"To: " + toEmail + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	if err = smtp.SendMail(host+":"+port, smtp.PlainAuth("", login, password, host), from, []string{toEmail}, []byte(msg)); err != nil {
		return err
	}

	return nil
}
