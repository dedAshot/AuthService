package notificator

import (
	"log/slog"
	"net/smtp"
)

var (
	smtpClient struct {
		SMTPConfig
		auth smtp.Auth
	}
)

type SMTPConfig struct {
	email, password, host, port string
}

func NewSMTPConfig(email, password, host, port string) *SMTPConfig {
	return &SMTPConfig{
		email:    email,
		password: password,
		host:     host,
		port:     port,
	}
}

func ConnectToSMTP(cfg *SMTPConfig) {
	smtpClient.email, smtpClient.password, smtpClient.host, smtpClient.port = cfg.email, cfg.password, cfg.host, cfg.port
	smtpClient.auth = smtp.PlainAuth("Go robot", cfg.email, cfg.password, cfg.host)
}

func SendMsg(reciverEmail, msg, subject string) error {
	slog.Info("send message to", "email", reciverEmail)
	msg = "To: " + reciverEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		msg + "\r\n"

	err := smtp.SendMail(
		smtpClient.host+":"+smtpClient.port,
		smtpClient.auth,
		smtpClient.email,
		append(make([]string, 0), reciverEmail),
		[]byte(msg),
	)

	return err
}

//golang_testovoe
//googlepassword cykrpzrjknyqwcrf
