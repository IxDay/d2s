package types

import "encoding/gob"

type User struct {
	Name  string
	Email string
	Token string
}

func NewUser(name, email, token string) *User {
	return &User{Name: name, Email: email, Token: token}
}

func init() {
	gob.Register(&User{})
}
