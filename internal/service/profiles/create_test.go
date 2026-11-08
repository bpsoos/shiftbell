package profiles_test

import (
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	validation "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/bpsoos/shiftbell/internal/service/profiles"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create", func() {
	var persister *MockPersister
	var normalizer *MockNormalizer
	var service *profiles.Service
	BeforeEach(func() {
		persister = &MockPersister{}
		normalizer = &MockNormalizer{}
		DeferCleanup(func() {
			persister.AssertExpectations(GinkgoT())
			normalizer.AssertExpectations(GinkgoT())
		})
		service = profiles.NewService(
			&profiles.Deps{Persister: persister, Normalizer: normalizer},
		)
	})

	It("normalizes the name without changing the caller's input", func(ctx SpecContext) {
		params := &models.CreateParams{Name: " Álex "}
		profile := &models.Profile{Id: 7, Name: "Álex"}
		normalizer.On("NormalizeName", params.Name).Return("Álex", nil).Once()
		persister.On("Create", ctx, &models.CreateParams{Name: "Álex"}).
			Return(profile, nil).
			Once()
		result, err := service.Create(ctx, params)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(profile))
		Expect(params.Name).To(Equal(" Álex "))
	})

	It("rejects missing input without calling dependencies", func(ctx SpecContext) {
		result, err := service.Create(ctx, nil)
		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validation.ErrInvalidName))
	})

	DescribeTable(
		"preserves the reason for an invalid name",
		func(ctx SpecContext, reason error) {
			normalizer.On("NormalizeName", "").Return("", reason).Once()
			result, err := service.Create(ctx, &models.CreateParams{})
			Expect(result).To(BeNil())
			Expect(err).To(MatchError(validation.ErrInvalidName))
			Expect(err).To(MatchError(reason))
		},
		Entry("required", validation.ErrRequired),
		Entry("too long", validation.ErrTooLong),
		Entry("disallowed character", validation.ErrDisallowedCharacter),
		Entry("invalid encoding", validation.ErrInvalidUTF8),
	)

	DescribeTable(
		"rejects whitespace remaining after normalization",
		func(ctx SpecContext, name string) {
			normalizer.On("NormalizeName", name).Return(name, nil).Once()
			result, err := service.Create(ctx, &models.CreateParams{Name: name})
			Expect(result).To(BeNil())
			Expect(err).To(MatchError(validation.ErrInvalidName))
			Expect(err).To(MatchError(models.ErrNameContainsSpaces))
		},
		Entry("space", "Alex Smith"),
		Entry("tab", "Alex\tSmith"),
		Entry("newline", "Alex\nSmith"),
		Entry("nonbreaking space", "Alex\u00a0Smith"),
		Entry("em space", "Alex\u2003Smith"),
	)

	DescribeTable("preserves persistence errors", func(ctx SpecContext, failure error) {
		normalizer.On("NormalizeName", "Alex").Return("Alex", nil).Once()
		persister.On("Create", ctx, &models.CreateParams{Name: "Alex"}).
			Return(nil, failure).
			Once()
		result, err := service.Create(ctx, &models.CreateParams{Name: "Alex"})
		Expect(result).To(BeNil())
		Expect(errors.Is(err, failure)).To(BeTrue())
	},
		Entry("conflict", models.ErrNameConflict),
		Entry("limit reached", models.ErrLimitReached),
		Entry("storage", errors.New("storage failed")),
	)
})
