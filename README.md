
![Build](https://github.com/serban-badila/form3-account-client/actions/workflows/build.yml/badge.svg)

As requested, `docker-compose up` will build and run the tests in a dedicated container. 

Alternatively, there is a `Makefile` provided for convenience but you will need the `CGO_ENABLED` for the make commands. 

Integration and load tests require the docker network up and running.


### Things found about the API Server

- This is the more serious one: it seems I am able to perform a DOS attack on this dummy server instance and it happily forwards the attack onto the database instead of managing it (by rate limiting for instance); You can read more about this in the load tests. I can imagine in reality this kind of service sits behind an ingress or an  apiGateway and it may rely on these to manage incoming requests.

- The API docs are a bit misleading: some attributes are described as optional whereas they are required (e.g. the account's `type`), others are described as (conditionally) required when they are not (some of the bank account's attributes for instance). Some integration tests check for this.

- `PATCH` responses are breaking the contract specified in the API Docs, by responding with the payload `{"code": "PAGE_NOT_FOUND", "message": "Page not found"}`. I guess this is not even implemented. Found this while trying to test that DELETE can erase outdated versions, when I realized the only way I can increment the version as a client was through a `PATCH` (I left the implementation in with an integration test).



Ok, but enough about your service, now about this client: 
- The idea is to transparently forward any data validation errors back to the user, this is why data models have almost all fields optional (the client doesn't have much of an opinion about how the account attribute should look like). But there are some integration tests that check how and if some attributes are validated (I used these tests more like a behaviour discovery).

 - The user has the option to set a global timeout for executing each operation, then the client handles internally the retries and the backoff periods.
 
 - Has builtin structured logging.

 - All the functions bound to the client are safe to be used (abused) concurrently. These methods use internal contexts to synchronize the connections. 

### Example usage
```
package main

import (
	"fmt"

	"github.com/google/uuid"

	ac "go.form3-client.com/account"
)

func main() {
	apiHostAddress := "http://localhost:8080"
	accountClient := ac.NewAccountClient(apiHostAddress, ac.ClientTimeout)

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