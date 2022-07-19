package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/michaelpeterswa/good-morning-text/internal/handlers"
	"github.com/michaelpeterswa/good-morning-text/internal/logging"
	"github.com/michaelpeterswa/good-morning-text/internal/message"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func main() {
	logger, err := logging.InitZap()
	if err != nil {
		log.Panicf("could not acquire zap logger: %s", err.Error())
	}
	logger.Info("good-morning-text init...")

	timezone := "America/Los_Angeles"
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		logger.Fatal("could not load location", zap.String("location", timezone), zap.Error(err))
	}

	phoneNumbers, err := message.InitTwilioPhoneNumbers()
	if err != nil {
		logger.Fatal("could not init twilio phone numbers", zap.Error(err))
	}

	tC := message.InitTwilio()

	twilioClient := message.NewTwilioClient(tC, phoneNumbers, logger)

	cronString := os.Getenv("CRON_STRING")
	if cronString == "" {
		logger.Fatal("cron string not set")
	}

	c := cron.New(cron.WithLocation(loc))
	_, err = c.AddFunc(cronString, twilioClient.SendMessage())
	if err != nil {
		logger.Fatal("could not add cron job", zap.Error(err))
	}
	c.Start()

	r := mux.NewRouter()
	r.HandleFunc("/healthcheck", handlers.HealthcheckHandler)
	r.Handle("/metrics", promhttp.Handler())
	http.Handle("/", r)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Fatal("could not start http server", zap.Error(err))
	}
}
