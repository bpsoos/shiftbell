package profiles_test

import (
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	validation "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/bpsoos/shiftbell/internal/service/profiles"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Setup", func() {
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

	It("normalizes and atomically creates the first profile", func(ctx SpecContext) {
		normalizer.On("NormalizeName", " Alex ").Return("Alex", nil).Once()
		profile := &models.Profile{Id: 1, Name: "Alex"}
		persister.On("CreateInitial", ctx, &models.CreateParams{Name: "Alex"}).
			Return(profile, nil).
			Once()
		result, err := service.Setup(ctx, &models.CreateParams{Name: " Alex "})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(profile))
	})

	It("rejects invalid names before writing", func(ctx SpecContext) {
		normalizer.On("NormalizeName", "Alex Smith").Return("Alex Smith", nil).Once()
		result, err := service.Setup(ctx, &models.CreateParams{Name: "Alex Smith"})
		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validation.ErrInvalidName))
	})

	DescribeTable(
		"preserves initialization errors",
		func(ctx SpecContext, failure error) {
			normalizer.On("NormalizeName", "Alex").Return("Alex", nil).Once()
			persister.On("CreateInitial", ctx, &models.CreateParams{Name: "Alex"}).
				Return(nil, failure).
				Once()
			result, err := service.Setup(ctx, &models.CreateParams{Name: "Alex"})
			Expect(result).To(BeNil())
			Expect(errors.Is(err, failure)).To(BeTrue())
		},
		Entry("already set up", models.ErrAlreadyInitialized),
		Entry("storage", errors.New("storage failed")),
	)

	DescribeTable(
		"requires setup only when there are no profiles",
		func(ctx SpecContext, exists, required bool) {
			persister.On("HasProfiles", ctx).Return(exists, nil).Once()
			result, err := service.SetupRequired(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(required))
		},
		Entry("empty", false, true),
		Entry("populated", true, false),
	)

	It("preserves setup status errors", func(ctx SpecContext) {
		failure := errors.New("storage failed")
		persister.On("HasProfiles", ctx).Return(false, failure).Once()
		_, err := service.SetupRequired(ctx)
		Expect(errors.Is(err, failure)).To(BeTrue())
	})
})
