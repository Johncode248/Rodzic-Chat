package authorization

import (
	"context"
	"net/smtp"

	"github.com/Johncode248/Rodzic-Chat/logger"
	"github.com/Johncode248/Rodzic-Chat/variables"
)

func SendRegisterEmail(ctx context.Context, toEmail string) (err error) {
	ctx = logger.Enter(ctx, "authorization.SendRegisterEmail", toEmail)
	defer logger.Exit(ctx, "authorization.SendRegisterEmail", &err)

	host, port, login, password, from := variables.SMTPHost, variables.SMTPPort, variables.SMTPLogin, variables.SMTPPassword, variables.SMTPFrom

	subject := "Witamy w Rodzic Chat"
	body := `<div style="padding: 30px; background-color: #f9f9f9; border-radius: 0 0 10px 10px;">
        <h2 style="color: #2c3e50; text-align: center; margin-bottom: 25px;">Cześć 👋</h2>
        
        <p style="margin-bottom: 20px;">Serdecznie witamy w naszej społeczności! Dziękujemy, że do nas dołączyłeś/aś.</p>
        
        <p style="margin-bottom: 30px;">Twoje konto zostało pomyślnie utworzone i jest już gotowe do użycia.</p>
        
        <div style="background: white; padding: 15px; margin: 15px 0; border-radius: 8px; border-left: 4px solid #2575fc;">
            <span style="font-size: 20px;">✨</span> <strong>Od czego zacząć?</strong>
            <p style="margin: 10px 0 0 0;">Zapoznaj się z naszymi głównymi funkcjami i odkryj, jak możemy ułatwic twoją pracę.</p>
        </div>

        <p style="margin-bottom: 20px;">Jeśli masz jakiekolwiek pytania lub potrzebujesz pomocy, nasz zespół wsparcia jest zawsze gotowy, aby Ci pomóc.</p>
        
        <p>
            Pozdrawiamy serdecznie,<br>
            <strong>Zespół Rodzic Chat</strong>
        </p>

        <div style="text-align: center; margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; color: #7f8c8d; font-size: 14px;">
            <p style="margin: 5px 0;">© 2026 Flashy Code Stanisław Marcisz | Rodzic Chat</p>
            <p style="margin: 5px 0;">Otrzymałeś tę wiadomość, ponieważ zarejestrowałeś się w naszej aplikacji.</p>
        </div>
    </div>
</div>`

	msg := "MIME-Version: 1.0;\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\n" +
		"From: " + from + "\n" +
		"To: " + toEmail + "\n" +
		"Subject: " + subject + "\n\n" + body

	if err = smtp.SendMail(host+":"+port, smtp.PlainAuth("", login, password, host), from, []string{toEmail}, []byte(msg)); err != nil {
		return err
	}

	return nil
}
