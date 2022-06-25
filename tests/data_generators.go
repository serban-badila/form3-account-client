//go:build integration || load
// +build integration load

package tests

import (
	"fmt"
	"math/rand"

	"github.com/bluele/factory-go/factory"
	"github.com/google/uuid"
	"go.form3-client.com/account"
)

var (
	namePrefixes = [...]string{"First name", "Last name", "Third name", "Fourth name"}
	countries    = [...]string{"GB", "RO", "DK", "FR", "CH"}
)

var AccountDataFactory = factory.NewFactory(
	&account.AccountData{Type: "accounts"}).Attr("ID", func(args factory.Args) (interface{}, error) {
	return uuid.New().String(), nil
}).Attr("OrganisationID", func(args factory.Args) (interface{}, error) {
	return uuid.New().String(), nil
}).Attr("Type", func(args factory.Args) (interface{}, error) {
	return "accounts", nil
}).SubFactory("Attributes", AccountAttributesFactory)

var AccountAttributesFactory = factory.NewFactory(&account.AccountAttributes{}).Attr("Name", func(args factory.Args) (interface{}, error) {
	len := rand.Int()%account.MaxNames + 1
	names := make([]string, 0, len)
	for i := 0; i < len; i++ {
		names = append(names, fmt.Sprintf("%s %d", namePrefixes[i], rand.Int()%1000))
	}
	return names, nil
}).Attr("Country", func(args factory.Args) (interface{}, error) {
	index := rand.Int() % len(countries)
	return countries[index], nil
})
