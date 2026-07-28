package choretemplates

import (
	"context"
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Edit", func() {
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
		params := &models.EditChoreTemplateParams{
			Id:          42,
			Name:        " raw name ",
			Description: " raw description ",
		}
		persisted := &models.ChoreTemplate{
			Id:          42,
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
			Edit(ctx, &models.EditChoreTemplateParams{
				Id:          42,
				Name:        "Normalized name",
				Description: "Normalized description",
			}).
			Return(persisted, nil).
			Once()

		result, err := service.Edit(ctx, params)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(persisted))
		Expect(params).To(Equal(&models.EditChoreTemplateParams{
			Id:          42,
			Name:        " raw name ",
			Description: " raw description ",
		}))
	})

	It("rejects an invalid name", func() {
		normalizer.EXPECT().
			NormalizeName("invalid name").
			Return("", validationerrors.ErrRequired).
			Once()

		result, err := service.Edit(context.Background(), &models.EditChoreTemplateParams{
			Id:   42,
			Name: "invalid name",
		})

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

		result, err := service.Edit(context.Background(), &models.EditChoreTemplateParams{
			Id:          42,
			Name:        "name",
			Description: "invalid description",
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidDescription))
	})

	It("rejects missing parameters", func() {
		result, err := service.Edit(context.Background(), nil)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidName))
	})

	It("preserves persistence errors", func(ctx SpecContext) {
		persistErr := errors.New("persistence failed")
		normalizer.EXPECT().NormalizeName("name").Return("Name", nil).Once()
		normalizer.EXPECT().NormalizeDescription("description").Return("Description", nil).Once()
		persister.EXPECT().
			Edit(ctx, &models.EditChoreTemplateParams{
				Id:          42,
				Name:        "Name",
				Description: "Description",
			}).
			Return(nil, persistErr).
			Once()

		result, err := service.Edit(ctx, &models.EditChoreTemplateParams{
			Id:          42,
			Name:        "name",
			Description: "description",
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("edit chore template: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})
})
