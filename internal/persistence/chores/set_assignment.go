package chores

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/jmoiron/sqlx"
)

func (p *Persister) SetAssignment(
	ctx context.Context,
	params *models.SetAssignmentParams,
) (*models.ChoreDetails, error) {
	return executeInTransaction(
		ctx,
		p.db,
		"chore assignment",
		func(tx *sqlx.Tx) (*models.ChoreDetails, error) {
			return setChoreAssignment(ctx, tx, params)
		},
	)
}

func setChoreAssignment(
	ctx context.Context,
	tx *sqlx.Tx,
	params *models.SetAssignmentParams,
) (*models.ChoreDetails, error) {
	chore, err := getChore(ctx, tx, params.Id)
	if err != nil {
		return nil, err
	}
	var ownerName string
	err = tx.QueryRowContext(ctx, `select name from profiles where id = ?`, params.OwnerProfileId).
		Scan(&ownerName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrInvalidOwnerProfile
	}
	if err != nil {
		return nil, fmt.Errorf("find owner profile: %w", err)
	}
	if _, err := tx.ExecContext(
		ctx,
		`update chores set owner_profile_id = ? where id = ?`,
		params.OwnerProfileId,
		params.Id,
	); err != nil {
		return nil, fmt.Errorf("set chore assignment: %w", err)
	}
	chore.OwnerProfileId = &params.OwnerProfileId
	chore.OwnerProfileName = ownerName
	return chore, nil
}
