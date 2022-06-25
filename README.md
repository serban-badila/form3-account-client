### Run all the tests:   
  `make test-all`


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