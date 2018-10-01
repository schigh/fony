package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// FonyPayloadIndexHeader is the index of the payload we wish to see for a given endpoint
const FonyPayloadIndexHeader = "X-Fony-Index"

var (
	suiteFile   string
	port        string
	sequenceMap *endpointSequenceMap
	logger      *zerolog.Logger
)

func init() {
	// logger fields
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.MessageFieldName = "msg"
	zerolog.ErrorFieldName = "error_message"
	zerolog.TimestampFieldName = "timestamp"

	flag.StringVar(&suiteFile, "f", "./fony.json", "Absolute path to the fony suite file")
	flag.StringVar(&port, "p", "80", "http port (for local testing)")
}

func main() {
	flag.Parse()
	log.Info().Msg("starting up now")

	suite, ok := setupSuite()
	if !ok {
		os.Exit(1)
	}

	sequenceMap = &endpointSequenceMap{
		sequences: make(map[string]*endpointSequence, len(suite.Endpoints)),
	}

	e, ok := registerSuite(suite)
	if !ok {
		os.Exit(1)
	}

	e.Start(fmt.Sprintf("0.0.0.0:%s", port))
}
