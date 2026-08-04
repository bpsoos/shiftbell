package schedules_test

import (
	"errors"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	schedulesservice "github.com/bpsoos/shiftbell/internal/service/schedules"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deactivate", func() {
	var (
		persister     *MockPersister
		service       *schedulesservice.Service
		deactivatedAt time.Time
	)

	BeforeEach(func() {
		deactivatedAt = time.Date(2026, time.July, 29, 12, 30, 0, 0, time.UTC)
		persister = NewMockPersister(GinkgoT())
		service = schedulesservice.NewService(
			&schedulesservice.Deps{
				Persister: persister,
				Now:       func() time.Time { return deactivatedAt },
			},
			&schedulesservice.Config{},
		)
	})

	It(
		"atomically deactivates the schedule and removes its active chore",
		func(ctx SpecContext) {
			persisted := &models.Schedule{
				Id:            42,
				Name:          "Laundry",
				DeactivatedAt: &deactivatedAt,
			}
			persister.EXPECT().Deactivate(ctx, &models.DeactivateScheduleParams{
				Id:            42,
				DeactivatedAt: deactivatedAt,
			}).Return(persisted, nil).Once()

			result, err := service.Deactivate(ctx, 42)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeIdenticalTo(persisted))
		},
	)

	It("preserves persistence errors", func(ctx SpecContext) {
		persistErr := errors.New("persistence failed")
		persister.EXPECT().Deactivate(ctx, &models.DeactivateScheduleParams{
			Id:            42,
			DeactivatedAt: deactivatedAt,
		}).Return(nil, persistErr).Once()

		result, err := service.Deactivate(ctx, 42)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("deactivate schedule: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})
})
