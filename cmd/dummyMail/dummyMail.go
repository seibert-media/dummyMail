package main

import (
	"fmt"
	"runtime"

	"github.com/playnet-public/libs/log"

	flag "github.com/bborbe/flagenv"

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
	sentryDsn   = flag.String("sentrydsn", "", "sentry dsn key")
	dbgPtr      = flag.Bool("debug", false, "debug printing")
	versionPtr  = flag.Bool("version", true, "show or hide version info")

	sentry *raven.Client
)

func main() {
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
	return nil
}