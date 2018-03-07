package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/playnet-public/libs/log"
	"github.com/seibert-media/dummyMail/pkg/mail"

	raven "github.com/getsentry/raven-go"
	"github.com/golang/glog"
	"github.com/kolide/kit/version"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	app    = "dummyMail"
	appKey = "dummyMail"
)

var (
	maxprocsPtr = flag.Int("maxprocs", runtime.NumCPU(), "max go procs")
	sentryDsn   = flag.String("sentrydsn", "https://0484bfd72e85493296b64d4010a9c645:9d7b35795e36450d8b876f45b173aa07@sentry.io/299868", "sentry dsn key")
	dbgPtr      = flag.Bool("debug", false, "debug printing")
	versionPtr  = flag.Bool("version", true, "show or hide version info")

	apiKey       = flag.String("apiKey", "", "sendgrid api key")
	senderSuffix = flag.String("senderSuffix", "", "the domain from which e-mails are sent")
	countPtr     = flag.Int("count", 0, "count of e-mails generated")

	recipients mail.Recipients

	sentry *raven.Client
)

func main() {
	flag.Var(&recipients, "recipient", "recipient to send e-mail to")
	flag.Parse()

	if *versionPtr {
		fmt.Printf("-- //S/M %s --\n", app)
		version.PrintFull()
	}
	runtime.GOMAXPROCS(*maxprocsPtr)

	// prepare glog
	defer glog.Flush()
	glog.CopyStandardLogTo("info")

	var zapFields []zapcore.Field
	// hide app and version information when debugging
	if !*dbgPtr {
		zapFields = []zapcore.Field{
			zap.String("app", appKey),
			zap.String("version", version.Version().Version),
		}
	}

	// prepare zap logging
	log := log.New(appKey, *sentryDsn, *dbgPtr).WithFields(zapFields...)
	defer log.Sync()
	log.Info("preparing")

	var err error

	// prepare sentry error logging
	sentry, err = raven.New(*sentryDsn)
	if err != nil {
		panic(err)
	}
	err = raven.SetDSN(*sentryDsn)
	if err != nil {
		panic(err)
	}
	errs := make(chan error)

	// run main code
	log.Info("starting")
	sentryErr, _ := raven.CapturePanicAndWait(func() {
		if err := do(log); err != nil {
			log.Fatal("fatal error encountered", zap.Error(err))
			raven.CaptureErrorAndWait(err, map[string]string{"isFinal": "true"})
			errs <- err
		}
	}, nil)
	if sentryErr != nil {
		err := sentryErr.(error)
		log.Error("panic", zap.Error(err))
	}
	log.Info("finished")
}

func do(log *log.Logger) error {
	sender := mail.Init(log, *apiKey, *senderSuffix, recipients.Array())
	for i := 0; i < *countPtr; i++ {
		mail := mail.Generate()
		log.Info("sending mail", zap.String("recipient", mail.RecipientEmail), zap.String("subject", mail.Subject))
		sender.Send(mail)
	}
	return nil
}
