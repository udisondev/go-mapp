package dto

import "github.com/udisondev/go-mapp/external"

type Person struct {
	FirstName   string
	LastName    string
	MiddleName  *string
	Age         *int
	MainAccount external.Account
	Account     []external.Account
	Profile     []Profile
	Type        PersonType
	Projects    []string
	Email string
}

type PersonType uint8
const (
	Simple PersonType = iota + 1
	Important
	Fun
	Crazy
)

type Profile struct {
	Phone string
}
