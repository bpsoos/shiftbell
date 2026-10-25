package acceptance_test

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/bpsoos/shiftbell/internal/testsupport/shiftbellapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
)

var _ = Describe("Schedule API", func() {
	It("discovers the schedules collection from home", func(ctx SpecContext) {
		baseURL := os.Getenv("SHIFTBELL_BASE_URL")
		Expect(baseURL).NotTo(BeEmpty())
		client := shiftbellapi.NewAPIClient(baseURL)

		By("discovering schedules from home")
		home, err := client.GetHome(ctx)
		Expect(err).NotTo(HaveOccurred())
		href := home.Home.Links.Href(shiftbellapi.RelationSchedules)
		Expect(href).NotTo(BeEmpty())

		By("following the collection link")
		collection, err := client.GetSchedules(ctx, href)
		Expect(err).NotTo(HaveOccurred())
		Expect(collection.Items).NotTo(BeNil())
		Expect(collection.Links).To(ContainElement(shiftbellapi.Relation{
			Rel: shiftbellapi.RelationSelf, Href: href,
		}))
	})

	It(
		"creates a scheduled chore from a template and lists its schedule",
		func(ctx SpecContext) {
			baseURL := os.Getenv("SHIFTBELL_BASE_URL")
			Expect(baseURL).NotTo(BeEmpty())
			client := shiftbellapi.NewAPIClient(baseURL)
			scope := uniqueChoreName("Template schedule")
			template := createChoreTemplate(
				ctx,
				client,
				discoverChoreTemplateCollection(
					ctx,
					client,
				),
				scope,
				"Reusable scheduled steps.",
			)
			form := getTemplateOneOffChoreForm(ctx, client,
				discoverChoreCollection(ctx, client), template.ChoreTemplate)
			interval := 7

			By("creating the schedule and its initial chore")
			result, err := client.CreateChore(
				ctx,
				form.Actions.Href(shiftbellapi.RelationCreate),
				shiftbellapi.CreateChoreParams{
					ChoreTemplateId: &template.ChoreTemplate.Id,
					ScheduleName:    "  " + scope + " weekly  ",
					IntervalDays:    &interval,
					Deadline:        "2020-02-03",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.StatusCode).To(Equal(http.StatusCreated))
			Expect(result.ErrorResponse).To(BeNil())
			Expect(result.SuccessResponse).NotTo(BeNil())
			created := result.SuccessResponse
			Expect(created.Chore.ScheduleId).To(gstruct.PointTo(BeNumerically(">", 0)))
			Expect(created.Chore.Name).To(Equal(template.ChoreTemplate.Name))
			Expect(
				created.Chore.Description,
			).To(Equal(template.ChoreTemplate.Description))
			Expect(created.Chore.Deadline).To(Equal("2020-02-03"))
			Expect(created.Chore.Status).To(Equal("active"))
			Expect(
				created.Location,
			).To(Equal(created.Chore.Links.Href(shiftbellapi.RelationSelf)))
			Expect(created.Actions.Href(shiftbellapi.RelationDelete)).To(BeEmpty())

			By("retrieving the chore with its persisted schedule association")
			retrieved, err := client.GetChore(
				ctx,
				created.Chore.Links.Href(shiftbellapi.RelationSelf),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrieved.StatusCode).To(Equal(http.StatusOK))
			Expect(retrieved.SuccessResponse).NotTo(BeNil())
			Expect(retrieved.SuccessResponse.Chore).To(Equal(created.Chore))
			Expect(retrieved.SuccessResponse.Actions).To(ConsistOf(created.Actions))

			By("finding the schedule through the collection discovered from home")
			deadline := "2020-02-03T00:00:00Z"
			Expect(
				browseAllSchedules(ctx, client),
			).To(ContainElement(shiftbellapi.Schedule{
				Id:                  *created.Chore.ScheduleId,
				Name:                scope + " weekly",
				ChoreTemplateId:     template.ChoreTemplate.Id,
				ChoreTemplateName:   template.ChoreTemplate.Name,
				IntervalDays:        interval,
				ActiveChoreId:       &created.Chore.Id,
				ActiveChoreDeadline: &deadline,
			}))

			By("finding the same association when browsing chores")
			chores, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
				Href: "/chores", Search: scope, Status: "active", Limit: 20,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(chores.Collection.Items).To(ConsistOf(created.Chore))
		},
	)

	It(
		"completes a scheduled chore exactly once and creates its successor from the current template",
		func(ctx SpecContext) {
			baseURL := os.Getenv("SHIFTBELL_BASE_URL")
			Expect(baseURL).NotTo(BeEmpty())
			client := shiftbellapi.NewAPIClient(baseURL)
			scope := uniqueChoreName("Scheduled completion")
			template := createChoreTemplate(
				ctx,
				client,
				discoverChoreTemplateCollection(
					ctx,
					client,
				),
				scope+" original",
				"Original template steps.",
			)
			form := getTemplateOneOffChoreForm(ctx, client,
				discoverChoreCollection(ctx, client), template.ChoreTemplate)
			interval := 7
			result, err := client.CreateChore(
				ctx,
				form.Actions.Href(shiftbellapi.RelationCreate),
				shiftbellapi.CreateChoreParams{
					ChoreTemplateId: &template.ChoreTemplate.Id,
					ScheduleName:    scope + " weekly",
					IntervalDays:    &interval,
					Deadline:        "2020-02-03",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.StatusCode).To(Equal(http.StatusCreated))
			Expect(result.SuccessResponse).NotTo(BeNil())
			created := result.SuccessResponse
			Expect(created.Chore.ScheduleId).NotTo(BeNil())
			completionHref := created.Actions.Href(shiftbellapi.RelationComplete)
			Expect(completionHref).NotTo(BeEmpty())

			By("editing the shared template before completing the original snapshot")
			edited, err := client.EditChoreTemplate(
				ctx,
				template.Actions.Href(shiftbellapi.RelationEdit),
				shiftbellapi.EditChoreTemplateParams{
					Name: scope + " updated", Description: "Updated template steps.",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(edited.StatusCode).To(Equal(http.StatusOK))

			By("completing on a date different from the initial deadline")
			completed, err := client.CompleteChore(
				ctx,
				completionHref,
				shiftbellapi.CompleteChoreParams{CompletedOn: "2020-03-26"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(completed.StatusCode).To(Equal(http.StatusOK))
			Expect(completed.ErrorResponse).To(BeNil())
			Expect(completed.SuccessResponse).NotTo(BeNil())
			expected := created.Chore
			expected.Status = "completed"
			completedOn := "2020-03-26"
			expected.CompletedOn = &completedOn
			Expect(completed.SuccessResponse.Chore).To(Equal(expected))

			By("finding exactly one active successor with a completion-based deadline")
			active, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
				Href: "/chores", Search: scope, Status: "active", Limit: 20,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(active.Collection.Items).To(HaveLen(1))
			successor := active.Collection.Items[0]
			Expect(successor.Id).NotTo(Equal(created.Chore.Id))
			Expect(successor.ScheduleId).To(Equal(created.Chore.ScheduleId))
			Expect(successor.Name).To(Equal(scope + " updated"))
			Expect(successor.Description).To(Equal("Updated template steps."))
			Expect(successor.Status).To(Equal("active"))
			Expect(successor.CompletedOn).To(BeNil())
			Expect(successor.Deadline).To(Equal("2020-04-02"))
			retrievedSuccessor, err := client.GetChore(
				ctx,
				successor.Links.Href(shiftbellapi.RelationSelf),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrievedSuccessor.StatusCode).To(Equal(http.StatusOK))
			Expect(retrievedSuccessor.SuccessResponse).NotTo(BeNil())
			Expect(retrievedSuccessor.SuccessResponse.Chore).To(Equal(successor))
			deadline := "2020-04-02T00:00:00Z"
			schedule := shiftbellapi.Schedule{
				Id:                  *created.Chore.ScheduleId,
				Name:                scope + " weekly",
				ChoreTemplateId:     template.ChoreTemplate.Id,
				ChoreTemplateName:   scope + " updated",
				IntervalDays:        interval,
				ActiveChoreId:       &successor.Id,
				ActiveChoreDeadline: &deadline,
			}
			Expect(browseAllSchedules(ctx, client)).To(ContainElement(schedule))

			By(
				"repeating completion with a different date without creating another successor",
			)
			repeated, err := client.CompleteChore(
				ctx,
				completionHref,
				shiftbellapi.CompleteChoreParams{CompletedOn: "2020-03-27"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(repeated.StatusCode).To(Equal(http.StatusOK))
			Expect(repeated.ErrorResponse).To(BeNil())
			Expect(repeated.SuccessResponse).To(Equal(completed.SuccessResponse))
			active, err = client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
				Href: "/chores", Search: scope, Status: "active", Limit: 20,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(active.Collection.Items).To(ConsistOf(successor))
			history, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
				Href: "/chores", Search: scope, Status: "completed", Limit: 20,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(history.Collection.Items).To(ConsistOf(expected))
			Expect(browseAllSchedules(ctx, client)).To(ContainElement(schedule))
		},
	)

	It(
		"rejects a duplicate active schedule name without creating another schedule or chore",
		func(ctx SpecContext) {
			baseURL := os.Getenv("SHIFTBELL_BASE_URL")
			Expect(baseURL).NotTo(BeEmpty())
			client := shiftbellapi.NewAPIClient(baseURL)
			scope := uniqueChoreName("Schedule name conflict")
			collection := discoverChoreTemplateCollection(ctx, client)
			template := createChoreTemplate(
				ctx,
				client,
				collection,
				scope+" original",
				"Original chore.",
			)
			otherTemplate := createChoreTemplate(
				ctx,
				client,
				collection,
				scope+" rejected",
				"No chore should be created.",
			)
			form := getTemplateOneOffChoreForm(ctx, client,
				discoverChoreCollection(ctx, client), template.ChoreTemplate)
			interval := 7
			name := scope + " weekly"
			createdResult, err := client.CreateChore(
				ctx,
				form.Actions.Href(shiftbellapi.RelationCreate),
				shiftbellapi.CreateChoreParams{
					ChoreTemplateId: &template.ChoreTemplate.Id,
					ScheduleName:    name,
					IntervalDays:    &interval,
					Deadline:        "2020-02-03",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdResult.StatusCode).To(Equal(http.StatusCreated))
			Expect(createdResult.SuccessResponse).NotTo(BeNil())
			created := createdResult.SuccessResponse
			Expect(created.Chore.ScheduleId).NotTo(BeNil())
			deadline := "2020-02-03T00:00:00Z"
			schedule := shiftbellapi.Schedule{
				Id:                  *created.Chore.ScheduleId,
				Name:                name,
				ChoreTemplateId:     template.ChoreTemplate.Id,
				ChoreTemplateName:   template.ChoreTemplate.Name,
				IntervalDays:        interval,
				ActiveChoreId:       &created.Chore.Id,
				ActiveChoreDeadline: &deadline,
			}
			Expect(browseAllSchedules(ctx, client)).To(ContainElement(schedule))

			By(
				"requesting the same active schedule name with different case and surrounding whitespace",
			)
			conflict, err := client.CreateChore(
				ctx,
				form.Actions.Href(shiftbellapi.RelationCreate),
				shiftbellapi.CreateChoreParams{
					ChoreTemplateId: &otherTemplate.ChoreTemplate.Id,
					ScheduleName:    "  " + strings.ToUpper(name) + "  ",
					IntervalDays:    &interval,
					Deadline:        "2020-03-03",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(conflict.StatusCode).To(Equal(http.StatusConflict))
			Expect(conflict.SuccessResponse).To(BeNil())
			Expect(conflict.ErrorResponse).NotTo(BeNil())
			Expect(
				conflict.ErrorResponse.Error,
			).To(Equal("schedule name conflicts with an active schedule"))
			Expect(conflict.ErrorResponse.Actions).To(ConsistOf(shiftbellapi.Relation{
				Rel:  shiftbellapi.RelationCreate,
				Href: form.Actions.Href(shiftbellapi.RelationCreate),
			}))

			By("verifying no additional schedule or chore was created")
			schedules := browseAllSchedules(ctx, client)
			Expect(schedules).To(ContainElement(schedule))
			Expect(
				schedules,
			).NotTo(ContainElement(HaveField("ChoreTemplateId", otherTemplate.ChoreTemplate.Id)))
			chores, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
				Href: "/chores", Search: scope, Status: "active", Limit: 20,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(chores.Collection.Items).To(ConsistOf(created.Chore))
			retrieved, err := client.GetChore(
				ctx,
				created.Chore.Links.Href(shiftbellapi.RelationSelf),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrieved.StatusCode).To(Equal(http.StatusOK))
			Expect(retrieved.SuccessResponse).NotTo(BeNil())
			Expect(retrieved.SuccessResponse.Chore).To(Equal(created.Chore))
		},
	)

	It(
		"rejects a deactivated template without leaving a schedule or chore",
		func(ctx SpecContext) {
			baseURL := os.Getenv("SHIFTBELL_BASE_URL")
			Expect(baseURL).NotTo(BeEmpty())
			client := shiftbellapi.NewAPIClient(baseURL)
			scope := uniqueChoreName("Inactive template schedule")
			template := createChoreTemplate(
				ctx,
				client,
				discoverChoreTemplateCollection(
					ctx,
					client,
				),
				scope,
				"Inactive scheduled steps.",
			)
			form := getTemplateOneOffChoreForm(ctx, client,
				discoverChoreCollection(ctx, client), template.ChoreTemplate)

			By("deactivating the selected template before submission")
			deactivated, err := client.DeactivateChoreTemplate(
				ctx,
				template.Actions.Href(shiftbellapi.RelationDeactivate),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(deactivated.StatusCode).To(Equal(http.StatusOK))
			Expect(deactivated.ErrorResponse).To(BeNil())
			Expect(deactivated.ChoreTemplate.DeactivatedAt).NotTo(BeNil())
			interval := 7

			result, err := client.CreateChore(
				ctx,
				form.Actions.Href(shiftbellapi.RelationCreate),
				shiftbellapi.CreateChoreParams{
					ChoreTemplateId: &template.ChoreTemplate.Id,
					ScheduleName:    scope,
					IntervalDays:    &interval,
					Deadline:        "2020-02-03",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.StatusCode).To(Equal(http.StatusUnprocessableEntity))
			Expect(result.SuccessResponse).To(BeNil())
			Expect(result.ErrorResponse).NotTo(BeNil())
			Expect(result.ErrorResponse.Error).To(Equal("chore template inactive"))

			By("verifying neither resource was created")
			Expect(
				browseAllSchedules(ctx, client),
			).NotTo(ContainElement(HaveField("Name", scope)))
			chores, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
				Href: "/chores", Search: scope, Status: "active", Limit: 20,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(chores.Collection.Items).To(BeEmpty())
		},
	)
})

func browseAllSchedules(
	ctx context.Context,
	client *shiftbellapi.APIClient,
) []shiftbellapi.Schedule {
	GinkgoHelper()
	home, err := client.GetHome(ctx)
	Expect(err).NotTo(HaveOccurred())
	href := home.Home.Links.Href(shiftbellapi.RelationSchedules)
	Expect(href).NotTo(BeEmpty())
	items := make([]shiftbellapi.Schedule, 0)
	for href != "" {
		collection, err := client.GetSchedules(ctx, href)
		Expect(err).NotTo(HaveOccurred())
		items = append(items, collection.Items...)
		href = collection.Links.Href(shiftbellapi.RelationNext)
	}
	return items
}
