//go:build load
// +build load

// same as the integration tests, these assume the API server is running on the default host

package tests

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.form3-client.com/account"
)

const HOST_ADDRESS = "http://localhost:8080"

type result struct {
	id  string
	err error
}

func TestConcurrentCreates(t *testing.T) {
	iterations := 20
	resultChan := make(chan *result, iterations)
	ac := account.NewAccountClient(HOST_ADDRESS, account.ClientTimeout)

	fireConcurentCreates(ac, iterations, resultChan)

	for i := 0; i < iterations; i++ {
		select {
		case res := <-resultChan:
			assert.NotEqual(t, res.id, "")
			assert.Nil(t, res.err)
		}
	}
}

// TestStressAPI overload the server and check all requests are handled properly by the client
// I am surprised here because  according to the API docs I would
// expect a 429 status http.Response from the server.
// it seems I am atually able to DDOS the API server and its database...
//
// also while looking into the API container logs I sometimes see the database has too many open connections
// and again I would expect the server be in charge internally of properly managing the db connections
func TestDDOS(t *testing.T) {
	timeout := time.Duration(10) * time.Second
	wayTooManyIterations := 1000
	resultChan := make(chan *result, wayTooManyIterations)
	ac := account.NewAccountClient(HOST_ADDRESS, timeout)

	fireConcurentCreates(ac, wayTooManyIterations, resultChan)

	var errorSample error
	for i := 0; i < wayTooManyIterations; i++ {
		var res *result
		select {
		case res = <-resultChan:
		}
		if errorSample = res.err; errorSample != nil {
			break
		} else if i == wayTooManyIterations-1 { // in the rare event the server can handle this much load
			return
		}

	}
	assert.NotNil(t, errorSample)
	assert.True(t,
		strings.Contains(errorSample.Error(), fmt.Sprintf("exceeded %v client's total timeout", timeout)) || strings.Contains(errorSample.Error(), "does not exist"))
}

func fireConcurentCreates(ac *account.AccountClient, iterations int, resultChan chan<- *result) {
	for i := 0; i < iterations; i++ {
		go func() {
			id, err := ac.CreateAccount(AccountDataFactory.MustCreate().(*account.AccountData))
			resultChan <- &result{id, err}
		}()
	}
}
