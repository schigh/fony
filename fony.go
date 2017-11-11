package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"goji.io"
	"goji.io/pat"
)

// FonyResponse contains the data returned in a specific response
type FonyResponse struct {
	Headers    map[string]string `json:"headers"`
	Payload    interface{}       `json:"payload"`
	StatusCode int               `json:"status_code"`
}

// FonyEndpoint represents a single endpoint served by fony
type FonyEndpoint struct {
	URL       string         `json:"url"`
	Verb      string         `json:"verb"`
	Responses []FonyResponse `json:"responses"`
}

// FonySuite encloses the suite of endpoints served by fony
type FonySuite struct {
	GlobalHeaders map[string]string `json:"global_headers"`
	Endpoints     []FonyEndpoint    `json:"endpoints"`
}

// patFunction is an alias to pat http verb functions
type patFunction func(url string) *pat.Pattern

// FonyPayloadIndexHeader is the index of the payload we wish to see for a given endpoint
const FonyPayloadIndexHeader = "X-Fony-Index"

var (
	suiteFile string
)

func init() {
	flag.StringVar(&suiteFile, "f", "./suite.json", "Absolute path to the fony suite file")
}

// setup prepares the suite for service
func setup() (*FonySuite, bool) {
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

	suite := &FonySuite{}
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

// getPatFunction will get the appropriate pat function based on HTTP verb
func getPatFunction(ep FonyEndpoint) (patFunction, error) {
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

func errOut(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(500)
	w.Write([]byte("Fony failed to process this endpoint.  Check your logs to find the root cause"))
}

// processEndpoint will set each endpoint and handler in the muxxer
func processEndpoint(ep FonyEndpoint, mux *goji.Mux, globals map[string]string) error {
	f, err := getPatFunction(ep)
	if err != nil {
		return err
	}

	mux.HandleFunc(f(ep.URL), func(w http.ResponseWriter, r *http.Request) {
		// by default, all responses return this header.  It can be overwritten in the global
		// headers or the endpoint-specific header
		w.Header().Set("Content-Type", "application/json")
		for k := range globals {
			w.Header().Set(k, globals[k])
		}

		// get the index header if set
		var index uint64
		indexHeader := r.Header.Get(FonyPayloadIndexHeader)
		if indexHeader != "" {
			index, err = strconv.ParseUint(indexHeader, 10, 8)
			if err != nil {
				log.Errorf("header index parse error: %+v", err)
				errOut(w)
				return
			}
		}

		var response *FonyResponse
		numResponses := len(ep.Responses)

		if numResponses > int(index) {
			response = &ep.Responses[index]
		} else if numResponses > 0 {
			response = &ep.Responses[0]
		} else {
			// there must be at least one response
			log.Error("There must be at least one response per endpoint")
			errOut(w)
			return
		}

		// specific endpoint headers will override globals if set
		for k := range response.Headers {
			w.Header().Set(k, response.Headers[k])
		}

		var data []byte
		if response.Payload != nil {
			data, err = json.Marshal(response.Payload)
			if err != nil {
				log.Errorf("payload marshal error: %+v", err)
				return
			}
		}

		if response.StatusCode == 0 {
			response.StatusCode = 200
		}
		w.WriteHeader(response.StatusCode)
		w.Write(data)
	})

	return nil
}

// register will register each endpoint in the muxxer
func register(suite *FonySuite) (*goji.Mux, bool) {
	mux := goji.NewMux()
	for _, ep := range suite.Endpoints {
		if err := processEndpoint(ep, mux, suite.GlobalHeaders); err != nil {
			log.Errorf("Error registering endpoint %s: %+v", ep.URL, err)
			return nil, false
		}
	}

	return mux, true
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
