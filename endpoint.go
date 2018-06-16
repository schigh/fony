package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"go.uber.org/zap"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// getEchoFunction will get the appropriate echo handler for the specified verb
func getEchoFunction(e *echo.Echo, ep *FonyEndpoint) (func(path string, h echo.Handler), error) {
	verb := strings.ToUpper(ep.Method)
	switch verb {
	case http.MethodGet:
		return e.Get, nil
	case http.MethodDelete:
		return e.Delete, nil
	case http.MethodHead:
		return e.Head, nil
	case http.MethodOptions:
		return e.Options, nil
	case http.MethodPut:
		return e.Put, nil
	case http.MethodPatch:
		return e.Patch, nil
	case http.MethodPost:
		return e.Post, nil
	case http.MethodTrace:
		return e.Trace, nil
	case http.MethodConnect:
		return e.Connect, nil
	}

	err := fmt.Errorf("unknown HTTP method %s in endpoint: %+v", verb, ep)
	logger.Error("unknown method", zap.Error(err))
	return nil, err
}

// errOut processes internal errors not related to the payload
func errOut(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(500)
	w.Write([]byte("Fony failed to process this endpoint.  Check your logs to find the root cause"))
}

// processEndpoint will set each endpoint and handler in the muxxer
func processEndpoint(ep *FonyEndpoint, e *echo.Echo, globals map[string]string) error {
	f, verbErr := getEchoFunction(e, ep)
	if verbErr != nil {
		return verbErr
	}

	if ep.RunSequential {
		sequenceMap.sequences[ep.URL] = &endpointSequence{
			capacity: len(ep.Responses),
		}
	} else {
		sequenceMap.sequences[ep.URL] = &endpointSequence{
			capacity: -1,
		}
	}

	f(ep.URL, func(w http.ResponseWriter, r *http.Request) {
		logger.Info(fmt.Sprintf("%s: %s", ep.Method, ep.URL))

		// By default, all responses return this header.  It can be overwritten in the global
		// headers or the endpoint-specific header
		w.Header().Set("Content-Type", "application/json")
		for k, v := range globals {
			w.Header().Set(k, v)
		}

		// get the index header if set
		indexHeader := r.Header.Get(FonyPayloadIndexHeader)
		var index uint64
		var parseIntErr error
		if indexHeader != "" {
			index, parseIntErr = strconv.ParseUint(indexHeader, 10, 8)
			if parseIntErr != nil {
				logger.Error("fony: error parsing index", zap.Error(parseIntErr))
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
			err := fmt.Errorf("there must be at least one response per endpoint")
			logger.Error("no response", zap.Error(err))
			errOut(w)
			return
		}

		// specific endpoint headers will override globals if set
		for k := range response.Headers {
			w.Header().Set(k, response.Headers[k])
		}

		var data []byte
		if response.Payload != nil {
			ct := w.Header().Get("Content-Type")
			if ct == "application/json" {
				var jsonErr error
				data, jsonErr = json.Marshal(response.Payload)
				if jsonErr != nil {
					logger.Error("payload marshal error", zap.Error(jsonErr))
					errOut(w)
					return
				}
			} else {
				switch response.Payload.(type) {
				case string:
					data = []byte(response.Payload.(string))
				case []byte:
					data = response.Payload.([]byte)
				default:
					logger.Error(fmt.Sprintf("fony: unable to process payload of type '%T'", response.Payload))
					errOut(w)
					return
				}
			}
		}

		if response.Delay > 0 {
			// filter the delay between 1 and 10000 ms
			delay := time.Duration(math.Min(math.Max(1, response.Delay), 10000))
			time.Sleep(delay * time.Millisecond)
		}

		// default to 200 if not set
		if response.StatusCode == 0 {
			response.StatusCode = http.StatusOK
		}
		w.WriteHeader(response.StatusCode)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		w.Write(data)

		logger.Debug("response data", zap.ByteString("data", data))
	})

	return nil
}
