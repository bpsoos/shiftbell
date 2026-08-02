package schedules

import (
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get", func() {
	var (
		persister *MockPersister
		service   *Service
	)

	BeforeEach(func() {
		persister = NewMockPersister(GinkgoT())
		service = NewService(&Deps{Persister: persister}, &Config{})
	})

	It("returns schedule details", func(ctx SpecContext) {
		details := &models.ScheduleDetails{
			Id:                42,
			Name:              "Laundry",
			ChoreTemplateId:   7,
			ChoreTemplateName: "Wash clothes",
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
		Expect(err).To(MatchError("get schedule: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})
})
