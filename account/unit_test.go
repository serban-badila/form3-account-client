//go:build unit
// +build unit

package account

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Retry on receiving a 5xx code from the server
func TestCreateAccountSucceedsAfter5xxResponse(t *testing.T) {
	// WHEN
	attempt := 0
	testId := "test id"
	serverWithInternalError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempt < 2 {
			w.WriteHeader(500)
			w.Write([]byte(`{"error_message": "user won't care about this error"}`))
			attempt++
		} else {
			w.WriteHeader(200)
			w.Write([]byte(fmt.Sprintf(`{"data": {"id": "%s"}}`, testId)))
		}
	}))
	defer serverWithInternalError.Close()

	ctx := context.Background()
	// THEN

	httpClient := &http.Client{Timeout: ClientTimeout}
	client := NewAccountClient(serverWithInternalError.URL, httpClient)
	acc, err := client.CreateAccount(ctx, &AccountData{})
	assert.Equal(t, testId, acc.ID)
	assert.Nil(t, err)
}

// Context timeout kicks in if server is down
func TestCreateAccountFailsWith5xxResponse(t *testing.T) {
	// WHEN
	serverWithInternalError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error_message": "user won't care about this error"}`))
	}))
	defer serverWithInternalError.Close()
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(2*time.Second))
	defer cancel()

	// THEN
	timeout := time.Duration(2 * time.Second)
	httpClient := &http.Client{Timeout: timeout}
	client := NewAccountClient(serverWithInternalError.URL, httpClient)
	_, err := client.CreateAccount(ctx, &AccountData{})
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "context deadline exceeded"))
}

// Handles requests before the context's timeout
func TestCreateAccountSuceedsWhenServerRespondsSlowly(t *testing.T) {
	// WHEN
	timeout := time.Duration(2 * time.Second)
	serverTakesTooLongToRepond := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(timeout/2 + time.Duration(20*time.Millisecond))
		w.WriteHeader(200)
		w.Write([]byte(`{"data": {"id": "dummy id"}}`))
	}))
	defer serverTakesTooLongToRepond.Close()
	ctx := context.Background()

	// THEN
	httpClient := &http.Client{Timeout: timeout}
	client := NewAccountClient(serverTakesTooLongToRepond.URL, httpClient)
	acc, err := client.CreateAccount(ctx, &AccountData{}) // acount data does not matter in this case
	assert.Nil(t, err)
	assert.Equal(t, "dummy id", acc.ID)
}

// Will wait for as long as it takes without any errors logged
func TestClientWithoutTimeoutWaitsIndefinitely(t *testing.T) {
	// WHEN
	serverTakesTooLongToRepond := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(2 * time.Second))
		w.WriteHeader(200)
		w.Write([]byte(`{"data": {"id": "dummy id"}}`))
	}))
	defer serverTakesTooLongToRepond.Close()
	httpClient := &http.Client{}
	client := NewAccountClient(serverTakesTooLongToRepond.URL, httpClient)
	var buf bytes.Buffer
	client.logger = client.logger.Output(&buf) // redirect logs to a buffer so we can assert them
	ctx := context.Background()

	// THEN
	acc, err := client.CreateAccount(ctx, &AccountData{}) // acount data does not matter in this case
	assert.Nil(t, err)
	assert.Equal(t, "dummy id", acc.ID)

	fetchedData, err := client.GetById(ctx, "does not matter")
	assert.Nil(t, err)
	assert.Equal(t, fetchedData.ID, "dummy id")
	assert.Equal(t, 0, buf.Len())
}
