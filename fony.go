package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"goji.io"
	"goji.io/pat"
	"gopkg.in/yaml.v2"
	"net/http"
)

// PhonyEndpoint represents a single endpoint served by fony
type PhonyEndpoint struct {
	URL             string            `json:"url" yaml:"url"`
	Verb            string            `json:"verb" yaml:"verb"`
	ResponseHeaders map[string]string `json:"response_headers" yaml:"response_headers"`
	ResponsePayload interface{}       `json:"response_payload" yaml:"response_payload"`
	ResponseCode    int               `json:"response_code" yaml:"response_code"`
}

// PhonySuite encloses the suite of endpoints served by fony
type PhonySuite struct {
	GlobalHeaders map[string]string `json:"global_headers" yaml:"global_headers"`
	Endpoints     []PhonyEndpoint   `json:"endpoints" yaml:"endpoints"`
}

// PatFunction is an alias to pat http verb functions
type PatFunction func(url string) *pat.Pattern

var (
	suiteFile string
)

func init() {
	flag.StringVar(&suiteFile, "f", "./suite.json", "Absolute pat to the fony suite file")
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
	case "json":
		if err = json.Unmarshal(data, suite); err != nil {
			log.Errorf("Unable to parse suite file: %+v", err)
			return nil, false
		}
	case "yml", "yaml":
		if err = yaml.Unmarshal(data, suite); err != nil {
			log.Errorf("Unable to parse suite file: %+v", err)
			return nil, false
		}
	default:
		log.Errorf("Unknown file extension: %s.  Only json or yaml files are acceptable", ext)
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
}

func processEndpoint(ep *PhonyEndpoint, mux *goji.Mux, globals map[string]string) error {
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
