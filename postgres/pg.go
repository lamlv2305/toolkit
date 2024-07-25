package postgres

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

type Postgres struct {
}

func (p Postgres) Now() pgtype.Timestamp {
	return p.Timestamp(time.Now())
}

func (p Postgres) Timestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:             t,
		InfinityModifier: 0,
		Valid:            true,
	}
}
