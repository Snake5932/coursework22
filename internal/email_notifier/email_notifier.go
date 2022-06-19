package email_notifier

import (
	"crypto/tls"
	"net"
	"net/smtp"
)

type Configuration struct {
	From       string
	Password   string
	ServerHost string
	ServerPort string
}

type Notifier struct {
	Config Configuration
}

func (notifier *Notifier) SendMail(to, subj, msg string) error {
	message := "From: " + notifier.Config.From + "\r\n"
	message += "To: " + to + "\r\n"
	message += "Subject: " + subj + "\r\n"
	message += "\r\n" + msg

	tlsconfig := &tls.Config{
		ServerName: notifier.Config.ServerHost,
	}
	auth := smtp.PlainAuth("", notifier.Config.From, notifier.Config.Password, notifier.Config.ServerHost)

	conn, err := net.Dial("tcp", notifier.Config.ServerHost+":"+notifier.Config.ServerPort)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, notifier.Config.ServerHost)
	if err != nil {
		return err
	}

	if err = c.StartTLS(tlsconfig); err != nil {
		return err
	}

	if err = c.Auth(auth); err != nil {
		return err
	}

	if err = c.Mail(notifier.Config.From); err != nil {
		return err
	}

	if err = c.Rcpt(to); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	w.Close()
	c.Quit()
	return nil
}
