package chores

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
)

func (p *Persister) Browse(
	ctx context.Context,
	params *choremodels.BrowseChoresParams,
) (_ *choremodels.ChorePage, returnErr error) {
	query := `
		select id, name, description, is_complete, completed_on, deadline
		from chores
		where is_complete = ?
			and (
				? = ''
				or instr(lower(name), lower(?)) > 0
				or instr(lower(coalesce(description, '')), lower(?)) > 0
			)
	`
	isComplete := params.Status == choremodels.ChoreStatusCompleted
	if isComplete {
		query += "order by completed_on desc, id desc\n"
	} else {
		query += "order by deadline asc, id asc\n"
	}
	query += "limit ? offset ?"

	rows, err := p.db.QueryContext(
		ctx,
		query,
		isComplete,
		params.Search,
		params.Search,
		params.Search,
		params.Limit+1,
		params.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("browse chores: %w", err)
	}
	defer func() {
		returnErr = errors.Join(returnErr, rows.Close())
	}()

	chores := make([]choremodels.Chore, 0, params.Limit+1)
	for rows.Next() {
		var (
			chore       choremodels.Chore
			description sql.NullString
			completed   bool
			completedOn sql.NullTime
			deadline    time.Time
		)
		if err := rows.Scan(
			&chore.Id,
			&chore.Name,
			&description,
			&completed,
			&completedOn,
			&deadline,
		); err != nil {
			return nil, fmt.Errorf("scan chore: %w", err)
		}
		chore.Status = choremodels.ChoreStatusActive
		if completed {
			chore.Status = choremodels.ChoreStatusCompleted
		}
		chore.Description = description.String
		chore.Deadline = deadline
		if completedOn.Valid {
			chore.CompletedOn = completedOn.Time
		}
		chores = append(chores, chore)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chores: %w", err)
	}

	more := len(chores) > params.Limit
	if more {
		chores = chores[:params.Limit]
	}
	return &choremodels.ChorePage{Chores: chores, More: more}, nil
}
