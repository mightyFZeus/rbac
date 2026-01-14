package main

import (
	"fmt"
	"net/smtp"

	"github.com/mightyfzeus/rbac/internal/env"
	"go.uber.org/zap"
)

func (app *application) SendMail(recipient, body, subject string) error {

	sender := env.GetString("EMAIL", "")
	appPassword := env.GetString("APP_PASSWORD", "")

	from := sender
	password := appPassword
	to := []string{recipient}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	msg := []byte(subject + "\n" + body)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) sendAdminInviteAsync(email, token string) {
	go func() {

		body := fmt.Sprintf(
			"This is your admin activation token. It expires in 24hrs.\n\n%s",
			token,
		)

		if err := app.SendMail(email, body, "Admin Invitation"); err != nil {

			app.logger.Error(
				"failed to send admin invite",
				zap.String("email", email),
				zap.Error(err),
			)

		}

	}()

}
