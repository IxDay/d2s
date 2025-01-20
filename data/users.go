package data

import "github.com/bokwoon95/sq"

type USERS struct {
	sq.TableStruct
	ID       sq.NumberField  `ddl:"primarykey"`
	EMAIL    sq.StringField  `ddl:"notnull default=''"`
	NAME     sq.StringField  `ddl:"notnull default=''"`
	PASSWORD sq.StringField  `ddl:"notnull default=''"`
	VERIFIED sq.BooleanField `ddl:"notnull default=FALSE"`
	CREATED  sq.TimeField
}
