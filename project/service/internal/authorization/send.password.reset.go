package authorization

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	userspg "github.com/Johncode248/Rodzic-Chat/project/common/users"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/users"
	"github.com/Johncode248/Rodzic-Chat/variables"
)

func SendPasswordResetEmail(ctx context.Context, toEmail string) (err error) {
	ctx = logger.Enter(ctx, "authorization.SendPasswordResetEmail", toEmail)
	defer logger.Exit(ctx, "authorization.SendPasswordResetEmail", &err)

	host, port, login, password, from := "smtp-relay.brevo.com", "587", "b09c81001@smtp-brevo.com", "xsmtpsib-ff275231bad30737cd02924872278bab84be0b51d42616850de032457e899e52-yLTKD1FUsS27PERo", "noreply@dom2sky.pl"

	userlist, err := users.Select(ctx, userspg.SelectArgument{Query: databasepg.Query{
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

	if len(userlist.Users) == 0 {
		return errors.New("no users found")
	}

	accessToken, err := credentials.GenerateAccessToken(ctx, userlist.Users[0].Id, toEmail)
	if err != nil {
		return err
	}

	resetLink := fmt.Sprintf("%s/password-reset?access_token=%s", variables.DOMEIN_F, accessToken)

	subject := "Resetowanie hasła 🔒"
	body := fmt.Sprintf(`<div style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="padding: 30px; background-color: #f9f9f9; border-radius: 10px;">
        <h2 style="color: #1f2937; text-align: center; margin-bottom: 25px;">Resetowanie hasła</h2>

        <p style="margin-bottom: 20px;">Otrzymaliśmy prośbę o zresetowanie hasła do Twojego konta w serwisie Dom2sky.pl. Jeśli to nie Ty wysłałeś/aś tę prośbę, zignoruj tę wiadomość — Twoje hasło pozostanie bez zmian.</p>

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
            <strong>Zespół Dom2sky.pl</strong>
        </p>

        <div style="text-align: center; margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; color: #7f8c8d; font-size: 14px;">
            <p style="margin: 5px 0;">© 2026 Dom2sky.pl. Wszystkie prawa zastrzeżone.</p>
            <p style="margin: 5px 0;">Bezpieczeństwo Twojego konta jest dla nas priorytetem. Ten link wygaśnie za 1 godzinę.</p>
            <p style="margin: 5px 0; font-size: 12px;">Jeśli nie rozpoznajesz tej aktywności, <a href="mailto:stanislaw@dom2sky.pl" style="color: #0ea5e9;">skontaktuj się z nami</a>.</p>
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
