package account

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	ACCOUNTS_HOST_ADDR = "ACCOUNTS_HOST_ADDR"
	ERROR_KEY          = "error_message"
	ID_KEY             = "id"
)

type AccountClient struct {
	Url string
}

func (ac *AccountClient) GetById(id string) (*AccountData, error) {
	return &AccountData{
		Attributes:     nil,
		ID:             id,
		OrganisationID: "my-org",
		Type:           "accounts",
		Version:        0,
	}, nil

	// &AccountData{
	// 	Attributes: &AccountAttributes{
	// 		AccountClassification:   "",
	// 		AccountMatchingOptOut:   false,
	// 		AccountNumber:           "",
	// 		AlternativeNames:        []string{},
	// 		BankID:                  "",
	// 		BankIDCode:              "",
	// 		BaseCurrency:            "",
	// 		Bic:                     "",
	// 		Country:                 "",
	// 		Iban:                    "",
	// 		JointAccount:            false,
	// 		Name:                    []string{},
	// 		SecondaryIdentification: "",
	// 		Status:                  "",
	// 		Switched:                false,
	// 	},
	// 	ID:             id,
	// 	OrganisationID: "",
	// 	Type:           "accounts",
	// 	Version:        int64(0),
	// }
}

type AccountCreateBody struct {
	Data *AccountData `json:"data,omitempty"`
}

func (ac *AccountClient) CreateAccount(account *AccountData) (string, error) {
	client := &http.Client{}

	encoded, err := json.Marshal(AccountCreateBody{Data: account})
	if err != nil {
		return "", fmt.Errorf("Could not json encode account data: %w", err)
	}

	buffer := bytes.NewBuffer(encoded)
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/organisation/accounts", ac.Url), buffer)
	request.Header.Set("Content-Type", "application/vnd.api+json")

	resp, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp)
	body, _ := ioutil.ReadAll(resp.Body)

	var deserialized interface{}
	json.Unmarshal(body, &deserialized)

	errorMsg, ok := deserialized.(map[string]interface{})[ERROR_KEY]
	if ok {
		return "", fmt.Errorf(errorMsg.(string))
	} else {
		return deserialized.(map[string]interface{})["data"].(map[string]interface{})[ID_KEY].(string), nil
	}
}

func NewAccountClient() *AccountClient {
	var host string
	host, ok := os.LookupEnv(ACCOUNTS_HOST_ADDR)
	if !ok {
		fmt.Printf("Missing %s in environment! Defaulting to localhost", ACCOUNTS_HOST_ADDR)
		host = "http://localhost:8080"
	}

	return &AccountClient{Url: host}
}
