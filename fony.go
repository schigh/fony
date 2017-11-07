package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"goji.io"
	"goji.io/pat"
)

// PhonyEndpoint represents a single endpoint served by fony
type PhonyEndpoint struct {
	URL             string            `json:"url"`
	Verb            string            `json:"verb"`
	ResponseHeaders map[string]string `json:"response_headers"`
	ResponsePayload interface{}       `json:"response_payload"`
	ResponseCode    int               `json:"response_code"`
}

// PhonySuite encloses the suite of endpoints served by fony
type PhonySuite struct {
	GlobalHeaders map[string]string `json:"global_headers"`
	Endpoints     []*PhonyEndpoint  `json:"endpoints"`
}

// patFunction is an alias to pat http verb functions
type patFunction func(url string) *pat.Pattern

var (
	suiteFile string
)

func init() {
	flag.StringVar(&suiteFile, "f", "./suite.json", "Absolute path to the fony suite file")
}

func setup() (*PhonySuite, bool) {
	flag.Parse()

	fileInfo, err := os.Stat(suiteFile)
	if err != nil {
		log.Errorf("fony suite file located at %s could not be found", suiteFile)
		return nil, false
	}
	if fileInfo.IsDir() {
		log.Errorf("the path at %s is a directory", suiteFile)
		return nil, false
	}

	data, err := ioutil.ReadFile(suiteFile)
	if err != nil {
		log.Errorf("Error reading suite file: %+v", err)
		return nil, false
	}

	suite := &PhonySuite{}
	ext := filepath.Ext(suiteFile)
	switch ext {
	case ".json":
		if err = json.Unmarshal(data, suite); err != nil {
			log.Errorf("Unable to parse suite file: %+v", err)
			return nil, false
		}
	default:
		log.Errorf("Unknown file extension: %s.  Only json files are acceptable", ext)
		return nil, false
	}

	return suite, true
}

func register(suite *PhonySuite) (*goji.Mux, bool) {
	mux := goji.NewMux()
	for _, ep := range suite.Endpoints {
		if err := processEndpoint(ep, mux, suite.GlobalHeaders); err != nil {
			log.Errorf("Error registering endpoint %s: %+v", ep.URL, err)
			return nil, false
		}
	}

	return mux, true
}

func getPatFunction(ep *PhonyEndpoint) (patFunction, error) {
	verb := strings.ToUpper(ep.Verb)
	switch verb {
	case "GET":
		return pat.Get, nil
	case "DELETE":
		return pat.Delete, nil
	case "HEAD":
		return pat.Head, nil
	case "OPTIONS":
		return pat.Options, nil
	case "PUT":
		return pat.Put, nil
	case "PATCH":
		return pat.Patch, nil
	case "POST":
		return pat.Post, nil
	}

	return nil, fmt.Errorf("unknown HTTP method: %s", verb)
}

func processEndpoint(ep *PhonyEndpoint, mux *goji.Mux, globals map[string]string) error {
	f, err := getPatFunction(ep)
	if err != nil {
		return err
	}

	mux.HandleFunc(f(ep.URL), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		for k := range globals {
			w.Header().Set(k, globals[k])
		}
		for k := range ep.ResponseHeaders {
			w.Header().Set(k, ep.ResponseHeaders[k])
		}

		data, err := json.Marshal(ep.ResponsePayload)
		if err != nil {
			log.Errorf("Payload marshal error: %+v", err)
			return
		}

		w.WriteHeader(ep.ResponseCode)
		w.Write(data)
	})

	return nil
}

func main() {
	suite, ok := setup()
	if !ok {
		os.Exit(1)
	}

	mux, ok := register(suite)
	if !ok {
		os.Exit(1)
	}

	log.Fatal(http.ListenAndServe("0.0.0.0:80", mux))
}
