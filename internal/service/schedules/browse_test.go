package schedules

import (
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
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
		service = NewService(
			&Deps{Persister: persister, Normalizer: normalizer},
			&Config{},
		)
	})

	It("normalizes search and defaults to active schedules", func(ctx SpecContext) {
		params := &models.BrowseSchedulesParams{
			Search: " raw search ",
			Offset: 5,
			Limit:  10,
		}
		page := &models.SchedulePage{
			Schedules: []models.Schedule{{Id: 1, Name: "Laundry"}},
			More:      true,
		}
		normalizer.EXPECT().
			NormalizeSearch(" raw search ").
			Return("Normalized search", nil).Once()
		persister.EXPECT().Browse(ctx, &models.BrowseSchedulesParams{
			Filter: models.ScheduleFilterActive,
			Search: "Normalized search",
			Offset: 5,
			Limit:  10,
		}).Return(page, nil).Once()

		result, err := service.Browse(ctx, params)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(page))
		Expect(params).To(Equal(&models.BrowseSchedulesParams{Search: " raw search ", Offset: 5, Limit: 10}))
	})

	It("browses deactivated schedules", func(ctx SpecContext) {
		page := &models.SchedulePage{}
		normalizer.EXPECT().NormalizeSearch("").Return("", nil).Once()
		persister.EXPECT().Browse(ctx, &models.BrowseSchedulesParams{
			Filter: models.ScheduleFilterDeactivated,
			Limit:  10,
		}).Return(page, nil).Once()

		result, err := service.Browse(ctx, &models.BrowseSchedulesParams{
			Filter: models.ScheduleFilterDeactivated,
			Limit:  10,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(page))
	})

	It("rejects an invalid filter", func() {
		result, err := service.Browse(
			GinkgoT().Context(),
			&models.BrowseSchedulesParams{Filter: "all", Limit: 10},
		)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidFilter))
	})

	It("rejects invalid pagination", func() {
		result, err := service.Browse(
			GinkgoT().Context(),
			&models.BrowseSchedulesParams{Offset: -1, Limit: 10},
		)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidOffset))

		result, err = service.Browse(
			GinkgoT().Context(),
			&models.BrowseSchedulesParams{Limit: 0},
		)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidLimit))
	})

	It("rejects an invalid search", func() {
		normalizer.EXPECT().
			NormalizeSearch("invalid search").
			Return("", validationerrors.ErrDisallowedCharacter).
			Once()

		result, err := service.Browse(
			GinkgoT().Context(),
			&models.BrowseSchedulesParams{Search: "invalid search", Limit: 10},
		)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidSearch))
	})

	It("preserves persistence errors", func(ctx SpecContext) {
		persistErr := errors.New("persistence failed")
		normalizer.EXPECT().NormalizeSearch("").Return("", nil).Once()
		persister.EXPECT().Browse(ctx, &models.BrowseSchedulesParams{
			Filter: models.ScheduleFilterActive,
			Limit:  10,
		}).Return(nil, persistErr).Once()

		result, err := service.Browse(ctx, &models.BrowseSchedulesParams{Limit: 10})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("browse schedules: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})
})
