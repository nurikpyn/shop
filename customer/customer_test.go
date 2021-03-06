package customer

import (
	"crypto/md5"
	"io"
	"log"
	"testing"

	"github.com/foomo/shop/address"
	"github.com/foomo/shop/test_utils"
	"github.com/foomo/shop/unique"
	"github.com/foomo/shop/utils"
	assert "github.com/stretchr/testify/require"
)

const (
	MOCK_EMAIL            = "Foo@Bar.com"
	MOCK_PASSWORD         = "supersafepassword!11"
	MOCK_EMAIL2           = "Alice@Bar.com"
	MOCK_PASSWORD2        = "evensaferpassword!11!§$%&"
	OPEN_DIFFS_IN_BROWSER = false
)

type (
	FooCustomProvider struct{}
	FooCustomer       struct{}
	FooAddress        struct{}
)

func (foo FooCustomProvider) NewAddressCustom() interface{} {
	return &FooAddress{}
}

func (foo FooCustomProvider) NewCustomerCustom() interface{} {
	return &FooCustomer{}
}

func createNewTestCustomer(email string) (*Customer, error) {
	mailContact := address.CreateMailContact(email)
	mailContact.ExternalID = unique.GetNewIDShortID()

	externalID := unique.GetNewIDShortID()
	addrKey := unique.GetNewIDShortID()

	h := md5.New()
	io.WriteString(h, addrKey)
	addrKeyHash := string(h.Sum(nil))

	return NewCustomer(addrKey, addrKeyHash, externalID, mailContact, FooCustomProvider{})
}

