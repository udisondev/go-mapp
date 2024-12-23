package main

import (
	"fmt"

	dto "github.com/udisondev/go-mapp/dto"
	"github.com/udisondev/go-mapp/cmd/mapper"
	"github.com/udisondev/go-mapp/external"
	"github.com/udisondev/go-mapp/user"
)

func main() {
	p, err := mapper.MapPersonToDomain(dto.Person{
		FirstName:   "Angoatoliy",
		LastName:    "Ivanovich",
		MiddleName:  nil,
		Age:         nil,
		MainAccount: external.Account{
			Login:    user.Login{
				Value: "Loginb",
			},
			Password: "Pass",
		},
		Account:     []external.Account{},
		Profile:     [][]dto.Profile{},
		Type:        dto.Important,
		Projects:    []string{"Jira"},
		Email:       "mail@mail.ro",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("p: %v\n", p)
}