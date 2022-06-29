//go:build unit
// +build unit

package account

import (
	"bytes"
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
		if attempt == 0 {
			w.WriteHeader(500)
			w.Write([]byte(`{"error_message": "user won't care about this error"}`))
			attempt++
		} else {
			w.WriteHeader(200)
			w.Write([]byte(fmt.Sprintf(`{"data": {"id": "%s"}}`, testId)))
		}
	}))
	defer serverWithInternalError.Close()

	// THEN
	client := NewAccountClient(serverWithInternalError.URL, ClientTimeout)
	id, err := client.CreateAccount(&AccountData{})
	assert.Equal(t, testId, id)
	assert.Nil(t, err)
}

// Global timeout kicks in if server is down
func TestCreateAccountFailsWith5xxResponse(t *testing.T) {
	// WHEN
	serverWithInternalError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error_message": "user won't care about this error"}`))
	}))
	defer serverWithInternalError.Close()

	// THEN
	timeout := time.Duration(2 * time.Second)
	client := NewAccountClient(serverWithInternalError.URL, timeout)
	id, err := client.CreateAccount(&AccountData{})
	assert.Equal(t, "", id)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), fmt.Sprintf("exceeded %v client's total timeout", timeout)))
}

// Does not use a request timeout
func TestCreateAccountSuceedsWhenServerRespondsSlowly(t *testing.T) {
	// WHEN
	timeout := time.Duration(2 * time.Second)
	serverTakesTooLongToRepond := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(timeout/2 + time.Duration(20*time.Millisecond))
		w.WriteHeader(200)
		w.Write([]byte(`{"data": {"id": "dummy id"}}`))
	}))
	defer serverTakesTooLongToRepond.Close()

	// THEN
	client := NewAccountClient(serverTakesTooLongToRepond.URL, timeout)
	id, err := client.CreateAccount(&AccountData{}) // acount data does not matter in this case
	assert.Nil(t, err)
	assert.Equal(t, "dummy id", id)
}

// Will wait for as long as it takes with not errors logged
func TestClientWithoutTimeoutWaitsIndefinitely(t *testing.T) {
	// WHEN
	serverTakesTooLongToRepond := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(2 * time.Second))
		w.WriteHeader(200)
		w.Write([]byte(`{"data": {"id": "dummy id"}}`))
	}))
	defer serverTakesTooLongToRepond.Close()
	client := NewAccountClient(serverTakesTooLongToRepond.URL, time.Duration(0))
	var buf bytes.Buffer
	client.logger = client.logger.Output(&buf) // redirect logs to a buffer so we can assert them

	// THEN

	id, err := client.CreateAccount(&AccountData{}) // acount data does not matter in this case
	assert.Nil(t, err)
	assert.Equal(t, "dummy id", id)

	fetchedData, err := client.GetById("does not matter")
	assert.Nil(t, err)
	assert.Equal(t, fetchedData.ID, "dummy id")
	assert.Equal(t, 0, buf.Len())
}
