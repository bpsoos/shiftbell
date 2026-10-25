package acceptance_test

import (
	"net/http"
	"os"
	"strconv"
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
		"navigates from the schedule collection to its details, template, and active chore",
		func(ctx SpecContext) {
			baseURL := os.Getenv("SHIFTBELL_BASE_URL")
			Expect(baseURL).NotTo(BeEmpty())
			client := shiftbellapi.NewAPIClient(baseURL)
			scope := uniqueChoreName("Schedule navigation")
			template := createChoreTemplate(ctx, client,
				discoverChoreTemplateCollection(ctx, client), scope, "Reusable steps.")
			form := getTemplateOneOffChoreForm(ctx, client,
				discoverChoreCollection(ctx, client), template.ChoreTemplate)
			interval := 7
			created, err := client.CreateChore(
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
			Expect(created.StatusCode).To(Equal(http.StatusCreated))
			Expect(created.SuccessResponse).NotTo(BeNil())
			chore := created.SuccessResponse.Chore
			Expect(chore.ScheduleId).NotTo(BeNil())

			By("following the schedule link advertised by its collection item")
			var item shiftbellapi.Schedule
			Expect(
				browseAllSchedules(ctx, client),
			).To(ContainElement(HaveField("Id", *chore.ScheduleId), &item))
			href := item.Links.Href(shiftbellapi.RelationSelf)
			Expect(href).NotTo(BeEmpty())
			details, err := client.GetSchedule(ctx, href)
			Expect(err).NotTo(HaveOccurred())
			Expect(details.StatusCode).To(Equal(http.StatusOK))
			Expect(details.ErrorResponse).To(BeNil())
			Expect(details.SuccessResponse).NotTo(BeNil())
			Expect(details.SuccessResponse.Schedule).To(Equal(item))

			By("following the template and active chore links from schedule details")
			templateHref := details.SuccessResponse.Links.Href(
				shiftbellapi.RelationChoreTemplate,
			)
			Expect(templateHref).NotTo(BeEmpty())
			linkedTemplate, err := client.GetChoreTemplate(ctx, templateHref)
			Expect(err).NotTo(HaveOccurred())
			Expect(linkedTemplate.StatusCode).To(Equal(http.StatusOK))
			Expect(linkedTemplate.SuccessResponse).NotTo(BeNil())
			Expect(
				linkedTemplate.SuccessResponse.ChoreTemplate,
			).To(Equal(template.ChoreTemplate))
			choreHref := details.SuccessResponse.Links.Href(
				shiftbellapi.RelationActiveChore,
			)
			Expect(choreHref).NotTo(BeEmpty())
			linkedChore, err := client.GetChore(ctx, choreHref)
			Expect(err).NotTo(HaveOccurred())
			Expect(linkedChore.StatusCode).To(Equal(http.StatusOK))
			Expect(linkedChore.SuccessResponse).NotTo(BeNil())
			Expect(linkedChore.SuccessResponse.Chore).To(Equal(chore))

			By("following the collection link back to the schedule item")
			collectionHref := details.SuccessResponse.Links.Href(
				shiftbellapi.RelationCollection,
			)
			Expect(collectionHref).NotTo(BeEmpty())
			Expect(
				browseAllSchedulesFrom(ctx, client, collectionHref),
			).To(ContainElement(item))
		},
	)

	It("edits a schedule without changing its active chore", func(ctx SpecContext) {
		baseURL := os.Getenv("SHIFTBELL_BASE_URL")
		Expect(baseURL).NotTo(BeEmpty())
		client := shiftbellapi.NewAPIClient(baseURL)
		created := createScheduledChore(ctx, client, uniqueChoreName("Schedule edit"), 7)
		Expect(created.Chore.ScheduleId).NotTo(BeNil())
		var item shiftbellapi.Schedule
		Expect(
			browseAllSchedules(ctx, client),
		).To(ContainElement(HaveField("Id", *created.Chore.ScheduleId), &item))
		details, err := client.GetSchedule(
			ctx,
			item.Links.Href(shiftbellapi.RelationSelf),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(details.StatusCode).To(Equal(http.StatusOK))
		Expect(details.SuccessResponse).NotTo(BeNil())

		By("following the advertised edit action")
		editHref := details.SuccessResponse.Actions.Href(shiftbellapi.RelationEdit)
		Expect(editHref).NotTo(BeEmpty())
		name := item.Name + " updated"
		edited, err := client.EditSchedule(ctx, editHref, shiftbellapi.EditScheduleParams{
			Name: "  " + name + "  ", IntervalDays: 14,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(edited.StatusCode).To(Equal(http.StatusOK))
		Expect(edited.ErrorResponse).To(BeNil())
		expected := *details.SuccessResponse
		expected.Name = name
		expected.IntervalDays = 14
		Expect(edited.SuccessResponse).To(Equal(&expected))

		By("verifying the edit persisted and the active chore is unchanged")
		persisted, err := client.GetSchedule(
			ctx,
			item.Links.Href(shiftbellapi.RelationSelf),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(persisted.StatusCode).To(Equal(http.StatusOK))
		Expect(persisted.SuccessResponse).To(Equal(&expected))
		chore, err := client.GetChore(
			ctx,
			item.Links.Href(shiftbellapi.RelationActiveChore),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(chore.StatusCode).To(Equal(http.StatusOK))
		Expect(chore.SuccessResponse).NotTo(BeNil())
		Expect(chore.SuccessResponse.Chore).To(Equal(created.Chore))
	})

	It(
		"permanently deactivates a schedule while preserving completed history",
		func(ctx SpecContext) {
			baseURL := os.Getenv("SHIFTBELL_BASE_URL")
			Expect(baseURL).NotTo(BeEmpty())
			client := shiftbellapi.NewAPIClient(baseURL)
			created := createScheduledChore(
				ctx,
				client,
				uniqueChoreName("Schedule deactivation"),
				7,
			)
			completed, err := client.CompleteChore(
				ctx,
				created.Actions.Href(shiftbellapi.RelationComplete),
				shiftbellapi.CompleteChoreParams{CompletedOn: "2020-03-26"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(completed.StatusCode).To(Equal(http.StatusOK))
			Expect(completed.SuccessResponse).NotTo(BeNil())
			Expect(created.Chore.ScheduleId).NotTo(BeNil())
			var item shiftbellapi.Schedule
			Expect(
				browseAllSchedules(ctx, client),
			).To(ContainElement(HaveField("Id", *created.Chore.ScheduleId), &item))
			details, err := client.GetSchedule(
				ctx,
				item.Links.Href(shiftbellapi.RelationSelf),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(details.StatusCode).To(Equal(http.StatusOK))
			Expect(details.SuccessResponse).NotTo(BeNil())
			activeHref := details.SuccessResponse.Links.Href(
				shiftbellapi.RelationActiveChore,
			)
			Expect(activeHref).NotTo(BeEmpty())

			By("following the advertised permanent deactivation action")
			deactivationHref := details.SuccessResponse.Actions.Href(
				shiftbellapi.RelationDeactivate,
			)
			Expect(deactivationHref).NotTo(BeEmpty())
			deactivated, err := client.DeactivateSchedule(ctx, deactivationHref)
			Expect(err).NotTo(HaveOccurred())
			Expect(deactivated.StatusCode).To(Equal(http.StatusOK))
			Expect(deactivated.ErrorResponse).To(BeNil())
			Expect(deactivated.SuccessResponse).NotTo(BeNil())
			Expect(deactivated.SuccessResponse.Id).To(Equal(item.Id))
			Expect(deactivated.SuccessResponse.DeactivatedAt).NotTo(BeNil())
			Expect(deactivated.SuccessResponse.ActiveChoreId).To(BeNil())
			Expect(deactivated.SuccessResponse.ActiveChoreDeadline).To(BeNil())
			Expect(deactivated.SuccessResponse.Actions).To(BeEmpty())
			Expect(
				deactivated.SuccessResponse.Links.Href(shiftbellapi.RelationActiveChore),
			).To(BeEmpty())

			By("verifying the schedule is read-only and excluded from active schedules")
			persisted, err := client.GetSchedule(
				ctx,
				item.Links.Href(shiftbellapi.RelationSelf),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(persisted.StatusCode).To(Equal(http.StatusOK))
			Expect(persisted.SuccessResponse).To(Equal(deactivated.SuccessResponse))
			Expect(
				browseAllSchedules(ctx, client),
			).NotTo(ContainElement(HaveField("Id", item.Id)))

			By("verifying the active chore was removed but completed history remains")
			removed, err := client.GetChore(ctx, activeHref)
			Expect(err).NotTo(HaveOccurred())
			Expect(removed.StatusCode).To(Equal(http.StatusNotFound))
			history, err := client.GetChore(
				ctx,
				completed.SuccessResponse.Chore.Links.Href(shiftbellapi.RelationSelf),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(history.StatusCode).To(Equal(http.StatusOK))
			Expect(history.SuccessResponse).NotTo(BeNil())
			Expect(
				history.SuccessResponse.Chore,
			).To(Equal(completed.SuccessResponse.Chore))
		},
	)

	It(
		"returns not found for an unknown schedule with a usable collection link",
		func(ctx SpecContext) {
			baseURL := os.Getenv("SHIFTBELL_BASE_URL")
			Expect(baseURL).NotTo(BeEmpty())
			client := shiftbellapi.NewAPIClient(baseURL)

			result, err := client.GetSchedule(ctx, "/schedules/9223372036854775807")

			Expect(err).NotTo(HaveOccurred())
			Expect(result.StatusCode).To(Equal(http.StatusNotFound))
			Expect(result.SuccessResponse).To(BeNil())
			Expect(result.ErrorResponse).NotTo(BeNil())
			Expect(result.ErrorResponse.Error).To(Equal("schedule not found"))
			href := result.ErrorResponse.Links.Href(shiftbellapi.RelationCollection)
			Expect(href).NotTo(BeEmpty())
			collection, err := client.GetSchedules(ctx, href)
			Expect(err).NotTo(HaveOccurred())
			Expect(collection.Items).NotTo(BeNil())
		},
	)

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
				Links: shiftbellapi.Relations{
					{
						Rel:  shiftbellapi.RelationSelf,
						Href: "/schedules/" + strconv.Itoa(*created.Chore.ScheduleId),
					},
					{Rel: shiftbellapi.RelationCollection, Href: "/schedules"},
					{
						Rel: shiftbellapi.RelationChoreTemplate,
						Href: template.ChoreTemplate.Links.Href(
							shiftbellapi.RelationSelf,
						),
					},
					{
						Rel:  shiftbellapi.RelationActiveChore,
						Href: created.Chore.Links.Href(shiftbellapi.RelationSelf),
					},
				},
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
				Links: shiftbellapi.Relations{
					{
						Rel:  shiftbellapi.RelationSelf,
						Href: "/schedules/" + strconv.Itoa(*created.Chore.ScheduleId),
					},
					{Rel: shiftbellapi.RelationCollection, Href: "/schedules"},
					{
						Rel: shiftbellapi.RelationChoreTemplate,
						Href: template.ChoreTemplate.Links.Href(
							shiftbellapi.RelationSelf,
						),
					},
					{
						Rel:  shiftbellapi.RelationActiveChore,
						Href: successor.Links.Href(shiftbellapi.RelationSelf),
					},
				},
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
				Links: shiftbellapi.Relations{
					{
						Rel:  shiftbellapi.RelationSelf,
						Href: "/schedules/" + strconv.Itoa(*created.Chore.ScheduleId),
					},
					{Rel: shiftbellapi.RelationCollection, Href: "/schedules"},
					{
						Rel: shiftbellapi.RelationChoreTemplate,
						Href: template.ChoreTemplate.Links.Href(
							shiftbellapi.RelationSelf,
						),
					},
					{
						Rel:  shiftbellapi.RelationActiveChore,
						Href: created.Chore.Links.Href(shiftbellapi.RelationSelf),
					},
				},
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

	When("correcting scheduled completion dates", func() {
		var (
			client    *shiftbellapi.APIClient
			completed *shiftbellapi.CompleteChoreResponse
			successor shiftbellapi.Chore
			schedule  shiftbellapi.Schedule
			scope     string
		)

		BeforeEach(func(ctx SpecContext) {
			baseURL := os.Getenv("SHIFTBELL_BASE_URL")
			Expect(baseURL).NotTo(BeEmpty())
			client = shiftbellapi.NewAPIClient(baseURL)
			scope = uniqueChoreName("Scheduled correction")
			template := createChoreTemplate(
				ctx,
				client,
				discoverChoreTemplateCollection(
					ctx,
					client,
				),
				scope,
				"Preserve these snapshots.",
			)
			form := getTemplateOneOffChoreForm(ctx, client,
				discoverChoreCollection(ctx, client), template.ChoreTemplate)
			interval := 7
			created, err := client.CreateChore(
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
			Expect(created.StatusCode).To(Equal(http.StatusCreated))
			Expect(created.SuccessResponse).NotTo(BeNil())
			completion, err := client.CompleteChore(ctx,
				created.SuccessResponse.Actions.Href(shiftbellapi.RelationComplete),
				shiftbellapi.CompleteChoreParams{CompletedOn: "2020-03-26"})
			Expect(err).NotTo(HaveOccurred())
			Expect(completion.StatusCode).To(Equal(http.StatusOK))
			Expect(completion.SuccessResponse).NotTo(BeNil())
			completed = completion.SuccessResponse
			Expect(completed.Chore.ScheduleId).NotTo(BeNil())
			active, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
				Href: "/chores", Search: scope, Status: "active", Limit: 20,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(active.Collection.Items).To(HaveLen(1))
			successor = active.Collection.Items[0]
			Expect(successor.Deadline).To(Equal("2020-04-02"))
			deadline := "2020-04-02T00:00:00Z"
			schedule = shiftbellapi.Schedule{
				Id:                  *completed.Chore.ScheduleId,
				Name:                scope + " weekly",
				ChoreTemplateId:     template.ChoreTemplate.Id,
				ChoreTemplateName:   template.ChoreTemplate.Name,
				IntervalDays:        interval,
				ActiveChoreId:       &successor.Id,
				ActiveChoreDeadline: &deadline,
				Links: shiftbellapi.Relations{
					{
						Rel:  shiftbellapi.RelationSelf,
						Href: "/schedules/" + strconv.Itoa(*completed.Chore.ScheduleId),
					},
					{Rel: shiftbellapi.RelationCollection, Href: "/schedules"},
					{
						Rel: shiftbellapi.RelationChoreTemplate,
						Href: template.ChoreTemplate.Links.Href(
							shiftbellapi.RelationSelf,
						),
					},
					{
						Rel:  shiftbellapi.RelationActiveChore,
						Href: successor.Links.Href(shiftbellapi.RelationSelf),
					},
				},
			}
			Expect(browseAllSchedules(ctx, client)).To(ContainElement(schedule))
		})

		It(
			"recalculates the active successor deadline without replacing it",
			func(ctx SpecContext) {
				By("correcting the predecessor through its advertised action")
				correctedDate := "2020-03-28"
				correctionHref := completed.Actions.Href(
					shiftbellapi.RelationCorrectCompletion,
				)
				Expect(correctionHref).NotTo(BeEmpty())
				corrected, err := client.CorrectChoreCompletion(ctx, correctionHref,
					shiftbellapi.CorrectChoreCompletionParams{CompletedOn: correctedDate})
				Expect(err).NotTo(HaveOccurred())
				Expect(corrected.StatusCode).To(Equal(http.StatusOK))
				Expect(corrected.ErrorResponse).To(BeNil())
				Expect(corrected.SuccessResponse).NotTo(BeNil())
				expectedOriginal := completed.Chore
				expectedOriginal.CompletedOn = &correctedDate
				Expect(corrected.SuccessResponse.Chore).To(Equal(expectedOriginal))
				Expect(corrected.SuccessResponse.Actions).To(ConsistOf(completed.Actions))

				By("verifying only the existing successor's deadline changed")
				expectedSuccessor := successor
				expectedSuccessor.Deadline = "2020-04-04"
				retrieved, err := client.GetChore(
					ctx,
					successor.Links.Href(shiftbellapi.RelationSelf),
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved.StatusCode).To(Equal(http.StatusOK))
				Expect(retrieved.SuccessResponse).NotTo(BeNil())
				Expect(retrieved.SuccessResponse.Chore).To(Equal(expectedSuccessor))
				active, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href: "/chores", Search: scope, Status: "active", Limit: 20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(active.Collection.Items).To(ConsistOf(expectedSuccessor))
				deadline := "2020-04-04T00:00:00Z"
				schedule.ActiveChoreDeadline = &deadline
				Expect(browseAllSchedules(ctx, client)).To(ContainElement(schedule))

				By("repeating the correction without creating or replacing any chore")
				repeated, err := client.CorrectChoreCompletion(ctx, correctionHref,
					shiftbellapi.CorrectChoreCompletionParams{CompletedOn: correctedDate})
				Expect(err).NotTo(HaveOccurred())
				Expect(repeated.StatusCode).To(Equal(http.StatusOK))
				Expect(repeated.SuccessResponse).To(Equal(corrected.SuccessResponse))
				active, err = client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href: "/chores", Search: scope, Status: "active", Limit: 20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(active.Collection.Items).To(ConsistOf(expectedSuccessor))
				history, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href: "/chores", Search: scope, Status: "completed", Limit: 20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(history.Collection.Items).To(ConsistOf(expectedOriginal))
			},
		)

		It(
			"does not cascade a correction after the immediate successor has completed",
			func(ctx SpecContext) {
				By("completing the immediate successor and creating a later active chore")
				details, err := client.GetChore(
					ctx,
					successor.Links.Href(shiftbellapi.RelationSelf),
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(details.SuccessResponse).NotTo(BeNil())
				secondCompletion, err := client.CompleteChore(ctx,
					details.SuccessResponse.Actions.Href(shiftbellapi.RelationComplete),
					shiftbellapi.CompleteChoreParams{CompletedOn: "2020-04-03"})
				Expect(err).NotTo(HaveOccurred())
				Expect(secondCompletion.StatusCode).To(Equal(http.StatusOK))
				Expect(secondCompletion.SuccessResponse).NotTo(BeNil())
				active, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href: "/chores", Search: scope, Status: "active", Limit: 20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(active.Collection.Items).To(HaveLen(1))
				laterChore := active.Collection.Items[0]
				Expect(laterChore.Deadline).To(Equal("2020-04-10"))

				By("correcting the older predecessor without changing subsequent chores")
				correctedDate := "2020-03-28"
				corrected, err := client.CorrectChoreCompletion(ctx,
					completed.Actions.Href(shiftbellapi.RelationCorrectCompletion),
					shiftbellapi.CorrectChoreCompletionParams{CompletedOn: correctedDate})
				Expect(err).NotTo(HaveOccurred())
				Expect(corrected.StatusCode).To(Equal(http.StatusOK))
				Expect(corrected.SuccessResponse).NotTo(BeNil())
				expectedOriginal := completed.Chore
				expectedOriginal.CompletedOn = &correctedDate
				Expect(corrected.SuccessResponse.Chore).To(Equal(expectedOriginal))
				retrieved, err := client.GetChore(
					ctx,
					successor.Links.Href(shiftbellapi.RelationSelf),
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved.SuccessResponse).NotTo(BeNil())
				Expect(
					retrieved.SuccessResponse.Chore,
				).To(Equal(secondCompletion.SuccessResponse.Chore))
				active, err = client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href: "/chores", Search: scope, Status: "active", Limit: 20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(active.Collection.Items).To(ConsistOf(laterChore))
				history, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href: "/chores", Search: scope, Status: "completed", Limit: 20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(
					history.Collection.Items,
				).To(ConsistOf(expectedOriginal, secondCompletion.SuccessResponse.Chore))
				deadline := "2020-04-10T00:00:00Z"
				schedule.ActiveChoreId = &laterChore.Id
				schedule.ActiveChoreDeadline = &deadline
				schedule.Links = shiftbellapi.Relations{
					{
						Rel:  shiftbellapi.RelationSelf,
						Href: schedule.Links.Href(shiftbellapi.RelationSelf),
					},
					{
						Rel:  shiftbellapi.RelationCollection,
						Href: schedule.Links.Href(shiftbellapi.RelationCollection),
					},
					{
						Rel:  shiftbellapi.RelationChoreTemplate,
						Href: schedule.Links.Href(shiftbellapi.RelationChoreTemplate),
					},
					{
						Rel:  shiftbellapi.RelationActiveChore,
						Href: laterChore.Links.Href(shiftbellapi.RelationSelf),
					},
				}
				Expect(browseAllSchedules(ctx, client)).To(ContainElement(schedule))
			},
		)
	})

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
