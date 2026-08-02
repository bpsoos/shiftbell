package choretemplates

import (
	"errors"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deactivate", func() {
	var (
		persister     *MockPersister
		service       *Service
		deactivatedAt time.Time
	)

	BeforeEach(func() {
		deactivatedAt = time.Date(2026, time.July, 29, 12, 30, 0, 0, time.UTC)
		persister = NewMockPersister(GinkgoT())
		service = NewService(&Deps{
			Persister: persister,
			Now:       func() time.Time { return deactivatedAt },
		}, &Config{})
	})

	It("permanently deactivates a chore template", func(ctx SpecContext) {
		persisted := &models.ChoreTemplate{
			Id:            42,
			Name:          "Laundry",
			DeactivatedAt: &deactivatedAt,
		}
		persister.EXPECT().Deactivate(ctx, &models.DeactivateChoreTemplateParams{
			Id:            42,
			DeactivatedAt: deactivatedAt,
		}).Return(persisted, nil).Once()

		result, err := service.Deactivate(ctx, 42)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(persisted))
	})

	It(
		"preserves active schedule references that block deactivation",
		func(ctx SpecContext) {
			blocked := &models.ActiveScheduleReferencesError{
				Schedules: []models.ActiveScheduleReference{
					{Id: 7, Name: "Every two weeks"},
				},
			}
			persister.EXPECT().Deactivate(ctx, &models.DeactivateChoreTemplateParams{
				Id:            42,
				DeactivatedAt: deactivatedAt,
			}).Return(nil, blocked).Once()

			result, err := service.Deactivate(ctx, 42)

			Expect(result).To(BeNil())
			var actual *models.ActiveScheduleReferencesError
			Expect(errors.As(err, &actual)).To(BeTrue())
			Expect(actual).To(BeIdenticalTo(blocked))
		},
	)
})
