package account

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

// NewAccountClient create a client for a given host and with a specified http client. The timeout includes any
// retries.
func NewAccountClient(url string, client *http.Client) *AccountClient {
	newLogger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	ac := &AccountClient{
		url:         url,
		contentType: "application/vnd.api+json",
		httpClient:  client,
		logger:      newLogger,
	}

	return ac
}

func (ac *AccountClient) GetById(ctx context.Context, accountId string) (*AccountData, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/organisation/accounts/%s", ac.url, accountId), nil)
	if err != nil {
		return nil, fmt.Errorf("got an error while creating the request: %w", err)
	}

	result, err := ac.executeRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	return result.accountData, err

}

// CreateAccount upon succcessful account creation, returns the updated account object and a nil error
func (ac *AccountClient) CreateAccount(ctx context.Context, account *AccountData) (*AccountData, error) {
	encoded, err := json.Marshal(createRequestBody{Data: account})
	if err != nil {
		return &AccountData{}, fmt.Errorf("could not json encode account data: %w", err)
	}

	buffer := bytes.NewBuffer(encoded)
	request, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/v1/organisation/accounts", ac.url), buffer)
	if err != nil {
		return &AccountData{}, fmt.Errorf("got an error while creating the request: %w", err)
	}

	result, err := ac.executeRequest(ctx, request)
	if err != nil {
		return &AccountData{}, err
	}
	return result.accountData, err
}

func (ac *AccountClient) DeleteAccount(ctx context.Context, accountId string, version int64) error {

	request, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/v1/organisation/accounts/%s", ac.url, accountId), nil)
	if err != nil {
		return fmt.Errorf("got an error while creating the request: %w", err)
	}
	querry := url.Values{}
	querry.Add("version", fmt.Sprint(version))
	request.URL.RawQuery = querry.Encode()

	_, err = ac.executeRequest(ctx, request)

	return err
}

func (ac *AccountClient) UpdateAccount(ctx context.Context, account *AccountData) (*AccountData, error) {
	encoded, err := json.Marshal(createRequestBody{Data: account})
	if err != nil {
		return &AccountData{}, fmt.Errorf("could not json encode account data: %w", err)
	}

	buffer := bytes.NewBuffer(encoded)
	request, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/v1/organisation/account/%s", ac.url, account.ID), buffer)
	if err != nil {
		return &AccountData{}, fmt.Errorf("got an error while creating the request: %w", err)
	}
	request.Header.Set("Accept", ac.contentType)
	result, err := ac.executeRequest(ctx, request)
	if err != nil {
		return &AccountData{}, err
	}
	return result.accountData, err
}

func (ac *AccountClient) executeRequest(ctx context.Context, req *http.Request) (*processedResult, error) {
	req.Header.Set("Content-Type", ac.contentType)
	respChan := make(chan *processedResult, 1)
	ctxWithLogger := ac.logger.WithContext(ctx)
	go handleRequest(ctxWithLogger, respChan, ac.httpClient, req)

	select {
	case <-ctx.Done():
		return &processedResult{}, ctx.Err()
	case result := <-respChan:
		if result.err != nil {
			return nil, result.err
		}
		return result, nil
	}
}
