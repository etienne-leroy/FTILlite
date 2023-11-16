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
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type handlerContextKey struct{}
type newHandlerContextKey struct{}
type nodeContextKey struct{}
type opcodeContextKey struct{}
type dtypeContextKey struct{}
type indexContextKey struct{}

func parseHandlerID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := chi.URLParam(r, "handler_id")
		ctx := context.WithValue(r.Context(), handlerContextKey{}, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func parseNewHandlerID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := chi.URLParam(r, "newhandler_id")
		ctx := context.WithValue(r.Context(), newHandlerContextKey{}, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func parseNodeID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := chi.URLParam(r, "node_id")
		ctx := context.WithValue(r.Context(), nodeContextKey{}, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func parseOpcode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := chi.URLParam(r, "opcode")
		ctx := context.WithValue(r.Context(), opcodeContextKey{}, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func parseDtype(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := chi.URLParam(r, "dtype")
		ctx := context.WithValue(r.Context(), dtypeContextKey{}, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func parseIndex(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := chi.URLParam(r, "index")
		ctx := context.WithValue(r.Context(), indexContextKey{}, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
