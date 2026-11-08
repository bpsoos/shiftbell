package acceptance_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bpsoos/shiftbell/internal/testsupport/shiftbellapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chore ownership filtering JSON API", func() {
	var client *shiftbellapi.APIClient
	var collectionHref, search string
	var firstOwner, secondOwner int
	var first, second, other, unassigned *shiftbellapi.CreateChoreResponse

	BeforeEach(func(ctx SpecContext) {
		baseURL := os.Getenv("SHIFTBELL_BASE_URL")
		Expect(baseURL).NotTo(BeEmpty())
		client = shiftbellapi.NewAPIClient(baseURL)
		collection := discoverChoreCollection(ctx, client)
		collectionHref = collection.Links.Href(shiftbellapi.RelationSelf)

		home, err := client.GetHome(ctx)
		Expect(err).NotTo(HaveOccurred())
		profiles, err := client.GetProfiles(
			ctx,
			home.Home.Links.Href(shiftbellapi.RelationProfiles),
		)
		Expect(err).NotTo(HaveOccurred())
		a, err := client.CreateProfile(
			ctx,
			profiles.Actions.Href(shiftbellapi.RelationCreate),
			fmt.Sprintf("FilterA%d", time.Now().UnixNano()),
		)
		Expect(err).NotTo(HaveOccurred())
		b, err := client.CreateProfile(
			ctx,
			profiles.Actions.Href(shiftbellapi.RelationCreate),
			fmt.Sprintf("FilterB%d", time.Now().UnixNano()),
		)
		Expect(err).NotTo(HaveOccurred())
		firstOwner, secondOwner = a.Id, b.Id

		search = uniqueChoreName("Ownership filter")
		form := getManualOneOffChoreForm(ctx, client, collection)
		first = createManualOneOffChore(
			ctx,
			client,
			form,
			search+" first",
			"Steps",
			"2020-02-01",
		)
		second = createManualOneOffChore(
			ctx,
			client,
			form,
			search+" second",
			"Steps",
			"2020-02-02",
		)
		other = createManualOneOffChore(
			ctx,
			client,
			form,
			search+" other",
			"Steps",
			"2020-02-03",
		)
		unassigned = createManualOneOffChore(
			ctx,
			client,
			form,
			search+" unassigned",
			"Steps",
			"2020-02-04",
		)
		assignChoreOwner(ctx, client, first, firstOwner)
		assignChoreOwner(ctx, client, second, firstOwner)
		assignChoreOwner(ctx, client, other, secondOwner)
	})

	It("includes assigned and unassigned chores by default", func(ctx SpecContext) {
		page, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
			Href:   collectionHref,
			Search: search,
			Limit:  20,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(page.Collection.Items).To(ConsistOf(
			HaveField("Id", first.Chore.Id),
			HaveField("Id", second.Chore.Id),
			HaveField("Id", other.Chore.Id),
			HaveField("Id", unassigned.Chore.Id),
		))
	})

	It("includes assigned and unassigned chores in all mode", func(ctx SpecContext) {
		page, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
			Href:      collectionHref,
			OwnerMode: "all",
			Search:    search,
			Limit:     20,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(page.Collection.Items).To(ConsistOf(
			HaveField("Id", first.Chore.Id),
			HaveField("Id", second.Chore.Id),
			HaveField("Id", other.Chore.Id),
			HaveField("Id", unassigned.Chore.Id),
		))
	})

	It("paginates a profile's chores through collection links", func(ctx SpecContext) {
		page, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
			Href:      collectionHref,
			OwnerMode: "profile",
			OwnerId:   &firstOwner,
			Status:    "active",
			Search:    search,
			Limit:     1,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(page.Collection.Items).To(ConsistOf(HaveField("Id", first.Chore.Id)))
		Expect(page.Collection.More).To(BeTrue())

		nextHref := page.Collection.Links.Href(shiftbellapi.RelationNext)
		Expect(nextHref).NotTo(BeEmpty())
		next, err := client.GetChores(ctx, nextHref)
		Expect(err).NotTo(HaveOccurred())
		Expect(next.Collection.Items).To(ConsistOf(HaveField("Id", second.Chore.Id)))
		Expect(next.Collection.More).To(BeFalse())
		Expect(next.Collection.Links.Href(shiftbellapi.RelationNext)).To(BeEmpty())

		previousHref := next.Collection.Links.Href(shiftbellapi.RelationPrevious)
		Expect(previousHref).NotTo(BeEmpty())
		previous, err := client.GetChores(ctx, previousHref)
		Expect(err).NotTo(HaveOccurred())
		Expect(previous.Collection.Items).To(Equal(page.Collection.Items))
	})

	It("returns only unassigned chores", func(ctx SpecContext) {
		page, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
			Href:      collectionHref,
			OwnerMode: "unassigned",
			Search:    search,
			Limit:     20,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(page.Collection.Items).To(ConsistOf(HaveField("Id", unassigned.Chore.Id)))
	})

	It("filters completed chores by profile", func(ctx SpecContext) {
		completeOneOffChore(ctx, client, first, "2020-02-06")
		completeOneOffChore(ctx, client, other, "2020-02-06")
		completeOneOffChore(ctx, client, unassigned, "2020-02-06")

		page, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
			Href:      collectionHref,
			OwnerMode: "profile",
			OwnerId:   &firstOwner,
			Status:    "completed",
			Search:    search,
			Limit:     20,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(page.Collection.Items).To(ConsistOf(HaveField("Id", first.Chore.Id)))
	})
})

func assignChoreOwner(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	chore *shiftbellapi.CreateChoreResponse,
	ownerId int,
) {
	GinkgoHelper()
	assigned, err := client.SetChoreAssignment(
		ctx,
		chore.Actions.Href(shiftbellapi.RelationSetAssignment),
		ownerId,
	)
	Expect(err).NotTo(HaveOccurred())
	Expect(assigned.StatusCode).To(Equal(http.StatusOK))
	Expect(assigned.SuccessResponse).NotTo(BeNil())
	chore.Chore = assigned.SuccessResponse.Chore
	chore.Actions = assigned.SuccessResponse.Actions
}
