package choretemplates

import (
	"context"
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create", func() {
	var (
		persister  *MockPersister
		normalizer *MockNormalizer
		service    *Service
	)

	BeforeEach(func() {
		persister = NewMockPersister(GinkgoT())
		normalizer = NewMockNormalizer(GinkgoT())
		service = NewService(&Deps{
			Persister:  persister,
			Normalizer: normalizer,
		}, &Config{})
	})

	It("normalizes and persists the chore template", func(ctx SpecContext) {
		params := &models.CreateChoreTemplateParams{
			Name:        " raw name ",
			Description: " raw description ",
		}
		persisted := &models.ChoreTemplate{
			Id:          1,
			Name:        "Normalized name",
			Description: "Normalized description",
		}
		normalizer.EXPECT().
			NormalizeName(" raw name ").
			Return("Normalized name", nil).
			Once()
		normalizer.EXPECT().
			NormalizeDescription(" raw description ").
			Return("Normalized description", nil).
			Once()
		persister.EXPECT().
			Create(ctx, &models.CreateChoreTemplateParams{
				Name:        "Normalized name",
				Description: "Normalized description",
			}).
			Return(persisted, nil).
			Once()

		result, err := service.Create(ctx, params)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(persisted))
		Expect(params).To(Equal(&models.CreateChoreTemplateParams{
			Name:        " raw name ",
			Description: " raw description ",
		}))
	})

	It("rejects an invalid name", func() {
		normalizer.EXPECT().
			NormalizeName("invalid name").
			Return("", validationerrors.ErrRequired).
			Once()

		result, err := service.Create(
			context.Background(),
			&models.CreateChoreTemplateParams{
				Name: "invalid name",
			},
		)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidName))
	})

	It("rejects an invalid description", func() {
		normalizer.EXPECT().
			NormalizeName("name").
			Return("Normalized name", nil).
			Once()
		normalizer.EXPECT().
			NormalizeDescription("invalid description").
			Return("", validationerrors.ErrTooLong).
			Once()

		result, err := service.Create(
			context.Background(),
			&models.CreateChoreTemplateParams{
				Name:        "name",
				Description: "invalid description",
			},
		)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidDescription))
	})

	It("rejects missing parameters", func() {
		result, err := service.Create(context.Background(), nil)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidName))
	})

	It("preserves persistence errors", func(ctx SpecContext) {
		persistErr := errors.New("persistence failed")
		normalizer.EXPECT().NormalizeName("name").Return("Name", nil).Once()
		normalizer.EXPECT().
			NormalizeDescription("description").
			Return("Description", nil).
			Once()
		persister.EXPECT().
			Create(ctx, &models.CreateChoreTemplateParams{
				Name:        "Name",
				Description: "Description",
			}).
			Return(nil, persistErr).
			Once()

		result, err := service.Create(ctx, &models.CreateChoreTemplateParams{
			Name:        "name",
			Description: "description",
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("create chore template: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})

	It("preserves the existing template for a name conflict", func(ctx SpecContext) {
		conflict := &models.NameConflictError{ExistingId: 7}
		normalizer.EXPECT().NormalizeName("name").Return("Name", nil).Once()
		normalizer.EXPECT().
			NormalizeDescription("description").
			Return("Description", nil).
			Once()
		persister.EXPECT().
			Create(ctx, &models.CreateChoreTemplateParams{
				Name:        "Name",
				Description: "Description",
			}).
			Return(nil, conflict).
			Once()

		result, err := service.Create(ctx, &models.CreateChoreTemplateParams{
			Name:        "name",
			Description: "description",
		})

		Expect(result).To(BeNil())
		var actual *models.NameConflictError
		Expect(errors.As(err, &actual)).To(BeTrue())
		Expect(actual).To(BeIdenticalTo(conflict))
	})
})
