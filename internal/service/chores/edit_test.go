package chores

import (
	"context"
	"errors"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
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

	It("normalizes and persists an active one-off chore", func(ctx SpecContext) {
		deadline := time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC)
		input := &models.EditChoreParams{
			Id:                      42,
			Name:                    " raw name ",
			Description:             " raw description ",
			Deadline:                deadline,
			ScheduleId:              nil,
			AlsoUpdateChoreTemplate: false,
		}
		persisted := &models.ChoreDetails{
			Id:          42,
			Name:        "Normalized name",
			Description: "Normalized description",
			Deadline:    deadline,
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
			EditOneOff(ctx, &models.EditOneOffChoreParams{
				Id:          42,
				Name:        "Normalized name",
				Description: "Normalized description",
				Deadline:    deadline,
			}).
			Return(persisted, nil).
			Once()

		result, err := service.Edit(ctx, input)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(persisted))
		Expect(input).To(Equal(&models.EditChoreParams{
			Id:                      42,
			Name:                    " raw name ",
			Description:             " raw description ",
			Deadline:                deadline,
			ScheduleId:              nil,
			AlsoUpdateChoreTemplate: false,
		}))
	})

	It(
		"normalizes and persists an active scheduled chore without changing its deadline",
		func(ctx SpecContext) {
			scheduleId := 7
			deadline := time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC)
			input := &models.EditChoreParams{
				Id:                      42,
				Name:                    " raw name ",
				Description:             " raw description ",
				Deadline:                deadline,
				ScheduleId:              &scheduleId,
				AlsoUpdateChoreTemplate: true,
			}
			persisted := &models.ChoreDetails{
				Id:          42,
				ScheduleId:  &scheduleId,
				Name:        "Normalized name",
				Description: "Normalized description",
				Deadline:    deadline,
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
				EditScheduled(ctx, &models.EditScheduledChoreParams{
					Id:                      42,
					Name:                    "Normalized name",
					Description:             "Normalized description",
					AlsoUpdateChoreTemplate: true,
				}).
				Return(persisted, nil).
				Once()

			result, err := service.Edit(ctx, input)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeIdenticalTo(persisted))
			Expect(input).To(Equal(&models.EditChoreParams{
				Id:                      42,
				Name:                    " raw name ",
				Description:             " raw description ",
				Deadline:                deadline,
				ScheduleId:              &scheduleId,
				AlsoUpdateChoreTemplate: true,
			}))
		},
	)

	It("rejects an invalid name without persisting", func() {
		normalizer.EXPECT().
			NormalizeName("invalid name").
			Return("", validationerrors.ErrRequired).
			Once()

		result, err := service.Edit(context.Background(), &models.EditChoreParams{
			Name:     "invalid name",
			Deadline: time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC),
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidName))
	})

	It("rejects an invalid description without persisting", func() {
		normalizer.EXPECT().NormalizeName("name").Return("Name", nil).Once()
		normalizer.EXPECT().
			NormalizeDescription("invalid description").
			Return("", validationerrors.ErrTooLong).
			Once()

		result, err := service.Edit(context.Background(), &models.EditChoreParams{
			Name:        "name",
			Description: "invalid description",
			Deadline:    time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC),
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidDescription))
	})

	It("rejects a one-off chore without a deadline before persisting", func() {
		normalizer.EXPECT().NormalizeName("name").Return("Name", nil).Once()
		normalizer.EXPECT().
			NormalizeDescription("description").
			Return("Description", nil).
			Once()

		result, err := service.Edit(context.Background(), &models.EditChoreParams{
			Name:        "name",
			Description: "description",
			Deadline:    time.Time{},
			ScheduleId:  nil,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validationerrors.ErrInvalidDeadline))
	})

	It("preserves one-off persistence errors", func(ctx SpecContext) {
		deadline := time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC)
		persistErr := errors.New("persistence failed")
		normalizer.EXPECT().NormalizeName("name").Return("Name", nil).Once()
		normalizer.EXPECT().
			NormalizeDescription("description").
			Return("Description", nil).
			Once()
		persister.EXPECT().
			EditOneOff(ctx, &models.EditOneOffChoreParams{
				Id:          42,
				Name:        "Name",
				Description: "Description",
				Deadline:    deadline,
			}).
			Return(nil, persistErr).
			Once()

		result, err := service.Edit(ctx, &models.EditChoreParams{
			Id:          42,
			Name:        "name",
			Description: "description",
			Deadline:    deadline,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("edit one-off chore: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})

	It("preserves scheduled persistence errors", func(ctx SpecContext) {
		scheduleId := 7
		persistErr := errors.New("persistence failed")
		normalizer.EXPECT().NormalizeName("name").Return("Name", nil).Once()
		normalizer.EXPECT().
			NormalizeDescription("description").
			Return("Description", nil).
			Once()
		persister.EXPECT().
			EditScheduled(ctx, &models.EditScheduledChoreParams{
				Id:          42,
				Name:        "Name",
				Description: "Description",
			}).
			Return(nil, persistErr).
			Once()

		result, err := service.Edit(ctx, &models.EditChoreParams{
			Id:          42,
			ScheduleId:  &scheduleId,
			Name:        "name",
			Description: "description",
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("edit scheduled chore: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})
})
