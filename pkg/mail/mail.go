package mail

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/playnet-public/libs/log"
	"go.uber.org/zap"

	"github.com/icrowley/fake"
	sendgrid "github.com/sendgrid/sendgrid-go"
	mail "github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/wawandco/fako"
)

// Mail with dummy data
type Mail struct {
	Sender         string `fako:"full_name"`
	SenderEmail    string `fako:"custom_email"`
	RecipientEmail string `fako:"custom_recipient"`
	Subject        string `fako:"words"`
	Message        string `fako:"paragraphs"`
}

// Sender for sendgrid
type Sender struct {
	*sendgrid.Client
	log *log.Logger
}

// Recipient .
type Recipient string

func (r Recipient) String() string { return string(r) }

// Recipients array
type Recipients []Recipient

func (r *Recipients) String() string {
	var str []string
	for _, dir := range *r {
		str = append(str, dir.String())
	}
	return strings.Join(str, ",")
}

// Array of type []string from Recipients
func (r *Recipients) Array() []string {
	var str []string
	for _, dir := range *r {
		str = append(str, dir.String())
	}
	return str
}

// Set new Recipient to Recipients
func (r *Recipients) Set(value string) error {
	*r = append(*r, Recipient(value))
	return nil
}

// Init the custom fill function
func Init(log *log.Logger, apiKey, senderSuffix string, recipients []string) *Sender {
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
	s.log = log
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
	recipient := strings.Split(m.RecipientEmail, "@")[0]
	to := mail.NewEmail(recipient, m.RecipientEmail)
	message := mail.NewSingleEmail(from, m.Subject, to, m.Message, m.Message)
	response, err := s.Client.Send(message)
	if err != nil {
		s.log.Error("send error", zap.String("recipient", m.RecipientEmail), zap.String("subject", m.Subject), zap.Error(err))
	} else {
		s.log.Info("mail sent", zap.Int("statusCode", response.StatusCode), zap.String("body", response.Body), zap.String("recipient", m.RecipientEmail), zap.String("subject", m.Subject), zap.Error(err))
	}
}
