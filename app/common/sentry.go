package common

import (
	"fmt"
	sentry "github.com/getsentry/sentry-go"
)

func InitSentry() {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              "https://d668bb2865b1434cbc8d82ac23d6c5a7@o4504879866707968.ingest.sentry.io/4505068337299456",
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
}
