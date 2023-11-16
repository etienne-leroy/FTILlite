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
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/pkg/errors"
)

const (
	acceptRangeHeader  = "Accept-Ranges"
	rangeHeader        = "Range"
	contentRangeHeader = "Content-Range"
)

// Logger is an optional interface used for outputting debug logging
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// DownloadOptions are a set of options used while downloading a stream
type DownloadOptions struct {
	Timeout             time.Duration
	InitialHeadTimeout  time.Duration
	RetryWait           time.Duration
	Retries             int
	RetryWaitMultiplier float64
	BufferSize          int
	HTTPTransport       http.RoundTripper
	Logger              Logger
}

const (
	headTimeoutSeconds = 60
	retries            = 10
	//bufferSizeKb        = 128
	bufferSizeKb        = 1
	bytesInKb           = 1024
	retryWaitMultiplier = 1.61803398875 //GoldenMean
)

// DefaultOptions are the default options used when no options are specified by users of the library.
// Call this function to get a new default options struct where you can adjust only the things you need to.
func DefaultOptions() DownloadOptions {
	return DownloadOptions{
		Timeout:             time.Hour,
		InitialHeadTimeout:  time.Second * headTimeoutSeconds, //TODO: Make this shorter? Set to 60 for debugging
		Retries:             retries,
		RetryWait:           time.Second,
		RetryWaitMultiplier: retryWaitMultiplier, // Golden
		BufferSize:          bufferSizeKb * bytesInKb,
	}
}

// Infof checks if a logger is present and logs to it at info level if it is
func (o *DownloadOptions) Infof(format string, args ...interface{}) {
	if o.Logger == nil {
		return
	}
	o.Logger.Infof(format, args...)
}

// Errorf checks if a logger is present and logs to it at error level if it is
func (o *DownloadOptions) Errorf(format string, args ...interface{}) {
	if o.Logger == nil {
		return
	}
	o.Logger.Errorf(format, args...)
}

// DownloadStream streams the file at the given URL to the given writer while retrying any broken connections.
// Consecutive calls with the same URL and filePath will attempt to resume the download.
// The download stream written to the writer will then be replayed from the beginning of the download.
// With the given context the whole operation can be aborted.
func DownloadStream(ctx context.Context, url string, writer io.Writer) (arraylength int, err error) {
	return DownloadStreamOpts(ctx, url, writer, DefaultOptions())
}

// DownloadStreamOpts is the same as DownloadStream, but allows you to override the default options with own values.
// See DownloadStream for more information.
func DownloadStreamOpts(ctx context.Context, url string, writer io.Writer, options DownloadOptions) (arraylength int, err error) {
	contentLength, resumable, arraylength, err := fetchURLInfoTries(ctx, url, &options)
	if err != nil {
		return 0, err
	}

	// This is where the logic for fully resumable download should be added.
	written := int64(0)
	if contentLength == -1 {
		return arraylength, nil
	}

	if written == contentLength {
		return 0, fmt.Errorf("The HTTP download is already completed.")
	}

	if !resumable {
		options.Infof("dlstream.DownloadStreamOpts: Download not resumable")
	} else if written > 0 {
		options.Infof("dlstream.DownloadStreamOpts: Current file size: %d, resuming", written)
	}

	// If we are resuming a download, copy the existing file contents to the writer for replay
	if written > 0 {
		if err != nil {
			options.Errorf("dlstream.DownloadStreamOpts: Error seeking start of stream: %v", err)
			return 0, errors.Wrap(err, "error seeking file to start")
		}

		if err != nil {
			options.Errorf("dlstream.DownloadStreamOpts: Error replaying file stream: %v", err)
			return 0, errors.Wrap(err, "error replaying file stream")
		}
	}

	return startDownloadTries(ctx, url, contentLength, written, writer, &options)
}

// fetchURLInfoTries tries the configured amount of attempts at doing a HEAD request
// See FetchURLInfo for more information
func fetchURLInfoTries(ctx context.Context, url string, options *DownloadOptions) (contentLength int64, resumable bool, arraylength int, err error) {
	for i := 0; i < options.Retries; i++ {
		contentLength, resumable, arraylength, err = FetchURLInfo(ctx, url, options.InitialHeadTimeout, options.HTTPTransport)
		if err != nil && shouldRetryRequest(err) {
			options.Infof("dlstream.fetchURLInfoTries: Error fetching URL info: %v, retrying request", err)
			retryWait(options)
			continue
		}
		if err != nil {
			options.Errorf("dlstream.fetchURLInfoTries: Error fetching URL info: %v, unrecoverable error, will not retry", err)
			return -1, false, 0, err
		}

		options.Infof("dlstream.fetchURLInfoTries: Download size: %d, Resumable: %v", contentLength, resumable)
		return contentLength, resumable, arraylength, nil
	}

	return -1, false, arraylength, fmt.Errorf("The number of retrieval retries has been exceeded.")
}

