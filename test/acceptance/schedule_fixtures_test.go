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

func createDeactivatedSchedule(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	name string,
) *shiftbellapi.ScheduleRepresentation {
	GinkgoHelper()
	created := createScheduledChore(ctx, client, name, 7)
	Expect(created.Chore.ScheduleId).NotTo(BeNil())
	var item shiftbellapi.Schedule
	Expect(
		browseAllSchedules(ctx, client),
	).To(ContainElement(HaveField("Id", *created.Chore.ScheduleId), &item))
	details, err := client.GetSchedule(ctx, item.Links.Href(shiftbellapi.RelationSelf))
	Expect(err).NotTo(HaveOccurred())
	Expect(details.StatusCode).To(Equal(http.StatusOK))
	Expect(details.SuccessResponse).NotTo(BeNil())
	href := details.SuccessResponse.Actions.Href(shiftbellapi.RelationDeactivate)
	Expect(href).NotTo(BeEmpty())
	deactivated, err := client.DeactivateSchedule(ctx, href)
	Expect(err).NotTo(HaveOccurred())
	Expect(deactivated.StatusCode).To(Equal(http.StatusOK))
	Expect(deactivated.SuccessResponse).NotTo(BeNil())
	return deactivated.SuccessResponse
}
