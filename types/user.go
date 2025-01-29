package types

import "encoding/gob"

type User struct {
	Name  string
	Email string
}

func NewUser(name, email string) *User {
	return &User{Name: name, Email: email}
}

func init() {
	gob.Register(&User{})
}
