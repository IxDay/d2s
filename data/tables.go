package data

import "github.com/bokwoon95/sq"

type USERS struct {
	sq.TableStruct
	ID      sq.NumberField `ddl:"primarykey"`
	EMAIL   sq.StringField `ddl:"notnull unique default=''"`
	NAME    sq.StringField `ddl:"notnull default=''"`
	CREATED sq.TimeField   `ddl:"type=DATETIME"`
}
