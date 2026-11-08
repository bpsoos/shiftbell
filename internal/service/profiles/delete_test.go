package profiles_test

import (
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	"github.com/bpsoos/shiftbell/internal/service/profiles"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete", func() {
	It("deletes the requested profile", func(ctx SpecContext) {
		persister := NewMockPersister(GinkgoT())
		persister.On("Delete", ctx, 7).Return(nil).Once()
		service := profiles.NewService(&profiles.Deps{Persister: persister})
		Expect(service.Delete(ctx, 7)).To(Succeed())
	})

	DescribeTable("preserves persistence errors", func(ctx SpecContext, failure error) {
		persister := NewMockPersister(GinkgoT())
		persister.On("Delete", ctx, 7).Return(failure).Once()
		service := profiles.NewService(&profiles.Deps{Persister: persister})
		Expect(service.Delete(ctx, 7)).To(MatchError(failure))
	},
		Entry("missing profile", models.ErrNotFound),
		Entry("storage failure", errors.New("storage failed")),
	)
})
