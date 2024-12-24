package main

import (
	"github.com/udisondev/go-mapp/domain"
	d "github.com/udisondev/go-mapp/dto"
)

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
	//@errf ("there is a cutome err. I dont want to handle this: %v", pt)
	MapPersonTypeToDomain(pt d.PersonType) domain.PersonType

	//@qual -s=FirstName -t=.Firstname
	//@qual -s=Phone -t=.Profile.Number
	//@qual -t=.Firstname -s=FirstName
	//@ignore -t=.MainAccount
	MapPersonToDomain(p d.Person) (domain.Person, error)
}
