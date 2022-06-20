package account

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// handleRequest deals with retries and is controlled by the parent context
func handleRequest(ctx context.Context, resultChan chan *ProcessedResult, client *http.Client, request *http.Request) {
	retries := 0
	for {
		select { // exit early if context is cancelled
		case <-ctx.Done():
			break
		default:
		}

		if !handleRequestOnce(ctx, resultChan, client, request) {
			break
		}

		noise := rand.Int()%100 - 50
		backoff := int(math.Pow(1.5, float64(retries)))*500 + noise
		time.Sleep(time.Duration(backoff) * time.Millisecond)
		retries++
	}
}

// handleRequestOnce and return a boolean whether the request should be retried
func handleRequestOnce(ctx context.Context, resultChan chan *ProcessedResult, client *http.Client, request *http.Request) bool {
	body, statusCode, err := doAndReadBody(client, request)
	if err != nil {
		resultChan <- &ProcessedResult{nil, fmt.Errorf("got an error while processing the http request: %w", err)}
		return false
	}

	var deserializedOk CreateOkBody
	var deserializedNotOk CreateErrorBody

	switch statusCode {
	case 204: // can receive this on DELETE
		resultChan <- &ProcessedResult{nil, nil}
		return false
	case 200, 201:
		{
			err := json.Unmarshal(body, &deserializedOk)
			if err != nil {
				resultChan <- &ProcessedResult{nil, fmt.Errorf("unable to deserialize response body; error: %w", err)}
			}
			resultChan <- &ProcessedResult{deserializedOk.Data, nil}
			return false
		}
	case 400, 401, 403, 404, 405, 406, 409:
		{
			json.Unmarshal(body, &deserializedNotOk)
			resultChan <- &ProcessedResult{nil, fmt.Errorf(deserializedNotOk.ErrorMessage)}
			return false
		}
	case 429, 500, 502, 503, 504:
		return true
	default:
		{ // what if the server starts redirecting ?
			resultChan <- &ProcessedResult{nil, fmt.Errorf("unexpected response status code: %d", statusCode)}
			return false
		}
	}
}

func doAndReadBody(client *http.Client, request *http.Request) ([]byte, int, error) {
	resp, err := client.Do(request) // when the response is not nil these may be caused by redirects only;
	// and the from the API docs, the server doesn't redirect
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	return body, resp.StatusCode, nil
}
