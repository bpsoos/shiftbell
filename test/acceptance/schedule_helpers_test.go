package acceptance_test

import (
	"context"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/testsupport/shiftbellapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func createScheduledChore(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	name string,
	interval int,
) *shiftbellapi.CreateChoreResponse {
	GinkgoHelper()
	template := createChoreTemplate(
		ctx,
		client,
		discoverChoreTemplateCollection(ctx, client),
		name,
		"Reusable steps.",
	)
	form := getTemplateOneOffChoreForm(
		ctx,
		client,
		discoverChoreCollection(ctx, client),
		template.ChoreTemplate,
	)
	created, err := client.CreateChore(
		ctx,
		form.Actions.Href(shiftbellapi.RelationCreate),
		shiftbellapi.CreateChoreParams{
			ChoreTemplateId: &template.ChoreTemplate.Id, ScheduleName: name + " weekly",
			IntervalDays: &interval, Deadline: "2020-02-03",
		},
	)
	Expect(err).NotTo(HaveOccurred())
	Expect(created.StatusCode).To(Equal(http.StatusCreated))
	Expect(created.SuccessResponse).NotTo(BeNil())
	return created.SuccessResponse
}

func browseAllSchedules(
	ctx context.Context,
	client *shiftbellapi.APIClient,
) []shiftbellapi.Schedule {
	GinkgoHelper()
	home, err := client.GetHome(ctx)
	Expect(err).NotTo(HaveOccurred())
	href := home.Home.Links.Href(shiftbellapi.RelationSchedules)
	Expect(href).NotTo(BeEmpty())
	return browseAllSchedulesFrom(ctx, client, href)
}

func browseAllSchedulesFrom(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	href string,
) []shiftbellapi.Schedule {
	GinkgoHelper()
	items := make([]shiftbellapi.Schedule, 0)
	for href != "" {
		collection, err := client.GetSchedules(ctx, href)
		Expect(err).NotTo(HaveOccurred())
		items = append(items, collection.Items...)
		href = collection.Links.Href(shiftbellapi.RelationNext)
	}
	return items
}
