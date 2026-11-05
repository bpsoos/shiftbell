package profiles_test

import (
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	"github.com/bpsoos/shiftbell/internal/service/profiles"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("List", func() {
	It("returns the persisted list", func(ctx SpecContext) {
		persister := &MockPersister{}
		DeferCleanup(func() { persister.AssertExpectations(GinkgoT()) })
		items := []models.Profile{{Id: 3, Name: "Alex"}}
		persister.On("List", ctx).Return(items, nil).Once()
		service := profiles.NewService(&profiles.Deps{Persister: persister})
		result, err := service.List(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(items))
	})

	It("preserves storage errors", func(ctx SpecContext) {
		persister := &MockPersister{}
		DeferCleanup(func() { persister.AssertExpectations(GinkgoT()) })
		failure := errors.New("storage failed")
		persister.On("List", ctx).Return(nil, failure).Once()
		service := profiles.NewService(&profiles.Deps{Persister: persister})
		result, err := service.List(ctx)
		Expect(result).To(BeNil())
		Expect(errors.Is(err, failure)).To(BeTrue())
	})
})
