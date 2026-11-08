package profiles

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	validation "github.com/bpsoos/shiftbell/internal/models/validation"
)

type Persister interface {
	Create(context.Context, *models.CreateParams) (*models.Profile, error)
	List(context.Context) ([]models.Profile, error)
	HasProfiles(context.Context) (bool, error)
	HasMultipleProfiles(context.Context) (bool, error)
	CreateInitial(context.Context, *models.CreateParams) (*models.Profile, error)
}

type Normalizer interface {
	NormalizeName(string) (string, error)
}

type Deps struct {
	Persister  Persister
	Normalizer Normalizer
}

type Service struct {
	persister  Persister
	normalizer Normalizer
}

func NewService(deps *Deps) *Service {
	return &Service{persister: deps.Persister, normalizer: deps.Normalizer}
}

func (s *Service) normalize(params *models.CreateParams) (*models.CreateParams, error) {
	if params == nil {
		return nil, fmt.Errorf(
			"%w: %w",
			validation.ErrInvalidName,
			validation.ErrRequired,
		)
	}
	name, err := s.normalizer.NormalizeName(params.Name)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", validation.ErrInvalidName, err)
	}
	if strings.ContainsFunc(name, unicode.IsSpace) {
		return nil, fmt.Errorf(
			"%w: %w",
			validation.ErrInvalidName,
			models.ErrNameContainsSpaces,
		)
	}
	return &models.CreateParams{Name: name}, nil
}
