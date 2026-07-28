package chores

import (
	"errors"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Complete", func() {
	var (
		persister *MockPersister
		service   *Service
		now       time.Time
	)

	BeforeEach(func() {
		now = time.Date(2026, time.July, 28, 21, 30, 0, 0, time.UTC)
		persister = NewMockPersister(GinkgoT())
		service = NewService(
			&Deps{
				Persister: persister,
				Now:       func() time.Time { return now },
			},
			&Config{AppTimezone: time.FixedZone("UTC+3", 3*60*60)},
		)
	})

	It("completes a one-off chore", func(ctx SpecContext) {
		completedOn := time.Date(2026, time.July, 29, 0, 0, 0, 0, time.UTC)
		input := &models.CompleteChoreParams{Id: 42, CompletedOn: completedOn}
		persisted := &models.CompleteChoreResult{
			Chore: &models.Chore{
				Id:          42,
				Status:      models.ChoreStatusCompleted,
				CompletedOn: completedOn,
			},
		}
		persister.EXPECT().
			Complete(ctx, &models.CompleteChoreParams{
				Id:          42,
				CompletedOn: completedOn,
			}).
			Return(persisted, nil).
			Once()

		result, err := service.Complete(ctx, input)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(persisted))
		Expect(input).To(Equal(&models.CompleteChoreParams{Id: 42, CompletedOn: completedOn}))
	})

	It("completes a scheduled chore and returns its successor", func(ctx SpecContext) {
		scheduleId := 7
		completedOn := time.Date(2026, time.July, 28, 0, 0, 0, 0, time.UTC)
		input := &models.CompleteChoreParams{Id: 42, CompletedOn: completedOn}
		persisted := &models.CompleteChoreResult{
			Chore: &models.Chore{
				Id:         42,
				ScheduleId: &scheduleId,
				Status:     models.ChoreStatusCompleted,
			},
			Successor: &models.Chore{
				Id:         43,
				ScheduleId: &scheduleId,
				Status:     models.ChoreStatusActive,
			},
		}
		persister.EXPECT().
			Complete(ctx, &models.CompleteChoreParams{Id: 42, CompletedOn: completedOn}).
			Return(persisted, nil).
			Once()

		result, err := service.Complete(ctx, input)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(persisted))
		Expect(input).To(Equal(&models.CompleteChoreParams{Id: 42, CompletedOn: completedOn}))
	})

	It("rejects a completion date after the application-local current date", func() {
		result, err := service.Complete(GinkgoT().Context(), &models.CompleteChoreParams{
			Id:          42,
			CompletedOn: time.Date(2026, time.July, 30, 0, 0, 0, 0, time.UTC),
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidCompletionDate))
	})

	It("rejects a missing completion date", func() {
		result, err := service.Complete(GinkgoT().Context(), &models.CompleteChoreParams{Id: 42})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidCompletionDate))
	})

	It("preserves completion errors", func(ctx SpecContext) {
		completedOn := time.Date(2026, time.July, 28, 0, 0, 0, 0, time.UTC)
		persistErr := errors.New("persistence failed")
		persister.EXPECT().
			Complete(ctx, &models.CompleteChoreParams{Id: 42, CompletedOn: completedOn}).
			Return(nil, persistErr).
			Once()

		result, err := service.Complete(ctx, &models.CompleteChoreParams{
			Id:          42,
			CompletedOn: completedOn,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("complete chore: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})
})