func create2CustomersAndPerformSomeUpserts(t *testing.T) (*Customer, *Customer) {
	test_utils.DropAllCollections()
	customer, err := createNewTestCustomer(MOCK_EMAIL)
	if err != nil {
		t.Fatal(err)
	}
	// Perform 3 Upserts
	customer.Person.FirstName = "Foo"
	err = customer.Upsert()
	customer.Person.MiddleName = "Bob"
	err = customer.Upsert()
	customer.Person.LastName = "Bar"
	err = customer.Upsert()
	address := &address.Address{
		Person: &address.Person{
			Salutation: address.SalutationTypeMr,
			FirstName:  "Foo",
			MiddleName: "Bob",
			LastName:   "Bar",
		},
		Street:       "Holzweg",
		StreetNumber: "5",
		City:         "Bern",
		Country:      "CH",
		ZIP:          "1234",
	}
	err = customer.Upsert()
	// Create a second customer to make the history a little more interesting
	customer2, err := createNewTestCustomer(MOCK_EMAIL2)
	if err != nil {
		t.Fatal(err)
	}
	customer2.Person.FirstName = "Trent"
	customer2.Upsert()
	customer2.Person.MiddleName = "The"
	customer2.Upsert()
	customer2.Person.LastName = "Second"
	customer2, err = customer2.UpsertAndGetCustomer(nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = customer.AddAddress(address)
	if err != nil {
		t.Fatal(err)
	}
	err = customer.Upsert()
	if err != nil {
		t.Fatal(err)
	}
	customer.Person.FirstName = "Alice"
	customer.RemoveAddress(customer.GetAddresses()[0].GetID())
	customer, err = customer.UpsertAndGetCustomer(nil)
	if err != nil {
		t.Fatal(err)
	}
	return customer, customer2
}

func TestCustomerDelete(t *testing.T) {
	test_utils.DropAllCollections()
	customer, err := createNewTestCustomer(MOCK_EMAIL)
	if err != nil {
		t.Fatal(err)
	}

	// test user
	testCustomer, testCustomerErr := GetCustomerById(customer.Id, nil)
	if testCustomerErr != nil {
		t.Fatal(testCustomerErr)
	}

	if testCustomer.BsonId != customer.BsonId {
		t.Fatal("customer missmatch")
	}

	// delete guest
	delErr := customer.Delete()
	if delErr != nil {
		t.Fatal(delErr)
	}
}

func TestCustomerChangeAddress(t *testing.T) {
	test_utils.DropAllCollections()
	customer, err := createNewTestCustomer(MOCK_EMAIL)
	if err != nil {
		t.Fatal(err)
	}

	addr := &address.Address{
		Person: &address.Person{
			Salutation: address.SalutationTypeMr,
			FirstName:  "Foo",
			MiddleName: "Bob",
			LastName:   "Bar",
		},
		Street:       "Holzweg",
		StreetNumber: "5",
		City:         "Bern",
		Country:      "Schweiz",
		CountryCode:  "CH",
		ZIP:          "1234",
	}
	log.Println("Original Address:")
	utils.PrintJSON(addr)
	id, err := customer.AddAddress(addr)
	log.Println("Added Address with id ", id)
	if err != nil {
		t.Fatal(err)
	}
	addressNew := &address.Address{
		Id: id, // Set id of address we want to replace
		Person: &address.Person{
			Salutation: address.SalutationTypeMr,
			FirstName:  "FooChanged",
			MiddleName: "Bob",
			LastName:   "Bar",
		},
		Street:       "Steinweg",
		StreetNumber: "5",
		City:         "Bern",
		Country:      "Schweiz",
		CountryCode:  "CH",
		ZIP:          "1234",
	}
	err = customer.ChangeAddress(addressNew)
	if err != nil {
		t.Fatal(err)
	}

	changedAddress, err := customer.GetAddressById(id)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Changed Address:")
	utils.PrintJSON(changedAddress)

	if changedAddress.Street != "Steinweg" {
		t.Fatal("Expected customer.Person.FirstName == \"FooChanged\" but got " + changedAddress.Street)
	}
	if changedAddress.Person.FirstName != "FooChanged" {
		t.Fatal("Expected changedAddress.Person.FirstName == \"FooChanged\" but got " + changedAddress.Person.FirstName)
	}
}

func TestAddressComplete(t *testing.T) {
	type args struct {
		address *address.Address
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "all fields empty",
			args: args{
				address: &address.Address{},
			},
			wantErr: "7 errors occurred:\n\t* person is nil\n\t* address street is empty\n\t* address street number is empty\n\t* address zip is not valid\n\t* address city is empty\n\t* address country is empty\n\t* address country code is empty\n\n",
		},
		{
			name: "all person fields empty",
			args: args{
				address: &address.Address{
					Person:       &address.Person{},
					Street:       "x",
					StreetNumber: "x",
					ZIP:          "xxxx",
					City:         "x",
					Country:      "x",
					CountryCode:  "x",
				},
			},
			wantErr: "3 errors occurred:\n\t* person salutation is empty\n\t* person firstname is empty\n\t* person lastname is empty\n\n",
		},
		{
			name: "all fields set",
			args: args{
				address: &address.Address{
					Person: &address.Person{
						FirstName:  "x",
						LastName:   "x",
						Salutation: "x",
					},
					Street:       "x",
					StreetNumber: "x",
					ZIP:          "xxxx",
					City:         "x",
					Country:      "x",
					CountryCode:  "x",
				},
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.address.IsComplete()
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewCustomer(t *testing.T) {
	type args struct {
		addrkey        string
		addrkeyHash    string
		externalID     string
		mailContact    *address.Contact
		customProvider CustomerCustomProvider
	}
	tests := []struct {
		name    string
		args    args
		want    *Customer
		wantErr string
	}{
		{
			name: "all fields empty",
			args: args{
				addrkey:        "",
				addrkeyHash:    "",
				externalID:     "",
				mailContact:    nil,
				customProvider: nil,
			},
			want:    nil,
			wantErr: "4 errors occurred:\n\t* required addrkey is empty\n\t* required addrkeyHash is empty\n\t* required externalID is empty\n\t* custom provider not set\n\n",
		},
		{
			name: "all fields empty except mail contact",
			args: args{
				addrkey:     "",
				addrkeyHash: "",
				externalID:  "",
				mailContact: &address.Contact{
					ID:         "",
					ExternalID: "",
					Type:       "",
					Value:      "",
				},
				customProvider: nil,
			},
			want:    nil,
			wantErr: "6 errors occurred:\n\t* required addrkey is empty\n\t* required addrkeyHash is empty\n\t* required externalID is empty\n\t* required email address in mailContact.Value is empty\n\t* required mailContact must have string type \"email\"\n\t* custom provider not set\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCustomer(tt.args.addrkey, tt.args.addrkeyHash, tt.args.externalID, tt.args.mailContact, tt.args.customProvider)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Exactly(t, tt.want, got)
		})
	}
}
