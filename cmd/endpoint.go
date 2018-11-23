package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"

	"github.com/schigh/fony/domain"
)

// this runs when we done messed up...not when the client expects an error
func bailNow(rw http.ResponseWriter, err error) {
	rw.Header().Set("Content-Type", "application/json; charset=utf8")
	rw.Header().Set("X-Fony-Runtime-Error", "true")
	rw.WriteHeader(http.StatusInternalServerError)
	_, _ = rw.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
}

func getChiFunction(mux *chi.Mux, ep *domain.Endpoint) (func(string, http.HandlerFunc), error) {
	verb := strings.ToUpper(ep.Method)
	switch verb {
	case http.MethodGet:
		return mux.Get, nil
	case http.MethodPost:
		return mux.Post, nil
	case http.MethodPut:
		return mux.Put, nil
	case http.MethodDelete:
		return mux.Delete, nil
	case http.MethodPatch:
		return mux.Patch, nil
	case http.MethodOptions:
		return mux.Options, nil
	case http.MethodHead:
		return mux.Head, nil
	case http.MethodTrace:
		return mux.Trace, nil
	default:
		return nil, fmt.Errorf("unknown method '%s'", verb)
	}
}

func processEndpoint(mux *chi.Mux, ep *domain.Endpoint) error {
	f, err := getChiFunction(mux, ep)
	if err != nil {
		return err
	}

	capacity := -1
	if ep.RunSequential {
		capacity = len(ep.Responses)
	}
	sequenceMap.sequences[ep.URL] = &endpointSequence{
		capacity: capacity,
	}

	f(ep.URL, func(rw http.ResponseWriter, r *http.Request) {
		log.Info().Str("method", ep.Method).Str("url", ep.URL).Msg("[fony] - handling request")

		// default.  can be overridden
		rw.Header().Set("Content-Type", defaultContentType)

		// get index header if set
		indexHeader := r.Header.Get(FonyPayloadIndexHeader)
		var index uint64
		if indexHeader != "" {
			_index, convErr := strconv.ParseUint(indexHeader, 10, 8)
			if convErr != nil {
				log.Error().Err(convErr).Str("index_header", indexHeader).Msg("[fony] - error parsing index header")
				bailNow(rw, convErr)
				return
			}
			index = _index
		} else {
			// not looking for a particular index, which means we are running sequentially
			sequenceMap.Lock()
			s := sequenceMap.sequences[ep.URL]
			// capacity == -1 means that this is not a sequential request
			if s.capacity != -1 {
				index = uint64(s.next)
				s.next++
				if s.next >= s.capacity {
					s.next = 0
				}
			}
			sequenceMap.Unlock()
		}

		var response domain.Response
		numResponses := len(ep.Responses)

		switch {
		case numResponses > int(index):
			response = ep.Responses[index]
		case numResponses > 0:
			response = ep.Responses[0]
		default:
			err := errors.New("there must be at least one response per endpoint")
			log.Error().Err(err).Msg("[fony] - no response found")
			bailNow(rw, err)
			return
		}

		// specific endpoint headers will override globals if set
		for k := range response.Headers {
			rw.Header().Set(k, response.Headers[k])
		}

		var data []byte
		if response.Payload != nil {
			// json type
			if jsonRX.Match([]byte(rw.Header().Get("Content-Type"))) {
				_data, jsonErr := json.Marshal(response.Payload)
				if jsonErr != nil {
					log.Error().Err(jsonErr).Msg("[fony] - payload marshal error")
					bailNow(rw, jsonErr)
					return
				}
				data = _data
			} else {
				// other types
				switch response.Payload.(type) {
				case string:
					data = []byte(response.Payload.(string))
				case []byte:
					data = response.Payload.([]byte)
				default:
					err := fmt.Errorf("unable to process payload of type '%T'", response.Payload)
					log.Error().Err(err).Msg("[fony] - payload error")
					bailNow(rw, err)
					return
				}
			}
		}

		// delay the response if requested
		if response.Delay > 0 {
			// filter delay between 1  and 10000 ms
			delay := time.Duration(math.Min(math.Max(1, response.Delay), 10000))
			t := time.NewTimer(delay * time.Millisecond)
			<-t.C
		}

		// default status code to 200 if not set
		if response.StatusCode == 0 {
			response.StatusCode = http.StatusOK
		}

		// finish request
		rw.WriteHeader(response.StatusCode)
		rw.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		_, _ = rw.Write(data)
	})

	return nil
}
