//go:build unit
// +build unit

package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.form3-client.com/account"
)

// Retry on receiving a 5xx code from the server
func TestCreateAccountSucceedsAfter5xxResponse(t *testing.T) {
	// WHEN
	attempt := 0
	testId := "test id"
	serverWithInternalError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempt == 0 {
			w.WriteHeader(500)
			w.Write([]byte(`{"error_message": "client won't care about this error"}`))
			attempt++
		} else {
			w.WriteHeader(200)
			serialized, _ := json.Marshal(account.CreateOkBody{Data: &account.AccountData{ID: testId}})
			w.Write([]byte(serialized))
		}
	}))
	defer serverWithInternalError.Close()

	// THEN
	client := account.NewAccountClient(serverWithInternalError.URL)
	id, err := client.CreateAccount(&account.AccountData{})
	assert.Equal(t, id, testId)
	assert.Nil(t, err)
}

// Global timeout kicks in if server is down
func TestCreateAccountFailsWith5xxResponse(t *testing.T) {
	// WHEN
	serverWithInternalError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error_message": "client won't care about this error"}`))
	}))
	defer serverWithInternalError.Close()

	// THEN
	client := account.NewAccountClient(serverWithInternalError.URL)
	id, err := client.CreateAccount(&account.AccountData{})
	assert.Equal(t, id, "")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), fmt.Sprintf("exceeded %v client's total timeout", account.CLIENT_TIMEOUT)))
}

// Request timeout kicks in if server is too slow
func TestCreateAccountFailsWhenServerTimesOutOnRequest(t *testing.T) {
	// WHEN
	serverTakesTooLongToRepond := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(account.REQUEST_TIMEOUT + time.Duration(10*time.Millisecond)) // request timeout will tick before server responds
		w.WriteHeader(200)
		w.Write([]byte(`{"data": {"id": "dummy id"}}`))
	}))
	defer serverTakesTooLongToRepond.Close()

	// THEN
	client := account.NewAccountClient(serverTakesTooLongToRepond.URL)
	_, err := client.CreateAccount(&account.AccountData{}) // acount data does not matter in this case
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers"))
}
