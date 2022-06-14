//go:build integration
// +build integration

// these tests assume the service is accessible with the default configuration

package account

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.form3-client.com/account"
)

func TestCreateAccount(t *testing.T) {
	client := account.NewAccountClient()

	type testCase struct {
		name             string
		givenAccountdata *account.AccountData
		expctedErrror    error
		expectedResponse string
	}
	cases := []testCase{
		{
			name: "minimnal account data",
			givenAccountdata: &account.AccountData{
				ID:             uuid.New().String(),
				OrganisationID: uuid.New().String(),
				Type:           "accounts", // this seems to be required for the Create operation although the APIdocs say its optional...
				Attributes: &account.AccountAttributes{
					Country: "RO",
					Name:    []string{"Serban", "Badila"},
				},
			},
			expctedErrror: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := client.CreateAccount(tc.givenAccountdata)
			assert.Equal(t, tc.expctedErrror, err)
			assert.Equal(t, tc.givenAccountdata.ID, resp)
		})
	}
}

func TestCreateAccountWithExistingID(t *testing.T) {
	client := account.NewAccountClient()
	fixedID := uuid.New().String()

	// when
	accountData := AccountDataFactory.MustCreateWithOption(map[string]interface{}{"ID": fixedID}).(*account.AccountData)
	client.CreateAccount(accountData)

	// then
	anotherAccountData := AccountDataFactory.MustCreateWithOption(map[string]interface{}{"ID": fixedID}).(*account.AccountData)
	_, err := client.CreateAccount(anotherAccountData)
	assert.Error(t, err)
}