// FetchURLInfo does a HEAD request to see if the download URL is valid and returns the size of the content.
// Also checks if the download can be resumed by looking at capability headers.
func FetchURLInfo(ctx context.Context, url string, timeout time.Duration, transport http.RoundTripper) (contentLength int64, resumable bool, arraylength int, err error) {
	client := http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return -1, false, 0, errors.Wrap(err, "error creating head request")
	}
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return -1, false, 0, errors.Wrap(err, "error requesting url")
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return -1, false, 0, errors.Errorf("unexpected head status code %d", resp.StatusCode)
	}

	err = resp.Body.Close()
	if err != nil {
		return -1, false, 0, errors.Wrap(err, "error closing response body")
	}

	// Resumable is only possible if we can request ranges and know how big the file is gonna be.
	resumable = resp.Header.Get(acceptRangeHeader) == "bytes" && resp.ContentLength > 0

	arrayLength, err := strconv.Atoi(resp.Header.Get(ArrayElementLengthHeader))
	if err != nil {
		arrayLength = GetArrayLengthFromURL(url)
	}

	return resp.ContentLength, resumable, arrayLength, nil
}

func GetArrayLengthFromURL(url string) int {
	x := strings.Split(url, "/")
	typecode := types.TypeCode(x[len(x)-3])

	if typecode.IsBytearray() {
		return typecode.Length()
	}
	return 0
}

// startDownloadTries starts a loop that retries the download until it either finishes or the retries are depleted
func startDownloadTries(
	ctx context.Context, url string, contentLength, written int64, writer io.Writer, options *DownloadOptions,
) (arraylength int, err error) {
	buffer := make([]byte, options.BufferSize)

	// Loop that retries the download
	for i := 0; i < options.Retries; i++ {
		options.Infof("dlstream.startDownloadTries: Downloading %s from offset %d, total size: %d, attempt %d", url, written, contentLength, i)

		var bodyReader io.ReadCloser
		var arrayLength int = 0
		bodyReader, arrayLength, err = doDownloadRequest(ctx, url, written, contentLength, options)
		if err != nil && shouldRetryRequest(err) {
			options.Infof("dlstream.startDownloadTries: Error retrieving URL: %v, retrying request", err)
			retryWait(options)
			continue
		} else if err != nil {
			options.Errorf("dlstream.startDownloadTries: Error retrieving URL: %v, unrecoverable error, will not retry", err)
			return arrayLength, err
		}

		var shouldContinue bool
		written, shouldContinue, err = doCopyRequestBody(bodyReader, buffer, contentLength, written, writer, options)
		if shouldContinue {
			continue
		}
		return arrayLength, err
	}

	return arraylength, fmt.Errorf("The number of retrieval retries has been exceeded.")
}

// doCopyRequestBody copies the request body to the file and writer and reports back errors and progress
//
//revive:disable-next-line:cognitive-complexity TODO: Review cognitive complexity and possible refactor?
func doCopyRequestBody(
	bodyReader io.ReadCloser, buffer []byte, contentLength, written int64, writer io.Writer, options *DownloadOptions,
) (newWritten int64, shouldContinue bool, err error) {
	// Byte loop that copies from the download reader to the file and writer
	for {
		var bytesRead int
		var writerErr error
		bytesRead, err = bodyReader.Read(buffer)
		if bytesRead > 0 {
			_, writerErr = writer.Write(buffer[:bytesRead])
			if writerErr != nil {
				// If the writer at any point returns an error, we should abort and do nothing further
				closeErr := bodyReader.Close()
				if closeErr != nil {
					options.Errorf("dlstream.doCopyRequestBody: Error closing body reader: %v", err)
				}
				return written, false, writerErr // Bounce back the error
			}
		}

		written += int64(bytesRead)

		if err == io.EOF {
			_ = bodyReader.Close()
			if written != contentLength {
				options.Errorf("dlstream.doCopyRequestBody: Download done yet incomplete, total: %d, expected: %d", written, contentLength)
				return written, false, fmt.Errorf("The content length is not the expected length.")
			}
			options.Infof("dlstream.doCopyRequestBody: Download complete, %d bytes", written)
			return written, false, nil // YES, we have a complete download :)
		}
		if err != nil && shouldRetryRequest(err) {
			_ = bodyReader.Close()
			options.Infof(
				"dlstream.doCopyRequestBody: Error reading from response body: %v, total: %d, currently written: %d, retrying",
				err, contentLength, written,
			)
			retryWait(options)
			return written, true, err
		}
		if err != nil {
			_ = bodyReader.Close()
			options.Errorf(
				"dlstream.doCopyRequestBody: Error reading from response body: %v, total: %d, currently written: %d, unrecoverable error, will not retry",
				err, contentLength, written,
			)
			return written, false, err
		}
	}
}

