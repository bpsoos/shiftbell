package chores

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/jmoiron/sqlx"
)

type choreTemplateSnapshot struct {
	Name        string
	Description string
}

func (p *Persister) CreateTemplateOneOff(
	ctx context.Context,
	params *choremodels.CreateTemplateOneOffParams,
) (_ *choremodels.CreateChoreResult, returnErr error) {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin template one-off chore transaction: %w", err)
	}
	defer func() {
		returnErr = combineTransactionRollbackError(
			returnErr,
			tx.Rollback(),
			"template one-off chore",
		)
	}()

	snapshot, err := selectActiveTemplateSnapshot(ctx, tx, params.ChoreTemplateId)
	if err != nil {
		return nil, err
	}
	chore, err := insertOneOffChore(ctx, tx, &choremodels.CreateOneOffChoreParams{
		Name:        snapshot.Name,
		Description: snapshot.Description,
		Deadline:    params.Deadline,
	})
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit template one-off chore transaction: %w", err)
	}

	return &choremodels.CreateChoreResult{Chore: chore}, nil
}

func selectActiveTemplateSnapshot(
	ctx context.Context,
	tx *sqlx.Tx,
	choreTemplateId int,
) (*choreTemplateSnapshot, error) {
	var (
		name        string
		description sql.NullString
		deactivated sql.NullTime
	)

	err := tx.QueryRowContext(
		ctx,
		`
			select name, description, deactivated_at
			from chore_templates
			where id = ?
		`,
		choreTemplateId,
	).Scan(
		&name,
		&description,
		&deactivated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, choretemplatemodels.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select chore template snapshot: %w", err)
	}

	if deactivated.Valid {
		return nil, choretemplatemodels.ErrInactive
	}

	return &choreTemplateSnapshot{
		Name:        name,
		Description: description.String,
	}, nil
}
