package chores_test

import (
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	service "github.com/bpsoos/shiftbell/internal/service/chores"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Set assignment", func() {
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
		"accepts assignments regardless of chore status",
		func(ctx SpecContext, status models.ChoreStatus, owner int) {
			params := &models.SetAssignmentParams{Id: 42, OwnerProfileId: owner}
			persisted := &models.Chore{Id: 42, Status: status, OwnerProfileId: &owner}
			persister.EXPECT().SetAssignment(ctx, params).Return(persisted, nil).Once()
			result, err := subject.SetAssignment(ctx, params)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeIdenticalTo(persisted))
		},
		Entry("assigns active chores", models.ChoreStatusActive, 1),
		Entry("assigns completed chores", models.ChoreStatusCompleted, 1),
	)

	DescribeTable("preserves persistence errors", func(ctx SpecContext, failure error) {
		params := &models.SetAssignmentParams{Id: 42, OwnerProfileId: 1}
		persister.EXPECT().SetAssignment(ctx, params).Return(nil, failure).Once()
		result, err := subject.SetAssignment(ctx, params)
		Expect(result).To(BeNil())
		Expect(errors.Is(err, failure)).To(BeTrue())
	}, Entry("missing chore", models.ErrNotFound), Entry("missing profile", models.ErrInvalidOwnerProfile), Entry("database failure", errors.New("database failure")))
})
