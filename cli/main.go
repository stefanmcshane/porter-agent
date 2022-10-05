package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joeshaw/envdecode"
	"github.com/porter-dev/porter-agent/internal/envconf"
	"github.com/porter-dev/porter-agent/internal/logger"
	"github.com/porter-dev/porter-agent/pkg/logstore"
	"github.com/porter-dev/porter-agent/pkg/logstore/lokistore"
	"github.com/porter-dev/porter-agent/pkg/logstore/memorystore"
	flag "github.com/spf13/pflag"
)

type logWriter struct{}

func (lw *logWriter) Write(timestamp *time.Time, log string) error {
	fmt.Printf("%v: %s\n", timestamp, log)
	return nil
}

func main() {
	envDecoderConf := &envconf.EnvDecoderConf{}

	if err := envdecode.StrictDecode(envDecoderConf); err != nil {
		logger.NewErrorConsole(true).Fatal().Caller().Msgf("could not decode env conf: %v", err)
		os.Exit(1)
	}

	l := logger.NewConsole(envDecoderConf.Debug)

	var labels []string
	var start string
	var limit uint32

	flag.StringArrayVarP(&labels, "label", "l", []string{}, "labels to use for logstore")
	flag.StringVar(&start, "start", "", "start time to use for logstore")
	flag.Uint32Var(&limit, "limit", 0, "limit to use for logstore")

	flag.Parse()

	if len(labels) == 0 {
		l.Fatal().Caller().Msg("at least one label must be provided")
	}

	startTime, err := time.Parse(time.RFC3339, start)

	if start == "" || err != nil {
		l.Fatal().Caller().Msg("valid RFC 3339 time must be provided")
	}

	var logStore logstore.LogStore
	var logStoreKind string

	if envDecoderConf.LogStoreConf.LogStoreKind == "memory" {
		logStoreKind = "memory"
		logStore, err = memorystore.New("test", memorystore.Options{})
	} else {
		logStoreKind = "loki"
		logStore, err = lokistore.New("test", lokistore.LogStoreConfig{Address: envDecoderConf.LogStoreConf.LogStoreAddress})
	}

	if err != nil {
		l.Fatal().Caller().Msgf("%s-based log store setup failed: %v", logStoreKind, err)
	}

	stopChan := make(chan struct{}, 1)
	w := &logWriter{}

	labelsMap := make(map[string]string)

	for _, label := range labels {
		if key, val, found := strings.Cut(label, "="); !found || key == "" || val == "" {
			l.Fatal().Caller().Msgf("invalid label provided: %s", l)
		} else {
			labelsMap[key] = val
		}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig
		stopChan <- struct{}{}
	}()

	fmt.Println("TAILING WITH", labelsMap, startTime, limit)

	if err := logStore.Tail(logstore.TailOptions{
		Labels:               labelsMap,
		Start:                startTime,
		Limit:                limit,
		CustomSelectorSuffix: "event_store!=\"true\"",
	}, w, stopChan); err != nil {
		l.Fatal().Caller().Msgf("could not tail logs: %v", err)
	}
}
