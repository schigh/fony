package main

import "sync"

// FonyResponse contains the data returned in a specific response
type FonyResponse struct {
	// Headers are the expected headers returned in the fake response
	Headers map[string]string `json:"headers"`

	// Payload is the object (currently json only) returned in the response body
	Payload interface{} `json:"payload"`

	// StatusCode is the HTTP response code returned by the fake request
	StatusCode int `json:"status_code"`

	// Delay is the length of time, in milliseconds, that the response will block
	Delay float64 `json:"delay"`
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

type endpointSequence struct {
	capacity int
	next     int
}

type endpointSequenceMap struct {
	sync.Mutex
	sequences map[string]*endpointSequence
}
