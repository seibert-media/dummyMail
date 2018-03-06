package mail

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/icrowley/fake"
	sendgrid "github.com/sendgrid/sendgrid-go"
	mail "github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/wawandco/fako"
)

// Mail with dummy data
type Mail struct {
	Sender         string `fako:"full_name"`
	SenderEmail    string `fako:"custom_email"`
	Recipient      string `fako:"full_name"`
	RecipientEmail string `fako:"custom_recipient"`
	Subject        string `fako:"words"`
	Message        string `fako:"paragraphs"`
}

// Sender for sendgrid
type Sender struct {
	*sendgrid.Client
}

// Init the custom fill function
func Init(apiKey, senderSuffix string, recipients []string) *Sender {
	fako.Register("custom_email", func() string {
		m := fake.EmailAddress()
		mailArr := strings.Split(m, "@")
		return fmt.Sprintf("%s@%s", mailArr[0], senderSuffix)
	})
	rand.Seed(time.Now().Unix())
	fako.Register("custom_recipient", func() string {
		recipient := recipients[rand.Intn(len(recipients))]
		return recipient
	})
	s := new(Sender)
	s.Client = sendgrid.NewSendClient(apiKey)
	return s
}

// Generate new Mail with random data
func Generate() *Mail {
	m := &Mail{}
	fako.Fill(m)
	return m
}

// Send the mail via Sendgrid
func (s *Sender) Send(m *Mail) {
	from := mail.NewEmail(m.Sender, m.SenderEmail)
	to := mail.NewEmail(m.Recipient, m.RecipientEmail)
	message := mail.NewSingleEmail(from, m.Subject, to, m.Message, m.Message)
	response, err := s.Client.Send(message)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}
