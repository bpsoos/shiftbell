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

var _ = Describe("Chore assignment API", func() {
	var client *shiftbellapi.APIClient
	var first, second int

	BeforeEach(func(ctx SpecContext) {
		baseURL := os.Getenv("SHIFTBELL_BASE_URL")
		Expect(baseURL).NotTo(BeEmpty())
		client = shiftbellapi.NewAPIClient(baseURL)
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
			fmt.Sprintf("OwnerA%d", time.Now().UnixNano()),
		)
		Expect(err).NotTo(HaveOccurred())
		b, err := client.CreateProfile(
			ctx,
			profiles.Actions.Href(shiftbellapi.RelationCreate),
			fmt.Sprintf("OwnerB%d", time.Now().UnixNano()),
		)
		Expect(err).NotTo(HaveOccurred())
		first, second = a.Id, b.Id
	})

	It("assigns, reassigns and clears a chore's owner", func(ctx SpecContext) {
		created := createAssignmentChore(ctx, client)
		Expect(created.Chore.OwnerProfileId).To(BeNil())
		href := created.Actions.Href(shiftbellapi.RelationSetAssignment)
		Expect(href).NotTo(BeEmpty())
		expected := created.Chore

		By("assigning the chore")
		assigned, err := client.SetChoreAssignment(ctx, href, first)
		Expect(err).NotTo(HaveOccurred())
		Expect(assigned.StatusCode).To(Equal(http.StatusOK))
		Expect(assigned.SuccessResponse).NotTo(BeNil())
		expected.OwnerProfileId = &first
		Expect(assigned.SuccessResponse.Chore).To(matchChore(expected))
		expectAssignmentReads(ctx, client, expected)

		By("reassigning the chore")
		reassigned, err := client.SetChoreAssignment(ctx, href, second)
		Expect(err).NotTo(HaveOccurred())
		Expect(reassigned.StatusCode).To(Equal(http.StatusOK))
		Expect(reassigned.SuccessResponse).NotTo(BeNil())
		expected.OwnerProfileId = &second
		Expect(reassigned.SuccessResponse.Chore).To(matchChore(expected))
		expectAssignmentReads(ctx, client, expected)

		By("clearing the chore's owner")
		clearHref := reassigned.SuccessResponse.Actions.Href(
			shiftbellapi.RelationClearAssignment,
		)
		Expect(clearHref).NotTo(BeEmpty())
		cleared, err := client.ClearChoreAssignment(ctx, clearHref)
		Expect(err).NotTo(HaveOccurred())
		Expect(cleared.StatusCode).To(Equal(http.StatusOK))
		Expect(cleared.SuccessResponse).NotTo(BeNil())
		expected.OwnerProfileId = nil
		Expect(cleared.SuccessResponse.Chore).To(matchChore(expected))
		expectAssignmentReads(ctx, client, expected)
	})

	It("assigns an owner to a completed chore", func(ctx SpecContext) {
		created := createAssignmentChore(ctx, client)
		completed := completeOneOffChore(ctx, client, created, "2020-02-04")
		href := completed.Actions.Href(shiftbellapi.RelationSetAssignment)
		Expect(href).NotTo(BeEmpty())

		assigned, err := client.SetChoreAssignment(ctx, href, first)
		Expect(err).NotTo(HaveOccurred())
		Expect(assigned.StatusCode).To(Equal(http.StatusOK))
		Expect(assigned.SuccessResponse).NotTo(BeNil())
		expected := completed.Chore
		expected.OwnerProfileId = &first
		Expect(assigned.SuccessResponse.Chore).To(matchChore(expected))
		expectAssignmentReads(ctx, client, expected)
	})

	It(
		"completes an assigned scheduled chore with an unassigned successor",
		func(ctx SpecContext) {
			created := createScheduledChore(
				ctx,
				client,
				uniqueChoreName("Assigned schedule"),
				7,
			)
			assigned, err := client.SetChoreAssignment(
				ctx,
				created.Actions.Href(shiftbellapi.RelationSetAssignment),
				first,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(assigned.StatusCode).To(Equal(http.StatusOK))
			completed, err := client.CompleteChore(
				ctx,
				created.Actions.Href(shiftbellapi.RelationComplete),
				shiftbellapi.CompleteChoreParams{CompletedOn: "2020-02-04"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(completed.StatusCode).To(Equal(http.StatusOK))
			Expect(completed.SuccessResponse).NotTo(BeNil())
			Expect(completed.SuccessResponse.Chore.OwnerProfileId).To(Equal(&first))
			expectAssignmentReads(ctx, client, completed.SuccessResponse.Chore)
			page, err := client.BrowseChores(
				ctx,
				shiftbellapi.BrowseChoresParams{
					Href:   created.Chore.Links.Href(shiftbellapi.RelationCollection),
					Search: created.Chore.Name,
					Status: "active",
					Limit:  20,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(page.Collection.Items).To(HaveLen(1))
			successor := page.Collection.Items[0]
			Expect(successor.OwnerProfileId).To(BeNil())
			Expect(successor.ScheduleId).To(Equal(created.Chore.ScheduleId))
			expectAssignmentReads(ctx, client, successor)
		},
	)
})

func createAssignmentChore(
	ctx context.Context,
	client *shiftbellapi.APIClient,
) *shiftbellapi.CreateChoreResponse {
	GinkgoHelper()
	form := getManualOneOffChoreForm(ctx, client, discoverChoreCollection(ctx, client))
	return createManualOneOffChore(
		ctx,
		client,
		form,
		uniqueChoreName("Assignment"),
		"Steps",
		"2020-02-03",
	)
}

func expectAssignmentReads(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	expected shiftbellapi.Chore,
) {
	GinkgoHelper()
	detail, err := client.GetChore(ctx, expected.Links.Href(shiftbellapi.RelationSelf))
	Expect(err).NotTo(HaveOccurred())
	Expect(detail.StatusCode).To(Equal(http.StatusOK))
	Expect(detail.SuccessResponse.Chore).To(matchChore(expected))
	page, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
		Href: expected.Links.Href(
			shiftbellapi.RelationCollection,
		),
		Search: expected.Name,
		Status: expected.Status,
		Limit:  20,
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(page.Collection.Items).To(ContainElement(matchChore(expected)))
}
