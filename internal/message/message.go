package message

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"go.uber.org/zap"
)

var (
	ErrTwilioFromPhoneNumberNotSet = fmt.Errorf("twilio from phone number not set")
	ErrTwilioToPhoneNumberNotSet   = fmt.Errorf("twilio to phone number not set")
)

type TwilioClient struct {
	Client      *twilio.RestClient
	PhoneNumber *TwilioPhoneNumbers
	Logger      *zap.Logger
}

type TwilioPhoneNumbers struct {
	From string   `json:"from"`
	To   []string `json:"to"`
}

func NewTwilioClient(client *twilio.RestClient, tpn *TwilioPhoneNumbers, logger *zap.Logger) *TwilioClient {
	return &TwilioClient{
		Client:      client,
		PhoneNumber: tpn,
		Logger:      logger,
	}
}

func InitTwilio() *twilio.RestClient {
	return twilio.NewRestClient()
}

func InitTwilioPhoneNumbers() (*TwilioPhoneNumbers, error) {
	from := os.Getenv("TWILIO_FROM_NUMBER")
	if from == "" {
		return nil, ErrTwilioFromPhoneNumberNotSet
	}
	to := strings.Split(os.Getenv("TWILIO_TO_NUMBER"), ",")
	if len(to) == 0 {
		return nil, ErrTwilioToPhoneNumberNotSet
	}

	return &TwilioPhoneNumbers{
		From: from,
		To:   to,
	}, nil
}

// SendMessage sends a message via cron.
func (tc *TwilioClient) SendMessage() func() {
	return func() {
		for _, to := range tc.PhoneNumber.To {
			params := &openapi.CreateMessageParams{}
			params.SetTo(to)
			params.SetFrom(tc.PhoneNumber.From)
			params.SetBody("good morning, red bird!")

			resp, err := tc.Client.Api.CreateMessage(params)
			if err != nil {
				tc.Logger.Error("could not send message", zap.Error(err), zap.String("to", to), zap.String("from", tc.PhoneNumber.From))
			} else {
				response, _ := json.Marshal(*resp)
				tc.Logger.Info("message sent", zap.String("to", to), zap.String("from", tc.PhoneNumber.From), zap.String("response", string(response)))
			}
		}
	}
}
