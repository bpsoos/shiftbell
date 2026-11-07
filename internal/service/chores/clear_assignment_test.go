package chores_test

import (
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	service "github.com/bpsoos/shiftbell/internal/service/chores"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Clear assignment", func() {
	var persister *MockPersister
	var subject *service.Service

	BeforeEach(func() {
		persister = NewMockPersister(GinkgoT())
		subject = service.NewService(
			&service.Deps{Persister: persister},
			&service.Config{},
		)
	})

	DescribeTable(
		"clears assignments regardless of chore status",
		func(ctx SpecContext, status models.ChoreStatus) {
			persisted := &models.Chore{Id: 42, Status: status}
			persister.EXPECT().ClearAssignment(ctx, 42).Return(persisted, nil).Once()
			result, err := subject.ClearAssignment(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeIdenticalTo(persisted))
		},
		Entry("active", models.ChoreStatusActive),
		Entry("completed", models.ChoreStatusCompleted),
	)

	DescribeTable("preserves persistence errors", func(ctx SpecContext, failure error) {
		persister.EXPECT().ClearAssignment(ctx, 42).Return(nil, failure).Once()
		result, err := subject.ClearAssignment(ctx, 42)
		Expect(result).To(BeNil())
		Expect(errors.Is(err, failure)).To(BeTrue())
	}, Entry("missing chore", models.ErrNotFound), Entry("database failure", errors.New("database failure")))
})
