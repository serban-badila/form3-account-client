//go:build unit
// +build unit

package tests

import (
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
			w.Write([]byte(`{"error_message": "user won't care about this error"}`))
			attempt++
		} else {
			w.WriteHeader(200)
			w.Write([]byte(fmt.Sprintf(`{"data": {"id": "%s"}}`, testId)))
		}
	}))
	defer serverWithInternalError.Close()

	// THEN
	client := account.NewAccountClient(serverWithInternalError.URL, account.ClientTimeout)
	id, err := client.CreateAccount(&account.AccountData{})
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
	client := account.NewAccountClient(serverWithInternalError.URL, account.ClientTimeout)
	id, err := client.CreateAccount(&account.AccountData{})
	assert.Equal(t, "", id)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), fmt.Sprintf("exceeded %v client's total timeout", account.ClientTimeout)))
}

// Does not use a request timeout
func TestCreateAccountSuceedsWhenServerRespondsSlowly(t *testing.T) {
	// WHEN
	serverTakesTooLongToRepond := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(account.ClientTimeout/2 + time.Duration(20*time.Millisecond))
		w.WriteHeader(200)
		w.Write([]byte(`{"data": {"id": "dummy id"}}`))
	}))
	defer serverTakesTooLongToRepond.Close()

	// THEN
	client := account.NewAccountClient(serverTakesTooLongToRepond.URL, account.ClientTimeout)
	id, err := client.CreateAccount(&account.AccountData{}) // acount data does not matter in this case
	assert.Nil(t, err)
	assert.Equal(t, "dummy id", id)
}

// Will wait for as long as it takes
func TestClientWithoutTimeoutWaitsIndefinitely(t *testing.T) {
	// WHEN
	serverTakesTooLongToRepond := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(account.ClientTimeout + time.Duration(20*time.Millisecond))
		w.WriteHeader(200)
		w.Write([]byte(`{"data": {"id": "dummy id"}}`))
	}))
	defer serverTakesTooLongToRepond.Close()

	// THEN
	client := account.NewAccountClient(serverTakesTooLongToRepond.URL, time.Duration(0))
	id, err := client.CreateAccount(&account.AccountData{}) // acount data does not matter in this case
	assert.Nil(t, err)
	assert.Equal(t, "dummy id", id)
}
