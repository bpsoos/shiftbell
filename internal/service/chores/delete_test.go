package chores_test

import (
	"errors"

	choresservice "github.com/bpsoos/shiftbell/internal/service/chores"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete", func() {
	var (
		persister *MockPersister
		service   *choresservice.Service
	)

	BeforeEach(func() {
		persister = NewMockPersister(GinkgoT())
		service = choresservice.NewService(
			&choresservice.Deps{Persister: persister},
			&choresservice.Config{},
		)
	})

	It("deletes a chore by ID", func(ctx SpecContext) {
		persister.EXPECT().Delete(ctx, 42).Return(nil).Once()

		err := service.Delete(ctx, 42)

		Expect(err).NotTo(HaveOccurred())
	})

	It("preserves persistence errors", func(ctx SpecContext) {
		persistErr := errors.New("persistence failed")
		persister.EXPECT().Delete(ctx, 42).Return(persistErr).Once()

		err := service.Delete(ctx, 42)

		Expect(err).To(MatchError("delete chore: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})
})
