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
	"sync"

	log "github.com/sirupsen/logrus"
	"goji.io"
	"goji.io/pat"
)

// FonyResponse contains the data returned in a specific response
type FonyResponse struct {
	// Headers are the expected headers returned in the fake response
	Headers map[string]string `json:"headers"`

	// Payload is the object (currently json only) returned in the response body
	Payload interface{} `json:"payload"`

	// StatusCode is the HTTP response code returned by the fake request
	StatusCode int `json:"status_code"`
}

// FonyEndpoint represents a single endpoint served by fony
type FonyEndpoint struct {
	// URL is the endpoint url
	URL string `json:"url"`

	// Method is the HTTP method used to perform the fake request
	Method string `json:"method"`

	// Responses is the list of possible responses for any given endpoint
	Responses []FonyResponse `json:"responses"`

	// RunSequential will run all responses in order, and start at the
	// first request if the list reaches its end
	RunSequential bool `json:"run_sequential"`
}

// FonySuite encloses the suite of endpoints served by fony
type FonySuite struct {
	// GlobalHeaders are request headers applied to all requests
	GlobalHeaders map[string]string `json:"global_headers"`

	// Endpoints is a list of fake endpoints
	Endpoints []*FonyEndpoint `json:"endpoints"`
}

type endpointSequenceMap struct {
	sync.Mutex
	sequences map[string]*endpointSequence
}

type endpointSequence struct {
	capacity int
	next     int
}

// patFunction is an alias to pat http verb functions
type patFunction func(url string) *pat.Pattern

// FonyPayloadIndexHeader is the index of the payload we wish to see for a given endpoint
const FonyPayloadIndexHeader = "X-Fony-Index"

var (
	suiteFile   string
	port        string
	sequenceMap *endpointSequenceMap
	logger      *log.Entry
)

func init() {
	flag.StringVar(&suiteFile, "f", "./fony.json", "Absolute path to the fony suite file")
	flag.StringVar(&port, "p", "80", "http port (for local testing)")
	logger = log.NewEntry(log.New())
}

// setup prepares the suite for service
func setup() (*FonySuite, bool) {
	flag.Parse()

	suite := &FonySuite{}

	// this just makes sure the port flag can be parsed to an integer
	_, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		logger.Errorf("Error parsing port: %v", err)
		return nil, false
	}

	// look for the presence of a remote suite file
	suiteURL := os.Getenv("SUITE_URL")
	if suiteURL != "" {
		logger.Infof("Running remote suite file located at %s", suiteURL)
		resp, err := http.Get(suiteURL)
		if err != nil {
			logger.Errorf("Error fetching remote suite file: %v", err)
			return nil, false
		}
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Errorf("Error reading remote file: %v", err)
			return nil, false
		}

		if err = json.Unmarshal(data, suite); err != nil {
			logger.Errorf("Unable to parse remote suite file: %v", err)
			return nil, false
		}
	} else {
		fileInfo, err := os.Stat(suiteFile)
		if err != nil {
			logger.Errorf("fony suite file located at %s could not be found", suiteFile)
			return nil, false
		}
		if fileInfo.IsDir() {
			logger.Errorf("the path at %s is a directory", suiteFile)
			return nil, false
		}

		data, err := ioutil.ReadFile(suiteFile)
		if err != nil {
			logger.Errorf("Error reading suite file: %v", err)
			return nil, false
		}

		ext := filepath.Ext(suiteFile)
		switch ext {
		case ".json":
			if err = json.Unmarshal(data, suite); err != nil {
				logger.Errorf("Unable to parse suite file: %v", err)
				return nil, false
			}
		default:
			logger.Errorf("Unknown file extension: %s.  Only json files are acceptable", ext)
			return nil, false
		}
	}

	return suite, true
}

// getPatFunction will get the appropriate pat function based on HTTP verb
func getPatFunction(ep *FonyEndpoint) (patFunction, error) {
	verb := strings.ToUpper(ep.Method)
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

	err := fmt.Errorf("unknown HTTP method %s in endpoint: %+v", verb, ep)
	logger.Error(err.Error())

	return nil, err
}

// errOut processes internal errors not related to the payload
func errOut(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(500)
	w.Write([]byte("Fony failed to process this endpoint.  Check your logs to find the root cause"))
}

// processEndpoint will set each endpoint and handler in the muxxer
func processEndpoint(ep *FonyEndpoint, mux *goji.Mux, globals map[string]string) error {
	f, err := getPatFunction(ep)
	if err != nil {
		return err
	}

	// if running sequentially, set the capacity with the next
	// response pointing to the zero index.  Otherwise, set
	// capacity to -1
	if ep.RunSequential {
		sequenceMap.sequences[ep.URL] = &endpointSequence{
			capacity: len(ep.Responses),
		}
	} else {
		sequenceMap.sequences[ep.URL] = &endpointSequence{
			capacity: -1,
		}
	}

	mux.HandleFunc(f(ep.URL), func(w http.ResponseWriter, r *http.Request) {
		// By default, all responses return this header.  It can be overwritten in the global
		// headers or the endpoint-specific header
		logger.Infof("%s: %s", ep.Method, ep.URL)
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
				logger.Errorf("header index parse error: %v", err)
				errOut(w)
				return
			}
		} else {
			sequenceMap.Lock()
			s := sequenceMap.sequences[ep.URL]
			if s.capacity != -1 {
				index = uint64(s.next)
				s.next++
				if s.next >= s.capacity {
					s.next = 0
				}
			}
			sequenceMap.Unlock()
		}

		var response *FonyResponse
		numResponses := len(ep.Responses)

		if numResponses > int(index) {
			response = &ep.Responses[index]
		} else if numResponses > 0 {
			response = &ep.Responses[0]
		} else {
			// there must be at least one response
			logger.Error("There must be at least one response per endpoint")
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
				logger.Errorf("payload marshal error: %v", err)
				errOut(w)
				return
			}
		}

		if response.StatusCode == 0 {
			response.StatusCode = http.StatusOK
		}
		w.WriteHeader(response.StatusCode)
		w.Write(data)

		logger.Infof("resp data: %s", string(data))
	})

	return nil
}

// register will register each endpoint in the muxxer
func register(suite *FonySuite) (*goji.Mux, bool) {
	mux := goji.NewMux()
	for _, ep := range suite.Endpoints {
		if err := processEndpoint(ep, mux, suite.GlobalHeaders); err != nil {
			logger.Errorf("Error registering endpoint %s: %v", ep.URL, err)
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

	sequenceMap = &endpointSequenceMap{
		sequences: make(map[string]*endpointSequence, len(suite.Endpoints)),
	}

	mux, ok := register(suite)
	if !ok {
		os.Exit(1)
	}

	logger.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), mux))
}
