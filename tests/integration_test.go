//go:build integration
// +build integration

// these tests assume the service is accessible with the default configuration

package tests

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.form3-client.com/account"
)

const HOST_ADDRESS = "http://localhost:8080"

func TestCreateAccount(t *testing.T) {
	client := account.NewAccountClient(HOST_ADDRESS, account.ClientTimeout)

	type testCase struct {
		name             string
		givenAccountdata *account.AccountData
		expectedErrror   error
		expectedResponse string
	}
	cases := []testCase{
		{
			name: "succeeds with minimal account data",
			givenAccountdata: &account.AccountData{
				ID:             uuid.New().String(),
				OrganisationID: uuid.New().String(),
				Type:           "accounts", // this seems to be required for the Create operation although the APIdocs say its optional...
				Attributes: &account.AccountAttributes{
					Country: "RO",
					Name:    []string{"Serban", "Badila"},
				},
			},
			expectedErrror: nil,
		},
		{
			name:             "account type is required",
			givenAccountdata: AccountDataFactory.MustCreateWithOption(map[string]interface{}{"Type": "invalid type"}).(*account.AccountData),
			expectedErrror:   errors.New("validation failure list:\nvalidation failure list:\ntype in body should be one of [accounts]"),
		},
		{
			name: "country attrribute is required",
			givenAccountdata: AccountDataFactory.MustCreateWithOption(
				map[string]interface{}{"Attributes.Country": ""}).(*account.AccountData),
			expectedErrror: errors.New("validation failure list:\nvalidation failure list:\nvalidation failure list:\ncountry in body is required"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			resp, err := client.CreateAccount(tc.givenAccountdata)

			// THEN
			assert.Equal(t, tc.expectedErrror, err)
			if err == nil {
				assert.Equal(t, tc.givenAccountdata.ID, resp)
			}
		})
	}
}

func TestCreateAccountWithExistingIDFails(t *testing.T) {
	client := account.NewAccountClient(HOST_ADDRESS, account.ClientTimeout)
	fixedID := uuid.New().String()

	// WHEN
	accountData := AccountDataFactory.MustCreateWithOption(map[string]interface{}{"ID": fixedID}).(*account.AccountData)
	var err error
	_, err = client.CreateAccount(accountData)
	assert.Nil(t, err)

	// THEN
	anotherAccountData := AccountDataFactory.MustCreateWithOption(map[string]interface{}{"ID": fixedID}).(*account.AccountData)
	_, err = client.CreateAccount(anotherAccountData)
	assert.Error(t, err)
}
