package populate_test

import (
	"context"
	"errors"
	"testing"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	demomodels "github.com/bpsoos/shiftbell/internal/models/demo"
	"github.com/bpsoos/shiftbell/internal/service/populate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

func TestPopulate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Populate Service Suite")
}

var _ = Describe("Populate", func() {
	var data *MockDataPersister
	var chores *MockChorePersister
	var service *populate.Service
	var ctx context.Context
	var timezone *time.Location
	var now time.Time
	var failure error
	var input *demomodels.Data

	BeforeEach(func() {
		data = NewMockDataPersister(GinkgoT())
		chores = NewMockChorePersister(GinkgoT())
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(context.Background())
		DeferCleanup(cancel)
		timezone = time.FixedZone("demo", 2*60*60)
		now = time.Date(2026, 10, 28, 23, 0, 0, 0, time.UTC)
		failure = errors.New("persistence failed")
		input = &demomodels.Data{
			OneOffChores: []demomodels.OneOffChore{
				{
					Name:                "Custom errand",
					Description:         "Bring a bag",
					DeadlineDays:        3,
					SaveAsChoreTemplate: true,
				},
				{Name: "Finished errand", DeadlineDays: -2, CompletedDays: new(-1)},
			},
			ScheduledChores: []demomodels.ScheduledChore{
				{
					Name:         "Custom routine",
					Description:  "Clean the filter",
					DeadlineDays: 2,
					ScheduleName: "Every ten days",
					IntervalDays: 10,
				},
			},
		}
		service = populate.NewService(&populate.Deps{
			DataPersister:  data,
			ChorePersister: chores,
			Now:            func() time.Time { return now },
		}, &populate.Config{AppTimezone: timezone})
	})

	It("leaves existing data alone", func() {
		data.EXPECT().HasData(ctx).Return(true, nil).Once()
		Expect(service.Populate(ctx, input)).To(Succeed())
		Expect(chores.Calls).To(BeEmpty())
	})

	It("stops when the pre-flight check fails", func() {
		data.EXPECT().HasData(ctx).Return(false, failure).Once()
		Expect(service.Populate(ctx, input)).To(MatchError(failure))
		Expect(chores.Calls).To(BeEmpty())
	})

	It("creates the samples using local dates and the returned chore ID", func() {
		data.EXPECT().HasData(ctx).Return(false, nil).Once()
		var oneOffs []models.CreateManualOneOffParams
		var scheduled []models.CreateManualScheduledParams
		chores.EXPECT().CreateManualOneOff(ctx, mock.Anything).
			Run(func(_ context.Context, params *models.CreateManualOneOffParams) {
				oneOffs = append(oneOffs, *params)
			}).Return(&models.CreateChoreResult{Chore: &models.Chore{Id: 42}}, nil).Times(2)
		today := time.Date(2026, 10, 29, 0, 0, 0, 0, timezone)
		chores.EXPECT().MarkComplete(42, today.AddDate(0, 0, -1)).Return(nil).Once()
		chores.EXPECT().CreateManualScheduled(ctx, mock.Anything).
			Run(func(_ context.Context, params *models.CreateManualScheduledParams) {
				scheduled = append(scheduled, *params)
			}).Return(&models.CreateChoreResult{}, nil).Once()

		Expect(service.Populate(ctx, input)).To(Succeed())
		Expect(oneOffs).To(Equal([]models.CreateManualOneOffParams{
			{
				Name:                "Custom errand",
				Description:         "Bring a bag",
				Deadline:            today.AddDate(0, 0, 3),
				SaveAsChoreTemplate: true,
			},
			{Name: "Finished errand", Deadline: today.AddDate(0, 0, -2)},
		}))
		Expect(scheduled).To(Equal([]models.CreateManualScheduledParams{
			{
				Name:         "Custom routine",
				Description:  "Clean the filter",
				Deadline:     today.AddDate(0, 0, 2),
				ScheduleName: "Every ten days",
				IntervalDays: 10,
			},
		}))
	})

	It("stops on a one-off creation failure", func() {
		data.EXPECT().HasData(ctx).Return(false, nil).Once()
		chores.EXPECT().CreateManualOneOff(ctx, mock.Anything).Return(nil, failure).Once()
		Expect(service.Populate(ctx, input)).To(MatchError(failure))
	})

	It("stops on a completion failure", func() {
		data.EXPECT().HasData(ctx).Return(false, nil).Once()
		chores.EXPECT().CreateManualOneOff(ctx, mock.Anything).
			Return(&models.CreateChoreResult{Chore: &models.Chore{Id: 42}}, nil).Times(2)
		chores.EXPECT().MarkComplete(42, mock.Anything).Return(failure).Once()
		Expect(service.Populate(ctx, input)).To(MatchError(failure))
	})

	It("stops on a scheduled creation failure", func() {
		data.EXPECT().HasData(ctx).Return(false, nil).Once()
		chores.EXPECT().CreateManualOneOff(ctx, mock.Anything).
			Return(&models.CreateChoreResult{Chore: &models.Chore{Id: 42}}, nil).Times(2)
		chores.EXPECT().MarkComplete(42, mock.Anything).Return(nil).Once()
		chores.EXPECT().
			CreateManualScheduled(ctx, mock.Anything).
			Return(nil, failure).
			Once()
		Expect(service.Populate(ctx, input)).To(MatchError(failure))
	})

	It(
		"supports completion today without completing chores with an omitted completion",
		func() {
			input.OneOffChores = []demomodels.OneOffChore{
				{Name: "Done today", CompletedDays: new(0)},
				{Name: "Still pending"},
			}
			input.ScheduledChores = nil
			data.EXPECT().HasData(ctx).Return(false, nil).Once()
			chores.EXPECT().CreateManualOneOff(ctx, mock.Anything).
				Return(&models.CreateChoreResult{Chore: &models.Chore{Id: 42}}, nil).
				Times(2)
			chores.EXPECT().
				MarkComplete(42, time.Date(2026, 10, 29, 0, 0, 0, 0, timezone)).
				Return(nil).
				Once()
			Expect(service.Populate(ctx, input)).To(Succeed())
		},
	)

	It("accepts an empty dataset without creating anything", func() {
		data.EXPECT().HasData(ctx).Return(false, nil).Once()
		Expect(service.Populate(ctx, &demomodels.Data{})).To(Succeed())
		Expect(chores.Calls).To(BeEmpty())
	})
})
