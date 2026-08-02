package schedules

import (
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
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
		service = NewService(
			&Deps{Persister: persister, Normalizer: normalizer},
			&Config{},
		)
	})

	It("normalizes and persists the schedule without replacing its template", func(ctx SpecContext) {
		params := &models.EditScheduleParams{
			Id:           42,
			Name:         " raw name ",
			IntervalDays: 14,
		}
		persisted := &models.Schedule{
			Id:              42,
			Name:            "Normalized name",
			ChoreTemplateId: 7,
			IntervalDays:    14,
		}
		normalizer.EXPECT().NormalizeName(" raw name ").Return("Normalized name", nil).Once()
		persister.EXPECT().Edit(ctx, &models.EditScheduleParams{
			Id:           42,
			Name:         "Normalized name",
			IntervalDays: 14,
		}).Return(persisted, nil).Once()

		result, err := service.Edit(ctx, params)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(persisted))
		Expect(params).To(Equal(&models.EditScheduleParams{
			Id:           42,
			Name:         " raw name ",
			IntervalDays: 14,
		}))
	})

	It("rejects an invalid name", func() {
		normalizer.EXPECT().NormalizeName("invalid name").Return("", validationerrors.ErrRequired).Once()

		result, err := service.Edit(
			GinkgoT().Context(),
			&models.EditScheduleParams{Name: "invalid name", IntervalDays: 14},
		)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidName))
	})

	DescribeTable("rejects an invalid interval",
		func(intervalDays int) {
			normalizer.EXPECT().NormalizeName("name").Return("Name", nil).Once()

			result, err := service.Edit(
				GinkgoT().Context(),
				&models.EditScheduleParams{Name: "name", IntervalDays: intervalDays},
			)

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(validationerrors.ErrInvalidInterval))
		},
		Entry("below the minimum", 0),
		Entry("above the maximum", 3651),
	)

	It("preserves persistence errors", func(ctx SpecContext) {
		persistErr := errors.New("persistence failed")
		normalizer.EXPECT().NormalizeName("name").Return("Name", nil).Once()
		persister.EXPECT().Edit(ctx, &models.EditScheduleParams{
			Id:           42,
			Name:         "Name",
			IntervalDays: 14,
		}).Return(nil, persistErr).Once()

		result, err := service.Edit(ctx, &models.EditScheduleParams{
			Id:           42,
			Name:         "name",
			IntervalDays: 14,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("edit schedule: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})
})
