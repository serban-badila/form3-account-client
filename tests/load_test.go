//go:build load
// +build load

// same as the integration as a tests, these assume the API server is running on the default host

package tests

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.form3-client.com/account"
)

type result struct {
	acc *account.AccountData
	err error
}

func TestConcurrentCreates(t *testing.T) {
	iterations := 20
	resultChan := make(chan *result, iterations)
	ac := account.NewAccountClient(fetchAPIHostName(), account.ClientTimeout)

	fireConcurentCreates(ac, iterations, resultChan)

	for i := 0; i < iterations; i++ {
		select {
		case res := <-resultChan:
			assert.NotEqual(t, res.acc.ID, "")
			assert.Nil(t, res.err)
		}
	}
}

/*
I am surprised here because according to the API docs I would
expect a 429 status http.Response from the server.

the database has too many open connections (looking at the container logs)
and I would expect the server be in charge internally of properly managing the db connections

it seems I am atually able to DDOS the API server and its database...

also due to some error messages of the form "account id %%% already exists" I can guess the server doesn't interrupt execution in case the client closes the connection
I sometimes saw these when putting the server under load while the client was retrying the POST due to higher response times
*/
func TestDDOS(t *testing.T) {
	timeout := time.Duration(10) * time.Second
	wayTooManyIterations := 1000
	resultChan := make(chan *result, wayTooManyIterations)
	ac := account.NewAccountClient(fetchAPIHostName(), timeout)

	fireConcurentCreates(ac, wayTooManyIterations, resultChan)

	var errorSample error
	for i := 0; i < wayTooManyIterations; i++ {
		var res *result
		select {
		case res = <-resultChan:
		}
		if errorSample = res.err; errorSample != nil {
			break
		} else if i == wayTooManyIterations-1 { // in the (never observed) event the server can handle this much load
			return
		}

	}
	assert.NotNil(t, errorSample)
	assert.True(t,
		strings.Contains(errorSample.Error(), fmt.Sprintf("exceeded %v client's total timeout", timeout)) || strings.Contains(errorSample.Error(), "does not exist"))
} // the second error string is due to the client interrupting connections then retrying to send the POST while the server is overloaded

func fireConcurentCreates(ac *account.AccountClient, iterations int, resultChan chan<- *result) {
	for i := 0; i < iterations; i++ {
		go func() {
			acc, err := ac.CreateAccount(AccountDataFactory.MustCreate().(*account.AccountData))
			resultChan <- &result{acc, err}
		}()
	}
}
