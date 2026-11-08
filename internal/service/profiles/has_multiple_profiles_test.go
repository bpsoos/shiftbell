package profiles_test

import (
	"errors"

	"github.com/bpsoos/shiftbell/internal/service/profiles"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HasMultipleProfiles", func() {
	DescribeTable(
		"returns the persisted profile state",
		func(ctx SpecContext, multiple bool) {
			persister := NewMockPersister(GinkgoT())
			persister.On("HasMultipleProfiles", ctx).Return(multiple, nil).Once()
			service := profiles.NewService(&profiles.Deps{Persister: persister})
			result, err := service.HasMultipleProfiles(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(multiple))
		},
		Entry("zero or one profile", false),
		Entry("multiple profiles", true),
	)

	It("preserves storage errors", func(ctx SpecContext) {
		persister := NewMockPersister(GinkgoT())
		failure := errors.New("storage failed")
		persister.On("HasMultipleProfiles", ctx).Return(false, failure).Once()
		service := profiles.NewService(&profiles.Deps{Persister: persister})
		result, err := service.HasMultipleProfiles(ctx)
		Expect(result).To(BeFalse())
		Expect(errors.Is(err, failure)).To(BeTrue())
	})
})
