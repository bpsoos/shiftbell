package chores_test

import (
	"errors"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	choresservice "github.com/bpsoos/shiftbell/internal/service/chores"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CorrectCompletion", func() {
	var (
		persister *MockPersister
		service   *choresservice.Service
	)

	BeforeEach(func() {
		persister = NewMockPersister(GinkgoT())
		service = choresservice.NewService(&choresservice.Deps{
			Persister: persister,
			Now: func() time.Time {
				return time.Date(2026, time.July, 29, 10, 0, 0, 0, time.UTC)
			},
		}, &choresservice.Config{AppTimezone: time.UTC})
	})

	It("corrects a one-off chore completion date", func(ctx SpecContext) {
		completedOn := time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC)
		input := &models.CorrectCompletionParams{Id: 42, CompletedOn: completedOn}
		corrected := &models.Chore{
			Id:          42,
			Status:      models.ChoreStatusCompleted,
			CompletedOn: completedOn,
		}
		persisted := &models.CorrectCompletionResult{Chore: corrected}
		persister.EXPECT().
			CorrectCompletion(ctx, &models.CorrectCompletionParams{Id: 42, CompletedOn: completedOn}).
			Return(persisted, nil).
			Once()

		result, err := service.CorrectCompletion(ctx, input)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(corrected))
		Expect(
			input,
		).To(Equal(&models.CorrectCompletionParams{Id: 42, CompletedOn: completedOn}))
	})

	It(
		"atomically corrects a scheduled chore and its active successor",
		func(ctx SpecContext) {
			scheduleId := 7
			completedOn := time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC)
			input := &models.CorrectCompletionParams{Id: 42, CompletedOn: completedOn}
			corrected := &models.Chore{
				Id:          42,
				ScheduleId:  &scheduleId,
				Status:      models.ChoreStatusCompleted,
				CompletedOn: completedOn,
			}
			persisted := &models.CorrectCompletionResult{
				Chore: corrected,
				Successor: &models.Chore{
					Id:         43,
					ScheduleId: &scheduleId,
					Deadline:   time.Date(2026, time.August, 10, 0, 0, 0, 0, time.UTC),
				},
			}
			persister.EXPECT().
				CorrectCompletion(ctx, &models.CorrectCompletionParams{Id: 42, CompletedOn: completedOn}).
				Return(persisted, nil).
				Once()

			result, err := service.CorrectCompletion(ctx, input)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeIdenticalTo(corrected))
			Expect(
				input,
			).To(Equal(&models.CorrectCompletionParams{Id: 42, CompletedOn: completedOn}))
		},
	)

	It("rejects an invalid completion date", func() {
		result, err := service.CorrectCompletion(
			GinkgoT().Context(),
			&models.CorrectCompletionParams{
				Id:          42,
				CompletedOn: time.Date(2026, time.July, 30, 0, 0, 0, 0, time.UTC),
			},
		)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidCompletionDate))
	})

	It("preserves correction errors", func(ctx SpecContext) {
		completedOn := time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC)
		persistErr := errors.New("persistence failed")
		persister.EXPECT().
			CorrectCompletion(ctx, &models.CorrectCompletionParams{Id: 42, CompletedOn: completedOn}).
			Return(nil, persistErr).
			Once()

		result, err := service.CorrectCompletion(ctx, &models.CorrectCompletionParams{
			Id:          42,
			CompletedOn: completedOn,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("correct chore completion: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})
})
