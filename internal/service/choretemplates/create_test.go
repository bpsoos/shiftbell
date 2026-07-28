package choretemplates

import (
	"context"
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	serviceerrors "github.com/bpsoos/shiftbell/internal/service"
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
			Return("Normalized name", true).
			Once()
		normalizer.EXPECT().
			NormalizeDescription(" raw description ").
			Return("Normalized description", true).
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
			Return("", false).
			Once()

		result, err := service.Create(context.Background(), &models.CreateChoreTemplateParams{
			Name: "invalid name",
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(serviceerrors.ErrInvalidName))
	})

	It("rejects an invalid description", func() {
		normalizer.EXPECT().
			NormalizeName("name").
			Return("Normalized name", true).
			Once()
		normalizer.EXPECT().
			NormalizeDescription("invalid description").
			Return("", false).
			Once()

		result, err := service.Create(context.Background(), &models.CreateChoreTemplateParams{
			Name:        "name",
			Description: "invalid description",
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(serviceerrors.ErrInvalidDescription))
	})

	It("rejects missing parameters", func() {
		result, err := service.Create(context.Background(), nil)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(serviceerrors.ErrInvalidName))
	})

	It("preserves persistence errors", func(ctx SpecContext) {
		persistErr := errors.New("persistence failed")
		normalizer.EXPECT().NormalizeName("name").Return("Name", true).Once()
		normalizer.EXPECT().NormalizeDescription("description").Return("Description", true).Once()
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
		conflict := &NameConflictError{ExistingId: 7}
		normalizer.EXPECT().NormalizeName("name").Return("Name", true).Once()
		normalizer.EXPECT().NormalizeDescription("description").Return("Description", true).Once()
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
		var actual *NameConflictError
		Expect(errors.As(err, &actual)).To(BeTrue())
		Expect(actual).To(BeIdenticalTo(conflict))
	})
})
