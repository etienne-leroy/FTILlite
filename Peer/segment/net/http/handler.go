// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// EndpointSetupFunc represents a function which configures the endpoints on a Router
type EndpointSetupFunc func(r chi.Router)

// NewAnonHandler returns a new http.Handler that doesn't require certificate exchange.
func NewAnonHandler(setupFuncs ...EndpointSetupFunc) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.AllowContentType("application/json", "application/octet-stream"))

	for _, f := range setupFuncs {
		r.Group(f)
	}

	return r
}
