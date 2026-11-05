package profiles

import (
	"context"
	"strings"
	"unicode"

	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	validation "github.com/bpsoos/shiftbell/internal/models/validation"
)

type Persister interface {
	Create(context.Context, *models.CreateParams) (*models.Profile, error)
	List(context.Context) ([]models.Profile, error)
	HasProfiles(context.Context) (bool, error)
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
		return nil, validation.ErrInvalidName
	}
	name, err := s.normalizer.NormalizeName(params.Name)
	if err != nil || strings.ContainsFunc(name, unicode.IsSpace) {
		return nil, validation.ErrInvalidName
	}
	return &models.CreateParams{Name: name}, nil
}
