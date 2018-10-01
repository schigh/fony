package main

import (
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"
)

// FonyPayloadIndexHeader is the index of the payload we wish to see for a given endpoint
const FonyPayloadIndexHeader = "X-Fony-Index"

var (
	suiteFile   string
	port        string
	sequenceMap *endpointSequenceMap
	logger      *zap.Logger
)

func init() {
	flag.StringVar(&suiteFile, "f", "./fony.json", "Absolute path to the fony suite file")
	flag.StringVar(&port, "p", "80", "http port (for local testing)")
}

func main() {
	flag.Parse()
	logger, _ = zap.NewProduction()
	defer logger.Sync()
	logger.Info("starting up now")

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
