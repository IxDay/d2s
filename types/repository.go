package types

import "time"

type Repository struct {
	ID                                 int64
	Owner, Name, Description, Language string
	LastUpdated                        time.Time
}
