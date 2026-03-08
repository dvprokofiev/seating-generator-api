package mailer

import (
	"fmt"
	"net/smtp"

	"github.com/google/uuid"
)

type SMTPMailer struct {
	host     string
	port     string
	username string
	password string
	from     string
	baseURL  string
}

func NewSMTPMailer(host, port, user, pass, from, baseURL string) *SMTPMailer {
	return &SMTPMailer{
		host:     host,
		port:     port,
		username: user,
		password: pass,
		from:     from,
		baseURL:  baseURL,
	}
}

func (m *SMTPMailer) SendVerificationEmail(to string, token uuid.UUID) error {
	link := fmt.Sprintf("%s/verify?token=%s", m.baseURL, token.String())
	subject := "Subject: Подтверждение регистрации\r\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<html>
			<body>
				<h2>Добро пожаловать в генератор рассадок!</h2>
				<p>Пожалуйста, подтвердите ваш адрес электронной почты, перейдя по ссылке:</p>
				<p><a href="%s">Подтвердить почту</a></p>
				<br>
				<p>Ссылка действительна в течение 24 часов.</p>
			</body>
		</html>
	`, link)

	msg := []byte(subject + mime + body)
	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	auth := smtp.PlainAuth("", m.username, m.password, m.host)

	return smtp.SendMail(addr, auth, m.from, []string{to}, msg)
}
