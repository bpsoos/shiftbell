package choretemplates_test

import (
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	schedulemodels "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	servicepackage "github.com/bpsoos/shiftbell/internal/service/choretemplates"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("List template schedules", func() {
	var persister *MockPersister
	var service *servicepackage.Service
	BeforeEach(func() {
		persister = NewMockPersister(GinkgoT())
		service = servicepackage.NewService(
			&servicepackage.Deps{Persister: persister},
			&servicepackage.Config{},
		)
	})
	DescribeTable(
		"returns template context and schedules",
		func(ctx SpecContext, state schedulemodels.ScheduleFilter) {
			persister.EXPECT().
				Get(ctx, 8).
				Return(&models.ChoreTemplateDetails{Id: 8, Name: "Laundry"}, nil).
				Once()
			schedules := []models.ScheduleReference{{Id: 42, Name: "Weekly laundry"}}
			persister.EXPECT().ListSchedules(ctx, 8, state).Return(schedules, nil).Once()
			result, err := service.ListSchedules(ctx, 8, state)
			Expect(err).NotTo(HaveOccurred())
			Expect(
				result,
			).To(Equal(&models.TemplateSchedules{TemplateName: "Laundry", Schedules: schedules}))
		},
		Entry("active", schedulemodels.ScheduleFilterActive),
		Entry("deactivated", schedulemodels.ScheduleFilterDeactivated),
	)
	It("rejects invalid states before persistence", func(ctx SpecContext) {
		result, err := service.ListSchedules(ctx, 8, "unknown")
		Expect(result).To(BeNil())
		Expect(errors.Is(err, validationerrors.ErrInvalidFilter)).To(BeTrue())
	})
	DescribeTable(
		"preserves template load errors without listing schedules",
		func(ctx SpecContext, persistErr error) {
			persister.EXPECT().Get(ctx, 8).Return(nil, persistErr).Once()
			result, err := service.ListSchedules(
				ctx,
				8,
				schedulemodels.ScheduleFilterActive,
			)
			Expect(result).To(BeNil())
			Expect(errors.Is(err, persistErr)).To(BeTrue())
		},
		Entry("missing", models.ErrNotFound),
		Entry("storage", errors.New("storage")),
	)
	It("preserves schedule list errors", func(ctx SpecContext) {
		persistErr := errors.New("schedule query")
		persister.EXPECT().
			Get(ctx, 8).
			Return(&models.ChoreTemplateDetails{Id: 8, Name: "Laundry"}, nil).
			Once()
		persister.EXPECT().
			ListSchedules(ctx, 8, schedulemodels.ScheduleFilterActive).
			Return(nil, persistErr).
			Once()
		result, err := service.ListSchedules(ctx, 8, schedulemodels.ScheduleFilterActive)
		Expect(result).To(BeNil())
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})
})
