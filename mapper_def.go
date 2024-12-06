package main

import (
	"strings"

	"github.com/udisondev/go-mapp/domain"
	d "github.com/udisondev/go-mapp/dto"
)

// TODO добавить обработку slice
// TODO добавить обработку map
// TODO добавить обработку enum (with default)
// TODO work with err

// TODO добавить source path
// TODO expr
type Mapper interface {

	//@qual -s=Firstname -t=.FirstName
	//@qual -t=.LastName -mn=lastNameMapper
	//@qual -s=Number -t=.Profile.Phone
	//@ignore -t=.Email
	MapPersonToDTO(p domain.Person) d.Person

	//@emapper
	//@ignore -s=Vip -t=Crazy
	//@ignorecase
	MapPersonTypeToDto(pt domain.PersonType) (d.PersonType, error)

	//@emapper
	//@ignorecase
	//@ignore -s=Crazy -t=Vip
	//@errf ("I dont want to handle this: %v", pt)
	//@def Vip
	MapPersonTypeToDomain(pt d.PersonType) (domain.PersonType, error)

	//@qual -s=FirstName -t=.Firstname
	//@qual -s=Phone -t=.Profile.Number
	//@qual -t=.Firstname -s=FirstName
	//@ignore -t=.MainAccount
	MapPersonToDomain(p d.Person) (domain.Person, error)
}

func lastNameMapper(lastName string) string {
	return strings.ToLower(lastName)
}
