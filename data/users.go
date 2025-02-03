package data

import (
	"errors"
	"time"

	"github.com/bokwoon95/sq"
	"github.com/mattn/go-sqlite3"
	"github.com/platipy-io/d2s/types"
)

var users = sq.New[USERS]("")

func (c *DB) SaveUser(user *types.User) error {
	created := time.Now()
	er := sqlite3.Error{}

	_, err := sq.Exec(c.db, sq.
		InsertInto(users).
		Columns(users.EMAIL, users.NAME, users.CREATED).
		Values(user.Email, user.Name, created).
		SetDialect(sq.DialectSQLite))
	if errors.As(err, &er) && er.ExtendedCode == sqlite3.ErrConstraintUnique {
		return nil
	}
	return err
}
