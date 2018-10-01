package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/rs/zerolog/log"
)

// getEchoFunction will get the appropriate echo handler for the specified verb
func getEchoFunction(e *echo.Echo, ep *FonyEndpoint) (func(string, echo.HandlerFunc, ...echo.MiddlewareFunc), error) {
	verb := strings.ToUpper(ep.Method)
	switch verb {
	case http.MethodGet:
		return e.GET, nil
	case http.MethodDelete:
		return e.DELETE, nil
	case http.MethodHead:
		return e.HEAD, nil
	case http.MethodOptions:
		return e.OPTIONS, nil
	case http.MethodPut:
		return e.PUT, nil
	case http.MethodPatch:
		return e.PATCH, nil
	case http.MethodPost:
		return e.POST, nil
	case http.MethodTrace:
		return e.TRACE, nil
	case http.MethodConnect:
		return e.CONNECT, nil
	}

	err := fmt.Errorf("unknown HTTP method %s in endpoint: %+v", verb, ep)
	log.Error().Err(err).Str("verb", verb).Str("endpoint", ep.URL).Msg("[FONY]: unable to create route")
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

	f(ep.URL, func(ctx echo.Context) error {
		log.Info().Str("method", ep.Method).Str("url", ep.URL).Msg("[FONY]: handling request")

		// By default, all responses return this header.  It can be overwritten in the global
		// headers or the endpoint-specific header
		ctx.Response().Header().Set("Content-Type", "application/json")
		for k, v := range globals {
			ctx.Response().Header().Set(k, v)
		}

		// get the index header if set
		indexHeader := ctx.Request().Header.Get(FonyPayloadIndexHeader)
		var index uint64
		var parseIntErr error
		if indexHeader != "" {
			index, parseIntErr = strconv.ParseUint(indexHeader, 10, 8)
			if parseIntErr != nil {
				log.Error().Err(parseIntErr).Str("index_header", indexHeader).Msg("[FONY]: error parsing index header")
				errOut(ctx.Response())
				return parseIntErr
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
			log.Error().Err(err).Msg("no response found")
			errOut(ctx.Response())
			return err
		}

		// specific endpoint headers will override globals if set
		for k := range response.Headers {
			ctx.Response().Header().Set(k, response.Headers[k])
		}

		var data []byte
		if response.Payload != nil {
			ct := ctx.Response().Header().Get("Content-Type")
			if ct == "application/json" {
				var jsonErr error
				data, jsonErr = json.Marshal(response.Payload)
				if jsonErr != nil {
					log.Error().Err(jsonErr).Msg("[FONY]: payload marshal error")
					errOut(ctx.Response())
					return jsonErr
				}
			} else {
				switch response.Payload.(type) {
				case string:
					data = []byte(response.Payload.(string))
				case []byte:
					data = response.Payload.([]byte)
				default:
					err := fmt.Errorf("unable to process payload of type '%T'", response.Payload)
					log.Error().Err(err).Msg("[FONY]: payload error")
					errOut(ctx.Response())
					return err
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
		ctx.Response().WriteHeader(response.StatusCode)
		ctx.Response().Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		ctx.Response().Write(data)
		return nil
	})

	return nil
}
