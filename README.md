
![Build](https://github.com/serban-badila/form3-account-client/actions/workflows/build.yml/badge.svg)

`docker-compose up` will build and run the tests in a dedicated container. 

Alternatively, there is a `Makefile` provided for convenience but you will need the `CGO_ENABLED` for the make commands. 

Integration and load tests require the docker network up and running.

 - The user has the option to set a global timeout for executing each operation, then the client handles internally the retries and the backoff periods.
 
 - Has builtin structured logging.

 - All the functions bound to the client are safe to be used concurrently. 

### Example usage
```
package main

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"

	ac "go.form3-client.com/account"
)

func main() {
	apiHostAddress := "http://localhost:8080"
	httpClient := &http.Client{Timeout: ac.ClientTimeout}
	accountClient := ac.NewAccountClient(apiHostAddress, httpClient)

	accountData := ac.AccountData{
		ID:             uuid.New().String(),
		OrganisationID: uuid.New().String(),
		Type:           "accounts",
		Attributes: &ac.AccountAttributes{
			Country: "GB",
			Name:    []string{"John", "Doe"},
		},
	}
	accountId, err := accountClient.CreateAccount(&accountData)
	if err != nil {
		fmt.Println(accountId)
	}
}

```