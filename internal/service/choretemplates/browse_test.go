package choretemplates

import (
	"context"
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Browse", func() {
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

	It("normalizes search and defaults to active chore templates", func(ctx SpecContext) {
		params := &models.BrowseChoreTemplatesParams{
			Search: " raw search ",
			Offset: 5,
			Limit:  10,
		}
		page := &models.ChoreTemplatePage{
			ChoreTemplates: []models.ChoreTemplate{{Id: 1, Name: "Laundry"}},
			More:           true,
		}
		normalizer.EXPECT().
			NormalizeSearch(" raw search ").
			Return("Normalized search", nil).
			Once()
		persister.EXPECT().
			Browse(ctx, &models.BrowseChoreTemplatesParams{
				Filter: models.ChoreTemplateFilterActive,
				Search: "Normalized search",
				Offset: 5,
				Limit:  10,
			}).
			Return(page, nil).
			Once()

		result, err := service.Browse(ctx, params)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(page))
		Expect(params).To(Equal(&models.BrowseChoreTemplatesParams{
			Search: " raw search ",
			Offset: 5,
			Limit:  10,
		}))
	})

	It("browses deactivated chore templates", func(ctx SpecContext) {
		page := &models.ChoreTemplatePage{}
		normalizer.EXPECT().NormalizeSearch("").Return("", nil).Once()
		persister.EXPECT().
			Browse(ctx, &models.BrowseChoreTemplatesParams{
				Filter: models.ChoreTemplateFilterDeactivated,
				Search: "",
				Offset: 0,
				Limit:  10,
			}).
			Return(page, nil).
			Once()

		result, err := service.Browse(ctx, &models.BrowseChoreTemplatesParams{
			Filter: models.ChoreTemplateFilterDeactivated,
			Limit:  10,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(page))
	})

	It("rejects an invalid filter", func() {
		result, err := service.Browse(context.Background(), &models.BrowseChoreTemplatesParams{
			Filter: "all",
			Limit:  10,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidFilter))
	})

	It("rejects an invalid search", func() {
		normalizer.EXPECT().NormalizeSearch("invalid search").Return("", validationerrors.ErrDisallowedCharacter).Once()

		result, err := service.Browse(context.Background(), &models.BrowseChoreTemplatesParams{
			Search: "invalid search",
			Limit:  10,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidSearch))
	})

	It("rejects a negative offset", func() {
		result, err := service.Browse(context.Background(), &models.BrowseChoreTemplatesParams{
			Offset: -1,
			Limit:  10,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidOffset))
	})

	It("rejects a non-positive limit", func() {
		result, err := service.Browse(context.Background(), &models.BrowseChoreTemplatesParams{})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidLimit))
	})

	It("rejects missing parameters", func() {
		result, err := service.Browse(context.Background(), nil)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidLimit))
	})

	It("preserves persistence errors", func(ctx SpecContext) {
		persistErr := errors.New("persistence failed")
		normalizer.EXPECT().NormalizeSearch("search").Return("Search", nil).Once()
		persister.EXPECT().
			Browse(ctx, &models.BrowseChoreTemplatesParams{
				Filter: models.ChoreTemplateFilterActive,
				Search: "Search",
				Limit:  10,
			}).
			Return(nil, persistErr).
			Once()

		result, err := service.Browse(ctx, &models.BrowseChoreTemplatesParams{
			Search: "search",
			Limit:  10,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("browse chore templates: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})
})
