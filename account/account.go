package account

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
)

const (
	erorKey       = "error_message"
	idKey         = "id"
	ClientTimeout = time.Duration(5 * time.Second) // for convenience
)

// AccountClient All the bound methods are safe to run as coroutines
type AccountClient struct {
	url         string
	timeout     time.Duration
	contentType string
	httpClient  *http.Client
	logger      zerolog.Logger
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

// CreateAccount upon succcessfull account creation, returns the uuid of the account object and a nill error
func (ac *AccountClient) CreateAccount(account *AccountData) (string, error) {
	encoded, err := json.Marshal(createRequestBody{Data: account})
	if err != nil {
		return "", fmt.Errorf("could not json encode account data: %w", err)
	}

	var ctx context.Context
	if ac.timeout != time.Duration(0) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), ac.timeout)
		defer cancel()
	} else {
		ctx = context.Background()
	}
	buffer := bytes.NewBuffer(encoded)
	request, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/v1/organisation/accounts", ac.url), buffer)
	if err != nil {
		return "", fmt.Errorf("got an error while creating the request: %w", err)
	}
	request.Header.Set("Content-Type", ac.contentType)

	respChan := make(chan *processedResult, 1)
	ctxWithLogger := ac.logger.WithContext(ctx)
	go handleRequest(ctxWithLogger, respChan, ac.httpClient, request)

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("exceeded %v client's total timeout while trying to create the account", ac.timeout)
	case result := <-respChan:
		if result.err != nil {
			return "", result.err
		}
		return result.accountData.ID, nil
	}
}

// NewAccountClient create a client for a given host and with a specified timeout. The timeout includes any
// retries. Passing a time.Duration(0) will disable the timeout.
func NewAccountClient(url string, timeout time.Duration) *AccountClient {
	newLogger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	acc := &AccountClient{
		url:         url,
		contentType: "application/vnd.api+json",
		httpClient:  &http.Client{},
		logger:      newLogger,
	}
	if timeout.Nanoseconds() > 0 {
		acc.timeout = timeout
	}
	return acc
}
