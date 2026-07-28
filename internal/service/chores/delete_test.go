package chores

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete", func() {
	var (
		persister *MockPersister
		service   *Service
	)

	BeforeEach(func() {
		persister = NewMockPersister(GinkgoT())
		service = NewService(&Deps{Persister: persister}, &Config{})
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
