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
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
)

type SegmentClient struct {
	transport *http.Transport
	nodeID    int
}

func NewSegmentClient(nodeID int) *SegmentClient {
	t := &http.Transport{}
	return &SegmentClient{
		t,
		nodeID,
	}
}

func (s *SegmentClient) RequestTransmission(address string, handle string, newHandle string, dtype string, opcode string) error {
	endpoint, err := resolveSegmentURL(address, routeTransmitNodeID, handle, dtype, newHandle, s.nodeID, opcode)
	if err != nil {
		return err
	}

	err = request(s.transport, endpoint).
		withMethod(http.MethodPost).
		expect(http.StatusOK).
		submit()

	if err != nil {
		return err
	}

	return nil
}

func (s *SegmentClient) ReceiveTransmission(address string, handle string, tc types.TypeCode, index int, writer io.Writer) (int, error) {
	endpoint, err := resolveSegmentURL(address, routeTransmitIndexID, handle, tc, fmt.Sprint(index))
	var arraylength int = 0
	if err != nil {
		return arraylength, err
	}

	arraylength, err = request(s.transport, endpoint).
		withMethod(http.MethodGet).
		expect(http.StatusOK).
		retrieveDownload(writer)

	if err != nil {
		return arraylength, err
	}

	return arraylength, nil
}

func resolveSegmentURL(address string, route RoutePattern, params ...interface{}) (string, error) {
	if len(address) == 0 {
		return "", errors.New("Address is empty. cannot resolve.")
	}
	url := strings.Split(address, ":")
	add := url[0]
	port, _ := strconv.Atoi(url[1])

	return route.ToURL(add, int32(port), params...)
}
