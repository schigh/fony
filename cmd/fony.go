package main

import (
	"flag"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	JSONType = "application/json; charset=utf8"
)

var (
	version            string
	build              string
	suiteFile          string
	port               int
	defaultContentType string
	sequenceMap        *endpointSequenceMap
	jsonRX             *regexp.Regexp
)

func init() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.MessageFieldName = "msg"
	zerolog.ErrorFieldName = "err"
	zerolog.TimestampFieldName = "ts"

	flag.StringVar(&suiteFile, "f", "fony.json", "path to fony suite file")
	flag.IntVar(&port, "p", 80, "http port")

	defaultContentType = JSONType

	jsonRX = regexp.MustCompile(`(?i)^application/j(son|avascript)`)
}

func main() {
	flag.Parse()
	if version == "" {
		version = "local"
		build = "local"
	}
	log.Info().Msgf("fony version: %s | build: %s", version, build)

	sequenceMap = &endpointSequenceMap{
		sequences: make(map[string]*endpointSequence),
	}

	// parse the suite file
	suite, err := setupSuite(suiteFile)
	if err != nil {
		log.Fatal().Msgf("[fony] - fatal error encountered parsing suite: [%s]", err.Error())
	}

	//  extract any global defaults from suite
	if suite.DefaultContentType != "" {
		defaultContentType = suite.DefaultContentType
	}

	// register the suite's endpoints with the chi router
	router := chi.NewRouter()
	if err := registerSuite(router, suite); err != nil {
		log.Fatal().Msgf("[fony] - fatal error encountered registering suite: [%s]", err.Error())
	}

	_ = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), router)
}
