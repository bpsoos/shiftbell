package chores

import (
	"context"
	"errors"
	"time"

	"github.com/bpsoos/shiftbell/internal/models"
	serviceerrors "github.com/bpsoos/shiftbell/internal/service"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create", func() {
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

	DescribeTable(
		"creates a manual one-off chore with different arguments",
		func(
			input *models.CreateChoreInput,
			expected *models.CreateChoreResult,
			normalizedName string,
			normalizedDescription string,
		) {
			normalizer.EXPECT().
				NormalizeName(input.Name).
				Return(normalizedName, true).
				Once()
			normalizer.EXPECT().
				NormalizeDescription(input.Description).
				Return(normalizedDescription, true).
				Once()
			persister.EXPECT().
				CreateManualOneOff(context.Background(), &models.CreateManualOneOffInput{
					Name:                normalizedName,
					Description:         normalizedDescription,
					Deadline:            input.Deadline,
					SaveAsChoreTemplate: input.SaveAsChoreTemplate,
				}).
				Return(expected, nil).
				Once()

			result, err := service.Create(context.Background(), input)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeIdenticalTo(expected))
		},
		Entry(
			"with normalized whitespace and empty template flag",
			&models.CreateChoreInput{
				Name:                " raw name ",
				Description:         " raw description ",
				Deadline:            time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
				ChoreTemplateId:     nil,
				ScheduleName:        "",
				IntervalDays:        nil,
				SaveAsChoreTemplate: false,
			},
			&models.CreateChoreResult{
				Chore: &models.Chore{
					Id:          1,
					Name:        "Normalized name",
					Description: "Normalized description",
					Deadline:    time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
				},
			},
			"Normalized name",
			"Normalized description",
		),
		Entry(
			"with save as chore template",
			&models.CreateChoreInput{
				Name:                "another chore",
				Description:         "details",
				Deadline:            time.Date(2026, time.July, 28, 0, 0, 0, 0, time.UTC),
				ChoreTemplateId:     nil,
				ScheduleName:        "",
				IntervalDays:        nil,
				SaveAsChoreTemplate: true,
			},
			&models.CreateChoreResult{
				Chore: &models.Chore{
					Id:          2,
					Name:        "another chore",
					Description: "details",
					Deadline:    time.Date(2026, time.July, 28, 0, 0, 0, 0, time.UTC),
				},
			},
			"another chore",
			"details",
		),
	)

	It("rejects an invalid name without persisting", func() {
		normalizer.EXPECT().
			NormalizeName("invalid name").
			Return("", false).
			Once()

		result, err := service.Create(context.Background(), &models.CreateChoreInput{
			Name:                "invalid name",
			Description:         "description",
			Deadline:            time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
			ChoreTemplateId:     nil,
			ScheduleName:        "",
			IntervalDays:        nil,
			SaveAsChoreTemplate: false,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(serviceerrors.ErrInvalidName))
	})

	It("rejects an invalid description without persisting", func() {
		normalizer.EXPECT().
			NormalizeName("name").
			Return("Name", true).
			Once()
		normalizer.EXPECT().
			NormalizeDescription("invalid description").
			Return("", false).
			Once()

		result, err := service.Create(context.Background(), &models.CreateChoreInput{
			Name:                "name",
			Description:         "invalid description",
			Deadline:            time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
			ChoreTemplateId:     nil,
			ScheduleName:        "",
			IntervalDays:        nil,
			SaveAsChoreTemplate: false,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(serviceerrors.ErrInvalidDescription))
	})

	It("preserves manual one-off persistence errors", func(ctx SpecContext) {
		deadline := time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC)
		persistErr := errors.New("persistence failed")
		normalizer.EXPECT().
			NormalizeName("name").
			Return("Name", true).
			Once()
		normalizer.EXPECT().
			NormalizeDescription("description").
			Return("Description", true).
			Once()
		persister.EXPECT().
			CreateManualOneOff(ctx, &models.CreateManualOneOffInput{
				Name:        "Name",
				Description: "Description",
				Deadline:    deadline,
			}).
			Return(nil, persistErr).
			Once()

		result, err := service.Create(ctx, &models.CreateChoreInput{
			Name:                "name",
			Description:         "description",
			Deadline:            deadline,
			ChoreTemplateId:     nil,
			ScheduleName:        "",
			IntervalDays:        nil,
			SaveAsChoreTemplate: false,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("create manual one-off chore: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})

	It("creates a manual scheduled chore atomically", func(ctx SpecContext) {
		deadline := time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC)
		intervalDays := 14
		input := &models.CreateChoreInput{
			Name:                " raw chore name ",
			Description:         " raw description ",
			Deadline:            deadline,
			ChoreTemplateId:     nil,
			ScheduleName:        " raw schedule name ",
			IntervalDays:        &intervalDays,
			SaveAsChoreTemplate: false,
		}
		persisted := &models.CreateChoreResult{
			Chore:         &models.Chore{Id: 1},
			ChoreTemplate: &models.ChoreTemplate{Id: 2},
		}
		normalizer.EXPECT().NormalizeName(" raw chore name ").Return("Chore name", true).Once()
		normalizer.EXPECT().NormalizeDescription(" raw description ").Return("Description", true).Once()
		normalizer.EXPECT().NormalizeName(" raw schedule name ").Return("Schedule name", true).Once()
		persister.EXPECT().
			CreateManualScheduled(ctx, &models.CreateManualScheduledInput{
				Name:         "Chore name",
				Description:  "Description",
				Deadline:     deadline,
				ScheduleName: "Schedule name",
				IntervalDays: 14,
			}).
			Return(persisted, nil).
			Once()

		result, err := service.Create(ctx, input)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(persisted))
		Expect(input).To(Equal(&models.CreateChoreInput{
			Name:                " raw chore name ",
			Description:         " raw description ",
			Deadline:            deadline,
			ChoreTemplateId:     nil,
			ScheduleName:        " raw schedule name ",
			IntervalDays:        &intervalDays,
			SaveAsChoreTemplate: false,
		}))
	})

	It("creates a template-based scheduled chore atomically", func(ctx SpecContext) {
		choreTemplateId := 42
		deadline := time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC)
		intervalDays := 14
		input := &models.CreateChoreInput{
			Name:                "",
			Description:         "",
			Deadline:            deadline,
			ChoreTemplateId:     &choreTemplateId,
			ScheduleName:        " raw schedule name ",
			IntervalDays:        &intervalDays,
			SaveAsChoreTemplate: false,
		}
		persisted := &models.CreateChoreResult{
			Chore:         &models.Chore{Id: 1},
			ChoreTemplate: &models.ChoreTemplate{Id: choreTemplateId},
		}
		normalizer.EXPECT().NormalizeName(" raw schedule name ").Return("Schedule name", true).Once()
		persister.EXPECT().
			CreateTemplateScheduled(ctx, &models.CreateTemplateScheduledInput{
				ChoreTemplateId: choreTemplateId,
				Deadline:        deadline,
				ScheduleName:    "Schedule name",
				IntervalDays:    intervalDays,
			}).
			Return(persisted, nil).
			Once()

		result, err := service.Create(ctx, input)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(persisted))
		Expect(input).To(Equal(&models.CreateChoreInput{
			Name:                "",
			Description:         "",
			Deadline:            deadline,
			ChoreTemplateId:     &choreTemplateId,
			ScheduleName:        " raw schedule name ",
			IntervalDays:        &intervalDays,
			SaveAsChoreTemplate: false,
		}))
	})

	It("rejects an invalid manual schedule name without persisting", func() {
		intervalDays := 14
		normalizer.EXPECT().NormalizeName("chore name").Return("Chore name", true).Once()
		normalizer.EXPECT().NormalizeDescription("description").Return("Description", true).Once()
		normalizer.EXPECT().NormalizeName("invalid schedule name").Return("", false).Once()

		result, err := service.Create(context.Background(), &models.CreateChoreInput{
			Name:                "chore name",
			Description:         "description",
			Deadline:            time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
			ChoreTemplateId:     nil,
			ScheduleName:        "invalid schedule name",
			IntervalDays:        &intervalDays,
			SaveAsChoreTemplate: false,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(serviceerrors.ErrInvalidName))
	})

	It("rejects an invalid template-based schedule name without persisting", func() {
		choreTemplateId := 42
		intervalDays := 14
		normalizer.EXPECT().NormalizeName("invalid schedule name").Return("", false).Once()

		result, err := service.Create(context.Background(), &models.CreateChoreInput{
			Name:                "",
			Description:         "",
			Deadline:            time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
			ChoreTemplateId:     &choreTemplateId,
			ScheduleName:        "invalid schedule name",
			IntervalDays:        &intervalDays,
			SaveAsChoreTemplate: false,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(serviceerrors.ErrInvalidName))
	})

	DescribeTable(
		"rejects an invalid schedule interval without persisting",
		func(intervalDays int) {
			result, err := service.Create(context.Background(), &models.CreateChoreInput{
				Name:                "chore name",
				Description:         "description",
				Deadline:            time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
				ChoreTemplateId:     nil,
				ScheduleName:        "schedule name",
				IntervalDays:        &intervalDays,
				SaveAsChoreTemplate: false,
			})

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(serviceerrors.ErrInvalidInterval))
		},
		Entry("below the minimum", 0),
		Entry("above the maximum", 3651),
	)

	DescribeTable(
		"rejects an invalid template-based schedule interval without persisting",
		func(intervalDays int) {
			choreTemplateId := 42
			result, err := service.Create(context.Background(), &models.CreateChoreInput{
				Name:                "",
				Description:         "",
				Deadline:            time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
				ChoreTemplateId:     &choreTemplateId,
				ScheduleName:        "schedule name",
				IntervalDays:        &intervalDays,
				SaveAsChoreTemplate: false,
			})

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(serviceerrors.ErrInvalidInterval))
		},
		Entry("below the minimum", 0),
		Entry("above the maximum", 3651),
	)

	It("rejects a scheduled chore without a deadline before persisting", func() {
		intervalDays := 14

		result, err := service.Create(context.Background(), &models.CreateChoreInput{
			Name:                "chore name",
			Description:         "description",
			Deadline:            time.Time{},
			ChoreTemplateId:     nil,
			ScheduleName:        "schedule name",
			IntervalDays:        &intervalDays,
			SaveAsChoreTemplate: false,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(serviceerrors.ErrInvalidDeadline))
	})

	It("rejects a template-based scheduled chore without a deadline before persisting", func() {
		choreTemplateId := 42
		intervalDays := 14

		result, err := service.Create(context.Background(), &models.CreateChoreInput{
			Name:                "",
			Description:         "",
			Deadline:            time.Time{},
			ChoreTemplateId:     &choreTemplateId,
			ScheduleName:        "schedule name",
			IntervalDays:        &intervalDays,
			SaveAsChoreTemplate: false,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(serviceerrors.ErrInvalidDeadline))
	})

	It("preserves manual scheduled persistence errors", func(ctx SpecContext) {
		deadline := time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC)
		intervalDays := 14
		persistErr := errors.New("persistence failed")
		normalizer.EXPECT().NormalizeName("chore name").Return("Chore name", true).Once()
		normalizer.EXPECT().NormalizeDescription("description").Return("Description", true).Once()
		normalizer.EXPECT().NormalizeName("schedule name").Return("Schedule name", true).Once()
		persister.EXPECT().
			CreateManualScheduled(ctx, &models.CreateManualScheduledInput{
				Name:         "Chore name",
				Description:  "Description",
				Deadline:     deadline,
				ScheduleName: "Schedule name",
				IntervalDays: intervalDays,
			}).
			Return(nil, persistErr).
			Once()

		result, err := service.Create(ctx, &models.CreateChoreInput{
			Name:                "chore name",
			Description:         "description",
			Deadline:            deadline,
			ChoreTemplateId:     nil,
			ScheduleName:        "schedule name",
			IntervalDays:        &intervalDays,
			SaveAsChoreTemplate: false,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("create manual scheduled chore: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})

	It("preserves template-based scheduled persistence errors", func(ctx SpecContext) {
		choreTemplateId := 42
		deadline := time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC)
		intervalDays := 14
		persistErr := errors.New("persistence failed")
		normalizer.EXPECT().NormalizeName("schedule name").Return("Schedule name", true).Once()
		persister.EXPECT().
			CreateTemplateScheduled(ctx, &models.CreateTemplateScheduledInput{
				ChoreTemplateId: choreTemplateId,
				Deadline:        deadline,
				ScheduleName:    "Schedule name",
				IntervalDays:    intervalDays,
			}).
			Return(nil, persistErr).
			Once()

		result, err := service.Create(ctx, &models.CreateChoreInput{
			Name:                "",
			Description:         "",
			Deadline:            deadline,
			ChoreTemplateId:     &choreTemplateId,
			ScheduleName:        "schedule name",
			IntervalDays:        &intervalDays,
			SaveAsChoreTemplate: false,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("create template scheduled chore: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})

})
