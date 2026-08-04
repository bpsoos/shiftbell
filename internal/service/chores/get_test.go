package chores_test

import (
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	choresservice "github.com/bpsoos/shiftbell/internal/service/chores"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get", func() {
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

	It("returns chore details", func(ctx SpecContext) {
		details := &models.ChoreDetails{
			Id:          42,
			Status:      models.ChoreStatusActive,
			Name:        "Kitchen",
			Description: "Wash and fold.",
		}
		persister.EXPECT().Get(ctx, 42).Return(details, nil).Once()

		result, err := service.Get(ctx, 42)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(details))
	})

	It("preserves persistence errors", func(ctx SpecContext) {
		persistErr := errors.New("persistence failed")
		persister.EXPECT().Get(ctx, 42).Return(nil, persistErr).Once()

		result, err := service.Get(ctx, 42)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("getting chore: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})
})
