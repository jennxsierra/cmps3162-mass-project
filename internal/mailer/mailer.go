package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed templates
var templateFS embed.FS	// embed the files from the templates directory into the binary

// dialer is responsible for connecting to the SMTP server and sending emails
// sender is the email address that will appear in the "From" field of the email
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

// New creates a new Mailer instance with the provided
// SMTP server configuration and sender email address
func New(host string, port int, username, password, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Send sends an email to the specified recipient using the provided template file and data
func (m Mailer) Send(recipient, templateFile string, data any) error {
	templates, err := template.ParseFS(templateFS, fmt.Sprintf("templates/%s", templateFile))
	if err != nil {
		return err
	}

	// fill in the subject
	var subjectBuffer bytes.Buffer
	err = templates.ExecuteTemplate(&subjectBuffer, "subject", data)
	if err != nil {
		return err
	}

	// fill in the plain text body
	var plainBodyBuffer bytes.Buffer
	err = templates.ExecuteTemplate(&plainBodyBuffer, "plainBody", data)
	if err != nil {
		return err
	}

	// fill in the HTML body``
	var htmlBodyBuffer bytes.Buffer
	err = templates.ExecuteTemplate(&htmlBodyBuffer, "htmlBody", data)
	if err != nil {
		return err
	}

	// craft the email message and send it using the dialer
	// try up to 3 times, with a short delay between attempts
	var sendErr error
	for attempt := range 3 {
		message := mail.NewMessage()
		message.SetHeader("From", m.sender)
		message.SetHeader("To", recipient)
		message.SetHeader("Subject", subjectBuffer.String())
		message.SetBody("text/plain", plainBodyBuffer.String())
		message.AddAlternative("text/html", htmlBodyBuffer.String())

		sendErr = m.dialer.DialAndSend(message)
		if sendErr == nil {
			return nil
		}

		if attempt < 2 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	return sendErr
}
