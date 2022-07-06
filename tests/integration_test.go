//go:build integration
// +build integration

// these tests assume the service is accessible with the default configuration

package tests

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.form3-client.com/account"
)

func TestCreateAccount(t *testing.T) {
	client := account.NewAccountClient(fetchAPIHostName(), account.ClientTimeout)

	type testCase struct {
		name             string
		givenAccountdata *account.AccountData
		expectedErrror   error
	}

	minimalAccountData := &account.AccountData{
		ID:             uuid.New().String(),
		OrganisationID: uuid.New().String(),
		Type:           "accounts", // this seems to be required for the Create operation although the APIdocs say its optional...
		Attributes: &account.AccountAttributes{
			Country: "RO",
			Name:    []string{"Serban", "Badila"},
		},
	}
	cases := []testCase{
		{
			name:             "succeeds with minimal account data",
			givenAccountdata: minimalAccountData,
			expectedErrror:   nil,
		},
		{
			name:             "account type is required",
			givenAccountdata: AccountDataFactory.MustCreateWithOption(map[string]interface{}{"Type": "invalid type"}).(*account.AccountData),
			expectedErrror:   errors.New("response status code 400 with error message: validation failure list:\nvalidation failure list:\ntype in body should be one of [accounts]"),
		},
		{
			name: "country attrribute is required",
			givenAccountdata: AccountDataFactory.MustCreateWithOption(
				map[string]interface{}{"Attributes.Country": ""}).(*account.AccountData),
			expectedErrror: errors.New("response status code 400 with error message: validation failure list:\nvalidation failure list:\nvalidation failure list:\ncountry in body is required"),
		},
		{
			name: "invalid account id",
			givenAccountdata: AccountDataFactory.MustCreateWithOption(
				map[string]interface{}{"ID": "invalid-id"}).(*account.AccountData),
			expectedErrror: errors.New("response status code 400 with error message: validation failure list:\nvalidation failure list:\nid in body must be of type uuid: \"invalid-id\""),
		},
		{
			name: "country is validated",
			givenAccountdata: AccountDataFactory.MustCreateWithOption(
				map[string]interface{}{"Attributes.Country": "invalid"}).(*account.AccountData),
			expectedErrror: errors.New("response status code 400 with error message: validation failure list:\nvalidation failure list:\nvalidation failure list:\ncountry in body should match '^[A-Z]{2}$'"),
		},
		{
			name: "bank id code is validated",
			givenAccountdata: AccountDataFactory.MustCreateWithOption(
				map[string]interface{}{"Attributes.BankIDCode": "WRONGID121212"}).(*account.AccountData),
			expectedErrror: errors.New("response status code 400 with error message: validation failure list:\nvalidation failure list:\nvalidation failure list:\nbank_id_code in body should match '^[A-Z]{0,16}$'"),
		},
		{
			name: "bic code is validated",
			givenAccountdata: AccountDataFactory.MustCreateWithOption(
				map[string]interface{}{"Attributes.Bic": "WRONGBIC123213"}).(*account.AccountData),
			expectedErrror: errors.New("response status code 400 with error message: validation failure list:\nvalidation failure list:\nvalidation failure list:\nbic in body should match '^([A-Z]{6}[A-Z0-9]{2}|[A-Z]{6}[A-Z0-9]{5})$'"),
		},
		{
			name: "bank details are not country conditional",
			givenAccountdata: AccountDataFactory.MustCreateWithOption(
				map[string]interface{}{"Attributes.Country": "GB", "Attributes.Bic": "RIGHTBIC", "Attributes.BankID": "SOMEBANKID"}).(*account.AccountData),
			expectedErrror: nil,
		},
		{
			name: "iban is vaidated but not country conditonal", // for Canada APIdocs say it must be empty
			givenAccountdata: AccountDataFactory.MustCreateWithOption(
				map[string]interface{}{"Attributes.Country": "CA", "Attributes.Iban": "SB01AWESOMEIBAN"}).(*account.AccountData),
			expectedErrror: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			resp, err := client.CreateAccount(tc.givenAccountdata)

			// THEN
			assert.Equal(t, tc.expectedErrror, err, "submitted data: %s", tc.givenAccountdata)
			if err == nil {
				assert.Equal(t, tc.givenAccountdata, resp)
			}
		})
	}
}

func TestCreateAccountWithExistingIDFails(t *testing.T) {
	client := account.NewAccountClient(fetchAPIHostName(), account.ClientTimeout)
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

func TestCanFetch(t *testing.T) {
	ac := account.NewAccountClient(fetchAPIHostName(), account.ClientTimeout)

	t.Run("can fetch account data", func(t *testing.T) {
		// WHEN
		data := AccountDataFactory.MustCreate().(*account.AccountData)
		ac.CreateAccount(data)

		// THEN
		fetchedData, err := ac.GetById(data.ID)
		assert.Nil(t, err)
		assert.Equal(t, data, fetchedData)
	})

	t.Run("fetch invalid account", func(t *testing.T) {
		// WHEN
		fetchedData, err := ac.GetById("non existing id")

		// THEN
		assert.Nil(t, fetchedData)
		assert.NotNil(t, err)
		assert.Equal(t, "response status code 400 with error message: id is not a valid uuid", err.Error())

	})

	t.Run("fetch non existing account", func(t *testing.T) {
		// WHEN
		notInsertedAccount := AccountDataFactory.MustCreate().(*account.AccountData) // only need the generated ID
		fetchedData, err := ac.GetById(notInsertedAccount.ID)

		// THEN
		assert.Nil(t, fetchedData)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Sprintf("response status code 404 with error message: record %s does not exist", notInsertedAccount.ID), err.Error())

	})
}

func TestDelete(t *testing.T) {
	ac := account.NewAccountClient(fetchAPIHostName(), account.ClientTimeout)

	t.Run("can delete successfully", func(t *testing.T) {
		// WHEN
		data := AccountDataFactory.MustCreate().(*account.AccountData)
		ac.CreateAccount(data)

		// THEN
		err := ac.DeleteAccount(data.ID, data.Version)
		assert.NoError(t, err)
	})

	t.Run("can delete successfully", func(t *testing.T) {
		// WHEN
		data := AccountDataFactory.MustCreate().(*account.AccountData)
		ac.CreateAccount(data)

		// THEN
		err := ac.DeleteAccount(data.ID, data.Version)
		assert.NoError(t, err)
	})
}

func TestPatch(t *testing.T) {
	ac := account.NewAccountClient(fetchAPIHostName(), account.ClientTimeout)

	t.Run("cannot patch data; likely not implemented on the server side", func(t *testing.T) {
		// WHEN
		data := AccountDataFactory.MustCreate().(*account.AccountData)
		result, err := ac.CreateAccount(data)
		assert.NoError(t, err)
		result.Attributes.Country = "ZB"
		result.Version++
		_, err = ac.UpdateAccount(data)

		// THEN
		assert.Equal(t, err.Error(), "response status code 404 with error message: ")

	})
}
