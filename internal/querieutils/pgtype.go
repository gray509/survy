package querieutils

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func Text(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{
			String: "",
			Valid:  false,
		}
	}
	return pgtype.Text{
		String: *v,
		Valid:  true,
	}
}

func Time(v *time.Time) pgtype.Timestamptz {
	if v == nil {
		return pgtype.Timestamptz{
			Time:             time.Time{},
			InfinityModifier: 0,
			Valid:            false,
		}
	}
	return pgtype.Timestamptz{
		Time:             *v,
		InfinityModifier: 0,
		Valid:            true,
	}
}
