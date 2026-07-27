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

	DescribeTable("creates a manual one-off chore with different arguments", func(input *models.CreateChoreInput, expected *models.CreateChoreResult, normalizedName string, normalizedDescription string) {
		inputSnapshot := *input
		deadline := input.Deadline
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
				Deadline:            deadline,
				SaveAsChoreTemplate: input.SaveAsChoreTemplate,
			}).
			Return(expected, nil).
			Once()

		result, err := service.Create(context.Background(), input)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(expected))
		Expect(input).To(Equal(&inputSnapshot))
	},
	Entry("with normalized whitespace and empty template flag", &models.CreateChoreInput{
		Name:        " raw name ",
		Description: " raw description ",
		Deadline:    time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
		SaveAsChoreTemplate: false,
	}, &models.CreateChoreResult{
		Chore: &models.Chore{
			Id:          1,
			Name:        "Normalized name",
			Description: "Normalized description",
			Deadline:    time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
		},
	}, "Normalized name", "Normalized description"),
	Entry("with save as chore template", &models.CreateChoreInput{
		Name:                "another chore",
		Description:         "details",
		Deadline:            time.Date(2026, time.July, 28, 0, 0, 0, 0, time.UTC),
		SaveAsChoreTemplate: true,
	}, &models.CreateChoreResult{
		Chore: &models.Chore{
			Id:          2,
			Name:        "another chore",
			Description: "details",
			Deadline:    time.Date(2026, time.July, 28, 0, 0, 0, 0, time.UTC),
		},
	}, "another chore", "details"),
)

	It("rejects an invalid name without persisting", func() {
		normalizer.EXPECT().
			NormalizeName("invalid name").
			Return("", false).
			Once()

		result, err := service.Create(context.Background(), &models.CreateChoreInput{
			Name:        "invalid name",
			Description: "description",
			Deadline:    time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
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
			Name:        "name",
			Description: "invalid description",
			Deadline:    time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
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
			Name:        "name",
			Description: "description",
			Deadline:    deadline,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("create manual one-off chore: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})

})
