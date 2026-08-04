package choretemplates_test

import (
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	choretemplatesservice "github.com/bpsoos/shiftbell/internal/service/choretemplates"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get", func() {
	var (
		persister *MockPersister
		service   *choretemplatesservice.Service
	)

	BeforeEach(func() {
		persister = NewMockPersister(GinkgoT())
		service = choretemplatesservice.NewService(
			&choretemplatesservice.Deps{Persister: persister},
			&choretemplatesservice.Config{},
		)
	})

	It("returns chore template details", func(ctx SpecContext) {
		details := &models.ChoreTemplateDetails{
			ChoreTemplate:            models.ChoreTemplate{Id: 42, Name: "Laundry"},
			ActiveScheduleCount:      2,
			DeactivatedScheduleCount: 1,
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
		Expect(err).To(MatchError("get chore template: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})
})
