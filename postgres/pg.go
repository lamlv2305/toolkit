package postgres

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

func Now() pgtype.Timestamp {
	return Timestamp(time.Now())
}

func Timestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:             t,
		InfinityModifier: 0,
		Valid:            true,
	}
}
