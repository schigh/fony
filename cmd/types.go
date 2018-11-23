package main

import "sync"

type endpointSequence struct {
	capacity int
	next     int
}

type endpointSequenceMap struct {
	sync.Mutex
	sequences map[string]*endpointSequence
}

// FonyPayloadIndexHeader is the index of the payload we wish to see for a given endpoint
const FonyPayloadIndexHeader = "X-Fony-Index"
