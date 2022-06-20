package account

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	ERROR_KEY       = "error_message"
	ID_KEY          = "id"
	CLIENT_TIMEOUT  = time.Duration(5 * time.Second) // This includes any retries; also impacts some unit tests runtime
	REQUEST_TIMEOUT = time.Duration(1 * time.Second) // per request
)

// AccountClient All the bound methods are safe to run as coroutines
type AccountClient struct {
	url         string
	timeout     time.Duration
	contentType string
	httpClient  *http.Client
}

func (ac *AccountClient) GetById(id string) (*AccountData, error) {
	return &AccountData{
		Attributes:     nil,
		ID:             id,
		OrganisationID: "my-org",
		Type:           "accounts",
		Version:        0,
	}, nil
}

// CreateAccount upon succcessfull account creation, returns the uuid of the account object
func (ac *AccountClient) CreateAccount(account *AccountData) (string, error) {
	encoded, err := json.Marshal(CreateRequestBody{Data: account})
	if err != nil {
		return "", fmt.Errorf("could not json encode account data: %w", err)
	}
	buffer := bytes.NewBuffer(encoded)
	ctx, cancel := context.WithTimeout(context.Background(), ac.timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/v1/organisation/accounts", ac.url), buffer)
	if err != nil {
		return "", fmt.Errorf("got an error while creating the request: %w", err)
	}
	request.Header.Set("Content-Type", ac.contentType)

	respChan := make(chan *ProcessedResult, 1)
	go handleRequest(ctx, respChan, ac.httpClient, request)

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("exceeded %v client's total timeout while trying to create the account", ac.timeout)
	case result := <-respChan:
		if result.err != nil {
			return "", result.err
		}
		// fmt.Printf("%v", result.accountData)
		return result.accountData.ID, nil
	}
}

func NewAccountClient(url string) *AccountClient {
	return &AccountClient{
		url:         url,
		timeout:     CLIENT_TIMEOUT,
		contentType: "application/vnd.api+json",
		httpClient:  &http.Client{Timeout: REQUEST_TIMEOUT},
	}
}
