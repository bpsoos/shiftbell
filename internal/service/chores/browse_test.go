package chores

import (
	"context"
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	serviceerrors "github.com/bpsoos/shiftbell/internal/service"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Browse", func() {
	var (
		persister  *MockPersister
		normalizer *MockNormalizer
		service    *Service
	)

	BeforeEach(func() {
		persister = NewMockPersister(GinkgoT())
		normalizer = NewMockNormalizer(GinkgoT())
		service = NewService(&Deps{
			Persister:  persister,
			Normalizer: normalizer,
		}, &Config{})
	})

	It("defaults to active chores", func(ctx SpecContext) {
		params := &models.BrowseChoresParams{
			Status: "",
			Offset: 20,
			Limit:  10,
		}
		persisted := &models.ChorePage{
			Chores: []models.Chore{{Id: 4}},
			More:   true,
		}
		persister.EXPECT().
			Browse(ctx, &models.BrowseChoresParams{
				Status: models.ChoreStatusActive,
				Offset: 20,
				Limit:  10,
			}).
			Return(persisted, nil).
			Once()

		result, err := service.Browse(ctx, params)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(persisted))
		Expect(params).To(Equal(&models.BrowseChoresParams{
			Status: "",
			Offset: 20,
			Limit:  10,
		}))
	})

	It("browses completed chores", func(ctx SpecContext) {
		params := &models.BrowseChoresParams{
			Status: models.ChoreStatusCompleted,
			Offset: 0,
			Limit:  10,
		}
		persisted := &models.ChorePage{
			Chores: []models.Chore{{Id: 4, Status: models.ChoreStatusCompleted}},
		}
		persister.EXPECT().
			Browse(ctx, &models.BrowseChoresParams{
				Status: models.ChoreStatusCompleted,
				Offset: 0,
				Limit:  10,
			}).
			Return(persisted, nil).
			Once()

		result, err := service.Browse(ctx, params)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(persisted))
		Expect(params).To(Equal(&models.BrowseChoresParams{
			Status: models.ChoreStatusCompleted,
			Offset: 0,
			Limit:  10,
		}))
	})

	It("preserves persistence errors", func(ctx SpecContext) {
		persistErr := errors.New("persistence failed")
		persister.EXPECT().
			Browse(ctx, &models.BrowseChoresParams{
				Status: models.ChoreStatusActive,
				Offset: 0,
				Limit:  10,
			}).
			Return(nil, persistErr).
			Once()

		result, err := service.Browse(ctx, &models.BrowseChoresParams{
			Status: models.ChoreStatusActive,
			Offset: 0,
			Limit:  10,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError("browse chores: persistence failed"))
		Expect(errors.Is(err, persistErr)).To(BeTrue())
	})

	It("rejects an unsupported status", func() {
		result, err := service.Browse(context.Background(), &models.BrowseChoresParams{
			Status: "all",
			Offset: 0,
			Limit:  10,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(serviceerrors.ErrInvalidFilter))
	})

	It("rejects a negative offset", func() {
		result, err := service.Browse(context.Background(), &models.BrowseChoresParams{
			Status: "",
			Offset: -1,
			Limit:  10,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(serviceerrors.ErrInvalidOffset))
	})

	It("rejects a non-positive limit", func() {
		result, err := service.Browse(context.Background(), &models.BrowseChoresParams{
			Status: "",
			Offset: 0,
			Limit:  0,
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(serviceerrors.ErrInvalidLimit))
	})
})
