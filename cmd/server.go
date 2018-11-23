package main

import (
	"github.com/go-chi/chi"

	"github.com/schigh/fony/domain"
)

func registerSuite(router *chi.Mux, suite *domain.Suite) error {
	_, err := domain.EndpointSlice(suite.Endpoints).TryEach(func(endpoint *domain.Endpoint) error {
		return processEndpoint(router, endpoint)
	})
	if err != nil {
		return err
	}

	return nil
}
