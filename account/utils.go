package account

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

type RetryOnError struct {
	s   int   // response's status code
	err error // response's error
}

func (e RetryOnError) Error() string {
	return fmt.Sprintf("Response status code: %v; with error: %s", e.s, e.err)
}

// handleRequest deals with retries and is controlled by the parent context
func handleRequest(ctx context.Context, resultChan chan *processedResult, client *http.Client, request *http.Request) {
	var retryErr RetryOnError
	retries := 0
	for {
		select { // exit early if context is cancelled
		case <-ctx.Done():
			break
		default:
		}

		err := handleRequestOnce(ctx, resultChan, client, request)
		if !errors.As(err, &retryErr) {
			break
		}

		noise := rand.Int()%100 - 50
		backoff := int(math.Pow(1.5, float64(retries)))*500 + noise
		after := time.Duration(backoff) * time.Millisecond
		log.Ctx(ctx).Info().Str("endpoint", request.URL.Path).Msg(fmt.Sprintf("Retrying in %v", after))
		time.Sleep(after)
		retries++
	}
}

// handleRequestOnce and return an error whether the request should be retried
func handleRequestOnce(ctx context.Context, resultChan chan *processedResult, client *http.Client, request *http.Request) error {
	body, statusCode, err := doAndReadBody(client, request)
	var errorString string
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			errorString = urlErr.Error()
			log.Ctx(ctx).Error().Str("type", "RequestError").Bool("timeout", urlErr.Timeout()).Str("endpoint", urlErr.URL).Msg(errorString)
			return RetryOnError{statusCode, urlErr}
		} else {
			errorString = err.Error()
			log.Ctx(ctx).Error().Str("type", "ReadError").Msg(errorString)
			resultChan <- &processedResult{nil, fmt.Errorf("got an error while reading the response body: %w", err)}
			return nil
		}
	}

	var deserializedOk createOkBody
	var deserializedNotOk createErrorBody

	switch statusCode {
	case 204: // can receive this on DELETE
		resultChan <- &processedResult{nil, nil}
		return nil
	case 200, 201:
		{
			err := json.Unmarshal(body, &deserializedOk)
			if err != nil {
				resultChan <- &processedResult{nil, fmt.Errorf("unable to deserialize response body; error: %w", err)}
			}
			resultChan <- &processedResult{deserializedOk.Data, nil}
			return nil
		}
	case 400, 401, 403, 404, 405, 406, 409:
		{
			json.Unmarshal(body, &deserializedNotOk)
			resultChan <- &processedResult{nil, fmt.Errorf("response status code %d with error message: %s", statusCode, deserializedNotOk.ErrorMessage)}
			return nil
		}
	case 429, 500, 502, 503, 504:
		json.Unmarshal(body, &deserializedNotOk)
		log.Ctx(ctx).Error().Str("type", "ResponseError").Int("responseStatus", statusCode).Str("endpoint", request.URL.Path).Msg(deserializedNotOk.ErrorMessage)
		return RetryOnError{statusCode, errors.New(deserializedNotOk.ErrorMessage)}
	default:
		{ // what if the server starts redirecting ?
			resultChan <- &processedResult{nil, fmt.Errorf("unexpected response status code: %d", statusCode)}
			return nil
		}
	}
}

func doAndReadBody(client *http.Client, request *http.Request) ([]byte, int, error) {
	resp, err := client.Do(request) // when the response is not nil these may be caused by redirects only;
	// and from the API docs, the server doesn't redirect
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	return body, resp.StatusCode, nil
}
