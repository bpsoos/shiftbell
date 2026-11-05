package profiles_test

import (
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	validation "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/bpsoos/shiftbell/internal/service/profiles"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Initialize", func() {
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

	It("leaves setup to the user when no name is configured", func(ctx SpecContext) {
		Expect(service.Initialize(ctx, "")).To(Succeed())
	})

	It("ignores configuration once any profile exists", func(ctx SpecContext) {
		persister.On("HasProfiles", ctx).Return(true, nil).Once()
		Expect(service.Initialize(ctx, "even an invalid name")).To(Succeed())
	})

	It("creates the configured first profile", func(ctx SpecContext) {
		persister.On("HasProfiles", ctx).Return(false, nil).Once()
		normalizer.On("NormalizeName", " Alex ").Return("Alex", nil).Once()
		persister.On("CreateInitial", ctx, &models.CreateParams{Name: "Alex"}).
			Return(&models.Profile{Id: 1, Name: "Alex"}, nil).
			Once()
		Expect(service.Initialize(ctx, " Alex ")).To(Succeed())
	})

	It("accepts another initializer winning the race", func(ctx SpecContext) {
		persister.On("HasProfiles", ctx).Return(false, nil).Once()
		normalizer.On("NormalizeName", "Alex").Return("Alex", nil).Once()
		persister.On("CreateInitial", ctx, &models.CreateParams{Name: "Alex"}).
			Return(nil, models.ErrAlreadyInitialized).
			Once()
		Expect(service.Initialize(ctx, "Alex")).To(Succeed())
	})

	It("rejects invalid configured names before inserting", func(ctx SpecContext) {
		persister.On("HasProfiles", ctx).Return(false, nil).Once()
		normalizer.On("NormalizeName", " ").Return("", validation.ErrRequired).Once()
		Expect(
			errors.Is(service.Initialize(ctx, " "), validation.ErrInvalidName),
		).To(BeTrue())
	})

	It("propagates status lookup failures", func(ctx SpecContext) {
		failure := errors.New("storage failed")
		persister.On("HasProfiles", ctx).Return(false, failure).Once()
		Expect(errors.Is(service.Initialize(ctx, "Alex"), failure)).To(BeTrue())
	})

	It("propagates initial creation failures", func(ctx SpecContext) {
		failure := errors.New("storage failed")
		persister.On("HasProfiles", ctx).Return(false, nil).Once()
		normalizer.On("NormalizeName", "Alex").Return("Alex", nil).Once()
		persister.On("CreateInitial", ctx, &models.CreateParams{Name: "Alex"}).
			Return(nil, failure).
			Once()
		Expect(errors.Is(service.Initialize(ctx, "Alex"), failure)).To(BeTrue())
	})
})
