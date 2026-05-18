package email

import (
	"bytes"
	"fmt"
	"net/smtp"
)

type SMTPClient struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func NewSMTPClient(host, port, username, password, from string) *SMTPClient {
	return &SMTPClient{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

func (c *SMTPClient) SendEmail(to []string, subject, body string) error {
	auth := smtp.PlainAuth("", c.Username, c.Password, c.Host)

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", c.From))
	if len(to) > 0 {
		msg.WriteString(fmt.Sprintf("To: %s\r\n", to[0]))
	}
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n")
	msg.WriteString(body)

	addr := fmt.Sprintf("%s:%s", c.Host, c.Port)
	return smtp.SendMail(addr, auth, c.From, to, msg.Bytes())
}