func retryWait(options *DownloadOptions) {
	options.Infof("dlstream.retryWait: Waiting for %v", options.RetryWait)
	time.Sleep(options.RetryWait)
	options.RetryWait = time.Duration(float64(options.RetryWait) * options.RetryWaitMultiplier)
}

// doDownloadRequest sends an actual download request and returns the content length (again) and response body reader
//
//revive:disable-next-line:cognitive-complexity TODO: Review cognitive complexity and possible refactor?
//revive:disable-next-line:cyclomatic TODO: Review cyclomatic complexity and possible refactor?
func doDownloadRequest(ctx context.Context, url string, downloadFrom, totalContentLength int64, options *DownloadOptions) (body io.ReadCloser, arrayElementLength int, err error) {
	client := http.Client{
		Timeout:   options.Timeout,
		Transport: options.HTTPTransport,
	}

	// See: https://stackoverflow.com/a/29200933/3536354
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req = req.WithContext(ctx)

	if downloadFrom > 0 {
		req.Header.Set(rangeHeader, fmt.Sprintf("bytes=%d-", downloadFrom))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, errors.Wrap(err, "error requesting url")
	}

	a, err := strconv.Atoi(resp.Header.Get(ArrayElementLengthHeader))
	if err != nil {
		a = 0
	}

	if downloadFrom <= 0 {
		if resp.StatusCode == http.StatusNoContent {
			return nil, a, nil
		}
		if resp.StatusCode != http.StatusOK {
			return nil, a, errors.Errorf("unexpected download http status code %d", resp.StatusCode)
		}
		if resp.ContentLength != totalContentLength {
			return nil, a, errors.Errorf("unexpected response content-length (expected %d, got %d)", totalContentLength, resp.ContentLength)
		}
	} else {
		if resp.StatusCode != http.StatusPartialContent {
			return nil, a, errors.Errorf("unexpected download http status code %d", resp.StatusCode)
		}

		var respStart, respEnd, respTotal int64
		_, err = fmt.Sscanf(
			strings.ToLower(resp.Header.Get(contentRangeHeader)),
			"bytes %d-%d/%d",
			&respStart, &respEnd, &respTotal,
		)

		if err != nil {
			return nil, a, errors.Wrap(err, "error parsing response content-range header")
		}
		if respStart != downloadFrom {
			return nil, a, errors.Errorf("unexpected response range start (expected %d, got %d)", downloadFrom, respStart)
		}
		if respEnd != totalContentLength-1 {
			return nil, a, errors.Errorf("unexpected response range end (expected %d, got %d)", totalContentLength-1, respEnd)
		}
		if respTotal != totalContentLength {
			return nil, a, errors.Errorf("unexpected response range total (expected %d, got %d)", totalContentLength, respTotal)
		}
	}

	return resp.Body, a, nil
}

// shouldRetryRequest analyzes a given request error and determines whether its a good idea to retry the request
//
//revive:disable-next-line:cyclomatic TODO: Review cyclomatic complexity
//revive:disable-next-line:cognitive-complexity TODO: Review cognitive complexity
func shouldRetryRequest(err error) (shouldRetry bool) {
	if err == io.ErrUnexpectedEOF {
		return true
	}

	netErr, ok := err.(net.Error)
	if ok {
		return netErr.Temporary() || netErr.Timeout()
	}

	switch t := err.(type) {
	case *net.OpError:
		if t.Op == "dial" {
			// Unknown host
			return false
		}
		if t.Op == "read" {
			// Connection refused
			return true
		}

	case syscall.Errno:
		if t == syscall.ECONNREFUSED {
			// Connection refused
			return true
		}
		if t == syscall.ECONNRESET {
			// Connection reset
			return true
		}
		if t == syscall.ECONNABORTED {
			// Connection aborted
			return true
		}
	}

	return false
}

func getSize(stream io.Reader) int {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(stream)
	if err != nil {
		return -1
	}
	return buf.Len()
}
