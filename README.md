### Run all the tests:   
As requested, `docker-compose up` will build and run the tests in a dedicated container


### Things found about the API Server


- `PATCH` responses are breaking the contract specified in the API Docs, by responding with the payload `{"code": "PAGE_NOT_FOUND", "message": "Page not found"}`. I guess this is not even implemented. I found this while trying to test that DELETE can erase outdated versions, when I realized the only way I can increment the version as a client was through a PATCH (I left the client implementation in with an integration test).


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
		OrganisationID: "uuid-string",
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

All the functions bound to this client are safe to run concurrently.
