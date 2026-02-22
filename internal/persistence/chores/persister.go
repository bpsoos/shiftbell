package chores

import (
	"time"

	"github.com/bpsoos/shiftbell/internal/models"
	"github.com/jmoiron/sqlx"
)

type Persister struct {
	db *sqlx.DB
}

type PersisterDeps struct {
	Db *sqlx.DB
}

func NewPersister(deps *PersisterDeps) *Persister {
	return &Persister{
		db: deps.Db,
	}
}

func (p *Persister) GetBatch(offset int, limit int) (*models.GetChoreBatchResult, error) {
	return &models.GetChoreBatchResult{
		Chores: []models.Chore{
			{
				Id:           0,
				IsCompleted:  false,
				Description:  "test",
				IntervalDays: 14,
				Deadline:     time.Now(),
				CreatedAt:    time.Now(),
				CompletedAt:  time.Now(),
			},
			{
				Id:           1,
				IsCompleted:  false,
				Description:  "test 2",
				IntervalDays: 28,
				Deadline:     time.Now(),
				CreatedAt:    time.Now(),
				CompletedAt:  time.Now(),
			},
		},
		More: false,
	}, nil
}

func (p *Persister) PatchStatus(id int, isComplete bool) error {
	return nil
}
