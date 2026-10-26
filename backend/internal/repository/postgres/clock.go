package postgres

import (
	"context"
	"time"

	"github.com/PhamVanPhuc2k2/manage-company/pkg/postgres"
)

type ClockRepository struct {
	db *postgres.DB
}

func NewClockRepository(db *postgres.DB) *ClockRepository {
	return &ClockRepository{
		db: db,
	}
}

func (r *ClockRepository) Now(ctx context.Context) (time.Time, error) {
	var t time.Time
	if err := r.db.QueryRow(ctx, "SELECT NOW()").Scan(&t); err != nil {
		return time.Time{}, err
	}
	return t, nil
}
