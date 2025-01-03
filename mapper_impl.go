package main

import (
	"fmt"
	domain "github.com/udisondev/go-mapp/domain"
	dto "github.com/udisondev/go-mapp/dto"
	external "github.com/udisondev/go-mapp/external"
	user "github.com/udisondev/go-mapp/user"
)

// Code generated by go-mapp. DO NOT EDIT.
func MapPersonTypeToDto(pt domain.PersonType) (dto.PersonType, error) {
	switch pt {
	case domain.Important:
		return dto.Important, nil
	case domain.FUN:
		return dto.Fun, nil
	case domain.Simple:
		return dto.Simple, nil
	default:
		return 0, fmt.Errorf("unknown source enum: %v", pt)
	}
}

func MapPersonTypeToDomain(pt dto.PersonType) (domain.PersonType, error) {
	switch pt {
	case dto.Simple:
		return domain.Simple, nil
	case dto.Important:
		return domain.Important, nil
	case dto.Fun:
		return domain.FUN, nil
	default:
		return domain.Vip, nil
	}
}

func MapPersonToDTO(src domain.Person) dto.Person {
	target := dto.Person{}
	if src.Firstname != nil {
		target.FirstName = *src.Firstname
	}
	target.LastName = src.LastName
	target.MiddleName = &src.MiddleName
	target.Age = src.Age
	ttMainAccount, mapMainAccountErr := ossyjfdifp(src.MainAccount)
	if mapMainAccountErr != nil {
		panic(fmt.Sprintf("error mapping from 'Account.MainAccount' to 'Account.MainAccount': %v", mapMainAccountErr.Error()))
	}
	target.MainAccount = ttMainAccount
	ttAccountSlice := make([]external.Account, 0, len(src.Account))
	for _, it := range src.Account {
		ttAccount, mapAccountErr := ossyjfdifp(it)
		if mapAccountErr != nil {
			panic(fmt.Sprintf("error mapping from 'Account.Account' to 'Account.Account': %v", mapAccountErr.Error()))
		}
		ttAccountSlice = append(ttAccountSlice, ttAccount)
	}
	target.Account = ttAccountSlice
	if src.Profile != nil {
		ttProfile, mapProfileErr := xqqkibjxyi(*src.Profile)
		if mapProfileErr != nil {
			panic(fmt.Sprintf("error mapping from 'Profile.Profile' to 'Profile.Profile': %v", mapProfileErr.Error()))
		}
		target.Profile = ttProfile
	}
	ttType, mapTypeErr := MapPersonTypeToDto(src.Type)
	if mapTypeErr != nil {
		panic(fmt.Sprintf("error mapping from 'PersonType.Type' to 'PersonType.Type': %v", mapTypeErr.Error()))
	}
	target.Type = ttType
	target.Projects = src.Projects
	return target
}
func ossyjfdifp(src external.Account) (external.Account, error) {
	target := external.Account{}
	ttLogin, mapLoginErr := ftdfptdyph(src.Login)
	if mapLoginErr != nil {
		return external.Account{}, fmt.Errorf("error mapping from 'Login.Login' to 'Login.Login': %w", mapLoginErr)
	}
	target.Login = ttLogin
	target.Password = src.Password
	return target, nil
}
func ftdfptdyph(src user.Login) (user.Login, error) {
	target := user.Login{}
	target.Value = src.Value
	return target, nil
}
func xqqkibjxyi(src domain.Profile) (dto.Profile, error) {
	target := dto.Profile{}
	target.Phone = src.Number
	return target, nil
}
func MapPersonToDomain(src dto.Person) (domain.Person, error) {
	target := domain.Person{}
	target.Firstname = &src.FirstName
	target.LastName = src.LastName
	if src.MiddleName != nil {
		target.MiddleName = *src.MiddleName
	}
	target.Age = src.Age
	ttAccountSlice := make([]external.Account, 0, len(src.Account))
	for _, it := range src.Account {
		ttAccount, mapAccountErr := dfgbeupfeu(it)
		if mapAccountErr != nil {
			return domain.Person{}, fmt.Errorf("error mapping from 'Account.Account' to 'Account.Account': %w", mapAccountErr)
		}
		ttAccountSlice = append(ttAccountSlice, ttAccount)
	}
	target.Account = ttAccountSlice
	ttProfile, mapProfileErr := guclbondhb(src.Profile)
	if mapProfileErr != nil {
		return domain.Person{}, fmt.Errorf("error mapping from 'Profile.Profile' to 'Profile.Profile': %w", mapProfileErr)
	}
	target.Profile = &ttProfile
	ttType, mapTypeErr := MapPersonTypeToDomain(src.Type)
	if mapTypeErr != nil {
		return domain.Person{}, fmt.Errorf("error mapping from 'PersonType.Type' to 'PersonType.Type': %w", mapTypeErr)
	}
	target.Type = ttType
	target.Projects = src.Projects
	return target, nil
}
func dfgbeupfeu(src external.Account) (external.Account, error) {
	target := external.Account{}
	ttLogin, mapLoginErr := fptlapjksh(src.Login)
	if mapLoginErr != nil {
		return external.Account{}, fmt.Errorf("error mapping from 'Login.Login' to 'Login.Login': %w", mapLoginErr)
	}
	target.Login = ttLogin
	target.Password = src.Password
	return target, nil
}
func fptlapjksh(src user.Login) (user.Login, error) {
	target := user.Login{}
	target.Value = src.Value
	return target, nil
}
func guclbondhb(src dto.Profile) (domain.Profile, error) {
	target := domain.Profile{}
	target.Number = src.Phone
	return target, nil
}
