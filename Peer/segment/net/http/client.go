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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type requestBuilder struct {
	transport      *http.Transport
	requestMethod  string
	requestURL     string
	requestBody    interface{}
	responseBody   interface{}
	expectedStatus int
	timeout        time.Duration
}

func request(transport *http.Transport, url string) *requestBuilder {
	return &requestBuilder{transport, http.MethodGet, url, nil, nil, http.StatusOK, 0}
}

func (b *requestBuilder) withMethod(method string) *requestBuilder {
	b.requestMethod = method
	return b
}

func (b *requestBuilder) withBody(value interface{}) *requestBuilder {
	b.requestBody = value
	return b
}
func (b *requestBuilder) withTimeout(timeInSeconds time.Duration) *requestBuilder {
	b.timeout = timeInSeconds
	return b
}
func (b *requestBuilder) decodeResponseInto(responseBody interface{}) *requestBuilder {
	b.responseBody = responseBody
	return b
}
func (b *requestBuilder) expect(status int) *requestBuilder {
	b.expectedStatus = status
	return b
}

//revive:disable-next-line:cyclomatic TODO: Has high cyclomatic complexity - candidate for refactor.
//revive:disable-next-line:cognitive-complexity TODO: Has high cognitive complexity - candidate for refactor.
func (b *requestBuilder) submit() error {
	client := http.Client{
		Transport: b.transport,
		Timeout:   b.timeout,
	}

	request, err := createRequest(b)
	if err != nil {
		return err
	}

	//request.Close = true

	if b.responseBody != nil {
		request.Header.Set("Accept", "application/json")
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != b.expectedStatus {
		if resp.StatusCode == http.StatusInternalServerError {
			var v string

			if err = json.NewDecoder(resp.Body).Decode(&v); err == nil {
				return fmt.Errorf("error: %v", err)
			}
		}

		return fmt.Errorf("%v %v: expected HTTP %v but server returned HTTP %v", b.requestMethod, b.requestURL, b.expectedStatus, resp.Status)
	}

	if b.responseBody != nil {
		if err = json.NewDecoder(resp.Body).Decode(&b.responseBody); err != nil {
			return fmt.Errorf("unable to decode server response: %w", err)
		}
	} else {
		_, _ = io.Copy(ioutil.Discard, resp.Body)
	}

	return nil
}

func (b *requestBuilder) retrieveDownload(writer io.Writer) (int, error) {
	downloadOptions := DefaultOptions()
	downloadOptions.HTTPTransport = b.transport
	arraylength, err := DownloadStreamOpts(context.Background(), b.requestURL, writer, downloadOptions)

	if err != nil {
		return arraylength, err
	}

	return arraylength, nil
}

//revive:disable-next-line:cyclomatic TODO: Has high cyclomatic complexity - candidate for refactor.
//revive:disable-next-line:cognitive-complexity TODO: Has high cognitive complexity - candidate for refactor.
func createRequest(b *requestBuilder) (*http.Request, error) {
	var body io.Reader
	var ok bool
	var contentType string
	var err error
	var request *http.Request

	if b.requestBody != nil {
		switch b.requestBody.(type) {
		case io.Reader:
			body, ok = b.requestBody.(io.Reader)
			if !ok {
				return nil, fmt.Errorf("unable to process io.Reader")
			}
			contentType = "application/octet-stream"
		default:
			jsonStr, err := json.Marshal(b.requestBody)
			if err != nil {
				return nil, err
			}

			body = bytes.NewBuffer(jsonStr)

			contentType = "application/json"
		}
		request, err = http.NewRequest(b.requestMethod, b.requestURL, body)
		if err != nil {
			return nil, err
		}
		request.Header.Set("Content-Type", contentType)
	}

	if b.requestBody == nil {
		request, err = http.NewRequest(b.requestMethod, b.requestURL, nil)
		if err != nil {
			return nil, err
		}
	}
	return request, nil
}
