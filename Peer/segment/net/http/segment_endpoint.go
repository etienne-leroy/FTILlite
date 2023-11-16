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
	"bytes"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

const forwardSlash = "/"
const ArrayElementLengthHeader string = "arrayelementlength"

var routeTransmit = NewRoutePattern("/transmit/{handler_id}/{dtype}")
var routeTransmitIndexID = routeTransmit.SubRoute("/rec/{index}")
var routeTransmitNewHandlerID = routeTransmit.SubRoute("/req/{newhandler_id}")
var routeTransmitNodeID = routeTransmitNewHandlerID.SubRoute("/{node_id}/{opcode}")

type SegmentHandler interface {
	TransferBytes(address string, handle string, newHandle string, dtype string, opcode string) error
	GetVariable(h variables.Handle) (types.TypeVal, error)
	RequestTransferBytes(nodeAddress string, handle string, newHandle string, dtype string, opcode string) error
	SegmentNodes() map[string]string
}

func SegmentEndpoints(h SegmentHandler) EndpointSetupFunc {
	return func(r chi.Router) {
		r.Route(routeTransmit.Pattern(), func(r chi.Router) {
			r.Use(parseHandlerID)
			r.Use(parseDtype)
			r.Route(routeTransmitIndexID.Pattern(), func(r chi.Router) {
				r.Use(parseIndex)
				r.Get(forwardSlash, getSegmentTransmitBytes(h))
				r.Head(forwardSlash, getSegmentTransmitBytes(h))
			})
			r.Route(routeTransmitNewHandlerID.Pattern(), func(r chi.Router) {
				r.Use(parseNewHandlerID)
				r.Route(routeTransmitNodeID.Pattern(), func(r chi.Router) {
					r.Use(parseNodeID)
					r.Use(parseOpcode)
					r.Post(forwardSlash, postSegmentTransmitBytes(h))
				})
			})
		})
	}
}

func getSegmentTransmitBytes(h SegmentHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handlerID := variables.Handle(r.Context().Value(handlerContextKey{}).(string))
		//dtype := types.TypeCode(r.Context().Value(dtypeContextKey{}).(string))
		temp := r.Context().Value(indexContextKey{}).(string)
		index, _ := strconv.Atoi(temp)

		v, err := h.GetVariable(handlerID)

		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
		}

		arraylength := 0

		if v.TypeCode().GetBase() == types.BytearrayB {
			arraylength = v.TypeCode().Length()
		}

		b, err := v.GetBinaryArray(index)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
		}

		if len(b) > 0 {
			reader := bytes.NewReader(b)

			w.Header().Set("Content-Type", "application/octet-stream")

			w.Header().Add(ArrayElementLengthHeader, fmt.Sprint(arraylength))

			t := time.Time{}
			http.ServeContent(w, r, "", t, reader)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func postSegmentTransmitBytes(h SegmentHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handlerID := variables.Handle(r.Context().Value(handlerContextKey{}).(string))
		newHandlerID := variables.Handle(r.Context().Value(newHandlerContextKey{}).(string))
		nodeID := variables.Handle(r.Context().Value(nodeContextKey{}).(string))
		dtype := r.Context().Value(dtypeContextKey{}).(string)
		opcode := r.Context().Value(opcodeContextKey{}).(string)

		nodeAddress := h.SegmentNodes()[string(nodeID)]
		err := h.TransferBytes(nodeAddress, string(handlerID), string(newHandlerID), dtype, opcode)
		if err != nil {
			log.Printf("Error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// RoutePattern is a convenience type for encapsulating a route path pattern so that it can be
// reused when constructing the URL on the client.
type RoutePattern struct {
	path string

	base *RoutePattern
}

// NewRoutePattern creates a new RoutePattern.
func NewRoutePattern(path string) RoutePattern {
	return RoutePattern{path, nil}
}

// Pattern returns the relative part of the path which belongs to this RoutePattern instance. It
// does not include the path of parent routes.
func (r RoutePattern) Pattern() string { return string(r.path) }

var paramRegexp = regexp.MustCompile(`{\w+}`)

// ToURL creates a URL to this route taking into consideration the parent routes. The route
// parameters are substituted, in order, with the parameters passed to this function.
func (r RoutePattern) ToURL(host string, port int32, substitutionValues ...interface{}) (string, error) {
	p := strings.TrimPrefix(r.path, forwardSlash)

	for b := r.base; b != nil; b = b.base {
		p = strings.TrimPrefix(b.path, forwardSlash) + forwardSlash + p
	}

	replaceCount := 0
	p = paramRegexp.ReplaceAllStringFunc(p, func(s string) string {
		replaceCount++

		if replaceCount > len(substitutionValues) {
			return ""
		}

		return fmt.Sprintf("%v", substitutionValues[replaceCount-1])
	})

	if replaceCount != len(substitutionValues) {
		return "", fmt.Errorf(
			"the number of substitution values (%v) did not match the number of placeholders (%v) in the route pattern",
			len(substitutionValues),
			replaceCount,
		)
	}

	return fmt.Sprintf("http://%v:%v/%v", host, port, p), nil
}

// SubRoute creates a new RoutePattern with the current RoutePattern as the parent route.
func (r RoutePattern) SubRoute(path string) RoutePattern {
	return RoutePattern{path, &r}
}
