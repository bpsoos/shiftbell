package acceptance_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bpsoos/shiftbell/internal/testsupport/shiftbellapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
)

type prepareOneOffChoreForDeletion func(
	context.Context,
	*shiftbellapi.APIClient,
	*shiftbellapi.CreateChoreResponse,
) (shiftbellapi.Chore, shiftbellapi.Actions)

var _ = Describe("Chore API", func() {
	var client *shiftbellapi.APIClient

	BeforeEach(func() {
		baseURL := os.Getenv("SHIFTBELL_BASE_URL")
		Expect(baseURL).NotTo(BeEmpty())
		client = shiftbellapi.NewAPIClient(baseURL)
	})

	When("creating a one-off chore", func() {
		It("creates a manual one-off chore that can be retrieved", func(ctx SpecContext) {
			By("discovering the chore collection from home")
			collection := discoverChoreCollection(ctx, client)
			Expect(collection.Items).NotTo(BeNil())
			Expect(collection.Links).To(gstruct.MatchAllKeys(gstruct.Keys{
				shiftbellapi.RelationSelf: gstruct.MatchAllFields(gstruct.Fields{
					"Href": Equal("/chores"),
				}),
			}))
			Expect(collection.Actions).To(gstruct.MatchAllKeys(gstruct.Keys{
				shiftbellapi.ActionCreateChore: gstruct.MatchAllFields(gstruct.Fields{
					"Href":        Equal("/chores/new"),
					"Method":      Equal(http.MethodGet),
					"ContentType": BeEmpty(),
					"Fields":      BeEmpty(),
				}),
			}))

			form := getManualOneOffChoreForm(ctx, client, collection)

			By("creating a manual one-off chore")
			name := uniqueChoreName("Manual one-off")
			result, err := client.CreateChore(
				ctx,
				shiftbellapi.RequestParams{
					Method:      form.Action.Method,
					Href:        form.Action.Href,
					ContentType: form.Action.ContentType,
				},
				shiftbellapi.CreateChoreParams{
					Name:        "  " + name + "  ",
					Description: "  Wash and fold.  ",
					Deadline:    "2020-02-03",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.StatusCode).To(Equal(http.StatusCreated))
			Expect(result.ErrorResponse).To(BeNil())
			Expect(result.SuccessResponse).NotTo(BeNil())
			created := result.SuccessResponse
			Expect(created.Location).NotTo(BeEmpty())
			Expect(created.Chore).To(gstruct.MatchAllFields(gstruct.Fields{
				"Id":          BeNumerically(">", 0),
				"ScheduleId":  BeNil(),
				"Status":      Equal("active"),
				"Name":        Equal(name),
				"Description": Equal("Wash and fold."),
				"Deadline":    Equal("2020-02-03"),
				"CompletedOn": BeNil(),
				"Links": gstruct.MatchAllKeys(gstruct.Keys{
					shiftbellapi.RelationSelf: gstruct.MatchAllFields(gstruct.Fields{
						"Href": Equal(created.Location),
					}),
					shiftbellapi.RelationCollection: gstruct.MatchAllFields(
						gstruct.Fields{
							"Href": Equal("/chores"),
						},
					),
				}),
			}))
			Expect(created.Actions).To(gstruct.MatchAllKeys(gstruct.Keys{
				shiftbellapi.ActionEditChore: gstruct.MatchAllFields(gstruct.Fields{
					"Href":        Equal(created.Location),
					"Method":      Equal(http.MethodPatch),
					"ContentType": Equal("application/json"),
					"Fields": Equal([]shiftbellapi.ActionField{
						{Name: "name", Type: "string", Required: true},
						{
							Name:     "description",
							Type:     "string",
							Required: false,
						},
						{Name: "deadline", Type: "date", Required: true},
					}),
				}),
				shiftbellapi.ActionCompleteChore: gstruct.MatchAllFields(gstruct.Fields{
					"Href":        Equal(created.Location + "/completion"),
					"Method":      Equal(http.MethodPut),
					"ContentType": Equal("application/json"),
					"Fields": Equal([]shiftbellapi.ActionField{
						{
							Name:     "completed_on",
							Type:     "date",
							Required: true,
						},
					}),
				}),
				shiftbellapi.ActionDeleteChore: gstruct.MatchAllFields(gstruct.Fields{
					"Href":        Equal(created.Location),
					"Method":      Equal(http.MethodDelete),
					"ContentType": BeEmpty(),
					"Fields":      BeEmpty(),
				}),
			}))

			By("retrieving the created chore")
			retrieved, err := client.GetChore(ctx, shiftbellapi.GetChoreParams{
				Link: created.Chore.Links[shiftbellapi.RelationSelf],
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(retrieved.StatusCode).To(Equal(http.StatusOK))
			Expect(retrieved.ErrorResponse).To(BeNil())
			Expect(retrieved.SuccessResponse).NotTo(BeNil())
			Expect(retrieved.SuccessResponse.Chore).To(Equal(created.Chore))
			Expect(retrieved.SuccessResponse.Actions).To(Equal(created.Actions))
		})

		It(
			"creates a manual one-off chore and saves it as a chore template",
			func(ctx SpecContext) {
				collection := discoverChoreCollection(ctx, client)
				form := getManualOneOffChoreForm(ctx, client, collection)
				name := uniqueChoreName("Saved manual one-off")

				By("creating the chore with save-as-template enabled")
				result, err := client.CreateChore(
					ctx,
					shiftbellapi.RequestParams{
						Method:      form.Action.Method,
						Href:        form.Action.Href,
						ContentType: form.Action.ContentType,
					},
					shiftbellapi.CreateChoreParams{
						Name:                "  " + name + "  ",
						Description:         "  Reusable kitchen steps.  ",
						Deadline:            "2020-02-04",
						SaveAsChoreTemplate: true,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.StatusCode).To(Equal(http.StatusCreated))
				Expect(result.ErrorResponse).To(BeNil())
				Expect(result.SuccessResponse).NotTo(BeNil())
				created := result.SuccessResponse
				Expect(created.Location).NotTo(BeEmpty())
				Expect(created.Chore).To(gstruct.MatchAllFields(gstruct.Fields{
					"Id":          BeNumerically(">", 0),
					"ScheduleId":  BeNil(),
					"Status":      Equal("active"),
					"Name":        Equal(name),
					"Description": Equal("Reusable kitchen steps."),
					"Deadline":    Equal("2020-02-04"),
					"CompletedOn": BeNil(),
					"Links": gstruct.MatchAllKeys(gstruct.Keys{
						shiftbellapi.RelationSelf: gstruct.MatchAllFields(gstruct.Fields{
							"Href": Equal(created.Location),
						}),
						shiftbellapi.RelationCollection: gstruct.MatchAllFields(
							gstruct.Fields{"Href": Equal("/chores")},
						),
					}),
				}))
				Expect(created.Actions).NotTo(BeEmpty())

				By("browsing the uniquely scoped active chore collection")
				chores, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: name,
					Status: "active",
					Limit:  20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(chores.Collection.Items).To(gstruct.MatchAllElementsWithIndex(
					gstruct.IndexIdentity,
					gstruct.Elements{"0": Equal(created.Chore)},
				))

				By("browsing the newly saved active chore template")
				templateCollection := discoverChoreTemplateCollection(ctx, client)
				templates, err := client.BrowseChoreTemplates(
					ctx,
					shiftbellapi.BrowseChoreTemplatesParams{
						Href:   templateCollection.Links[shiftbellapi.RelationSelf].Href,
						Search: name,
						State:  "active",
						Limit:  20,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(templates.Collection.Items).To(gstruct.MatchAllElementsWithIndex(
					gstruct.IndexIdentity,
					gstruct.Elements{
						"0": gstruct.MatchAllFields(gstruct.Fields{
							"Id":            BeNumerically(">", 0),
							"Name":          Equal(name),
							"Description":   Equal("Reusable kitchen steps."),
							"DeactivatedAt": BeNil(),
							"Links": gstruct.MatchAllKeys(gstruct.Keys{
								shiftbellapi.RelationSelf: gstruct.MatchAllFields(
									gstruct.Fields{"Href": Not(BeEmpty())},
								),
								shiftbellapi.RelationCollection: gstruct.MatchAllFields(
									gstruct.Fields{"Href": Equal("/chore-templates")},
								),
							}),
						}),
					},
				))
			},
		)

		It("creates a template-based one-off chore", func(ctx SpecContext) {
			templateCollection := discoverChoreTemplateCollection(ctx, client)
			template := createChoreTemplate(
				ctx,
				client,
				templateCollection,
				uniqueChoreName("Template-based one-off"),
				"Reusable template steps.",
			).ChoreTemplate
			collection := discoverChoreCollection(ctx, client)
			form := getTemplateOneOffChoreForm(ctx, client, collection, template)

			By("creating a one-off chore from the selected template")
			templateId := template.Id
			result, err := client.CreateChore(
				ctx,
				shiftbellapi.RequestParams{
					Method:      form.Action.Method,
					Href:        form.Action.Href,
					ContentType: form.Action.ContentType,
				},
				shiftbellapi.CreateChoreParams{
					ChoreTemplateId: &templateId,
					Deadline:        "2020-02-05",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.StatusCode).To(Equal(http.StatusCreated))
			Expect(result.ErrorResponse).To(BeNil())
			Expect(result.SuccessResponse).NotTo(BeNil())
			created := result.SuccessResponse
			Expect(created.Location).NotTo(BeEmpty())
			Expect(created.Chore).To(gstruct.MatchAllFields(gstruct.Fields{
				"Id":          BeNumerically(">", 0),
				"ScheduleId":  BeNil(),
				"Status":      Equal("active"),
				"Name":        Equal(template.Name),
				"Description": Equal(template.Description),
				"Deadline":    Equal("2020-02-05"),
				"CompletedOn": BeNil(),
				"Links": gstruct.MatchAllKeys(gstruct.Keys{
					shiftbellapi.RelationSelf: gstruct.MatchAllFields(gstruct.Fields{
						"Href": Equal(created.Location),
					}),
					shiftbellapi.RelationCollection: gstruct.MatchAllFields(
						gstruct.Fields{"Href": Equal("/chores")},
					),
				}),
			}))
			Expect(created.Actions).NotTo(BeEmpty())

			By("browsing the uniquely scoped active chore collection")
			chores, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
				Href:   collection.Links[shiftbellapi.RelationSelf].Href,
				Search: template.Name,
				Status: "active",
				Limit:  20,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(chores.Collection.Items).To(gstruct.MatchAllElementsWithIndex(
				gstruct.IndexIdentity,
				gstruct.Elements{"0": Equal(created.Chore)},
			))

			By("retrieving the created chore")
			retrieved, err := client.GetChore(ctx, shiftbellapi.GetChoreParams{
				Link: created.Chore.Links[shiftbellapi.RelationSelf],
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(retrieved.StatusCode).To(Equal(http.StatusOK))
			Expect(retrieved.ErrorResponse).To(BeNil())
			Expect(retrieved.SuccessResponse).NotTo(BeNil())
			Expect(retrieved.SuccessResponse.Chore).To(Equal(created.Chore))
			Expect(retrieved.SuccessResponse.Actions).To(Equal(created.Actions))
		})

		It(
			"does not create a chore when save-as-template conflicts with an active template",
			func(ctx SpecContext) {
				templateCollection := discoverChoreTemplateCollection(ctx, client)
				name := uniqueChoreName("Save conflict")
				existing := createChoreTemplate(
					ctx,
					client,
					templateCollection,
					name,
					"Existing template description",
				)
				collection := discoverChoreCollection(ctx, client)
				form := getManualOneOffChoreForm(ctx, client, collection)

				By("attempting to save the chore under the active template name")
				conflict, err := client.CreateChore(
					ctx,
					shiftbellapi.RequestParams{
						Method:      form.Action.Method,
						Href:        form.Action.Href,
						ContentType: form.Action.ContentType,
					},
					shiftbellapi.CreateChoreParams{
						Name:                strings.ToUpper(name),
						Description:         "Conflicting chore description",
						Deadline:            "2020-02-06",
						SaveAsChoreTemplate: true,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(conflict.StatusCode).To(Equal(http.StatusConflict))
				Expect(conflict.SuccessResponse).To(BeNil())
				Expect(conflict.ErrorResponse).NotTo(BeNil())
				Expect(conflict.ErrorResponse.Error).To(Equal(
					"chore template name conflicts with an active chore template",
				))
				Expect(conflict.ErrorResponse.Links).To(BeEmpty())
				Expect(
					conflict.ErrorResponse.Actions,
				).To(gstruct.MatchAllKeys(gstruct.Keys{
					shiftbellapi.ActionCreateChore: Equal(*form.Action),
				}))

				By("proving no chore was created")
				chores, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: name,
					Status: "active",
					Limit:  20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(chores.Collection.Items).To(BeEmpty())

				By("proving the original template is unchanged")
				templates, err := client.BrowseChoreTemplates(
					ctx,
					shiftbellapi.BrowseChoreTemplatesParams{
						Href:   templateCollection.Links[shiftbellapi.RelationSelf].Href,
						Search: name,
						State:  "active",
						Limit:  20,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(templates.Collection.Items).To(gstruct.MatchAllElementsWithIndex(
					gstruct.IndexIdentity,
					gstruct.Elements{"0": Equal(existing.ChoreTemplate)},
				))
			},
		)

		It("does not create a chore from a deactivated template", func(ctx SpecContext) {
			templateCollection := discoverChoreTemplateCollection(ctx, client)
			template := createChoreTemplate(
				ctx,
				client,
				templateCollection,
				uniqueChoreName("Deactivated source"),
				"Reusable template steps.",
			)
			selfHref := template.ChoreTemplate.Links[shiftbellapi.RelationSelf].Href
			expectActiveChoreTemplateActions(template.Actions, selfHref)
			collection := discoverChoreCollection(ctx, client)
			form := getTemplateOneOffChoreForm(
				ctx,
				client,
				collection,
				template.ChoreTemplate,
			)

			By("deactivating the selected chore template")
			deactivateAction := template.Actions[shiftbellapi.ActionDeactivateTemplate]
			deactivated, err := client.DeactivateChoreTemplate(
				ctx,
				shiftbellapi.RequestParams{
					Method:      deactivateAction.Method,
					Href:        deactivateAction.Href,
					ContentType: deactivateAction.ContentType,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(deactivated.ChoreTemplate.DeactivatedAt).NotTo(BeNil())
			Expect(deactivated.Actions).To(BeEmpty())

			By("submitting the previously advertised template-based form")
			templateId := template.ChoreTemplate.Id
			rejected, err := client.CreateChore(
				ctx,
				shiftbellapi.RequestParams{
					Method:      form.Action.Method,
					Href:        form.Action.Href,
					ContentType: form.Action.ContentType,
				},
				shiftbellapi.CreateChoreParams{
					ChoreTemplateId: &templateId,
					Deadline:        "2020-02-07",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(rejected.StatusCode).To(Equal(http.StatusUnprocessableEntity))
			Expect(rejected.SuccessResponse).To(BeNil())
			Expect(rejected.ErrorResponse).NotTo(BeNil())
			Expect(rejected.ErrorResponse.Error).To(Equal("chore template inactive"))
			Expect(rejected.ErrorResponse.Links).To(gstruct.MatchAllKeys(gstruct.Keys{
				shiftbellapi.RelationCollection: gstruct.MatchAllFields(gstruct.Fields{
					"Href": Equal(collection.Links[shiftbellapi.RelationSelf].Href),
				}),
			}))
			Expect(rejected.ErrorResponse.Actions).To(gstruct.MatchAllKeys(gstruct.Keys{
				shiftbellapi.ActionCreateChore: Equal(*form.Action),
			}))

			By("proving no chore was created")
			chores, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
				Href:   collection.Links[shiftbellapi.RelationSelf].Href,
				Search: template.ChoreTemplate.Name,
				Status: "active",
				Limit:  20,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(chores.Collection.Items).To(BeEmpty())
		})
	})

	When("scheduled recurrence is requested", func() {
		It(
			"returns Not Implemented without persisting a manual scheduled chore",
			func(ctx SpecContext) {
				collection := discoverChoreCollection(ctx, client)
				templateCollection := discoverChoreTemplateCollection(ctx, client)
				scope := uniqueChoreName("Manual scheduled recurrence")
				recurrence := getManualChoreRecurrence(ctx, client, collection)

				By("choosing scheduled recurrence")
				result, err := client.GetChoreCreationStep(
					ctx,
					recurrence.Choices[1].Href,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.StatusCode).To(Equal(http.StatusNotImplemented))
				Expect(result.SuccessResponse).To(BeNil())
				Expect(result.ErrorResponse).NotTo(BeNil())
				Expect(result.ErrorResponse.Error).NotTo(BeEmpty())

				By("proving no chore was created")
				chores, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: scope,
					Status: "active",
					Limit:  20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(chores.Collection.Items).To(BeEmpty())

				By("proving no chore template was created")
				templates, err := client.BrowseChoreTemplates(
					ctx,
					shiftbellapi.BrowseChoreTemplatesParams{
						Href:   templateCollection.Links[shiftbellapi.RelationSelf].Href,
						Search: scope,
						State:  "active",
						Limit:  20,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(templates.Collection.Items).To(BeEmpty())
			},
		)

		It(
			"returns Not Implemented without changing a template-based scheduled chore source",
			func(ctx SpecContext) {
				templateCollection := discoverChoreTemplateCollection(ctx, client)
				scope := uniqueChoreName("Template scheduled recurrence")
				template := createChoreTemplate(
					ctx,
					client,
					templateCollection,
					scope,
					"Reusable scheduled steps.",
				).ChoreTemplate
				collection := discoverChoreCollection(ctx, client)
				recurrence := getTemplateChoreRecurrence(
					ctx,
					client,
					collection,
					template,
				)

				By("choosing scheduled recurrence")
				result, err := client.GetChoreCreationStep(
					ctx,
					recurrence.Choices[1].Href,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.StatusCode).To(Equal(http.StatusNotImplemented))
				Expect(result.SuccessResponse).To(BeNil())
				Expect(result.ErrorResponse).NotTo(BeNil())
				Expect(result.ErrorResponse.Error).NotTo(BeEmpty())

				By("proving no chore was created")
				chores, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: scope,
					Status: "active",
					Limit:  20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(chores.Collection.Items).To(BeEmpty())

				By("proving the selected chore template is unchanged")
				templates, err := client.BrowseChoreTemplates(
					ctx,
					shiftbellapi.BrowseChoreTemplatesParams{
						Href:   templateCollection.Links[shiftbellapi.RelationSelf].Href,
						Search: scope,
						State:  "active",
						Limit:  20,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(templates.Collection.Items).To(gstruct.MatchAllElementsWithIndex(
					gstruct.IndexIdentity,
					gstruct.Elements{"0": Equal(template)},
				))
			},
		)
	})

	When("browsing chores", func() {
		It(
			"searches active chores ordered by deadline and ID ascending",
			func(ctx SpecContext) {
				collection := discoverChoreCollection(ctx, client)
				form := getManualOneOffChoreForm(ctx, client, collection)
				scope := uniqueChoreName("Browse active")
				later := createManualOneOffChore(
					ctx,
					client,
					form,
					scope+" Later",
					"Matches by name.",
					"2020-03-03",
				)
				earlyFirst := createManualOneOffChore(
					ctx,
					client,
					form,
					scope+" Early first",
					"Matches by name.",
					"2020-03-01",
				)
				earlySecond := createManualOneOffChore(
					ctx,
					client,
					form,
					uniqueChoreName("Active description match"),
					scope+" matches by description.",
					"2020-03-01",
				)
				createManualOneOffChore(
					ctx,
					client,
					form,
					uniqueChoreName("Unrelated active"),
					"Must not match the scoped search.",
					"2020-03-01",
				)
				completed := createManualOneOffChore(
					ctx,
					client,
					form,
					scope+" Completed",
					"Must be excluded by status.",
					"2020-03-01",
				)
				completeOneOffChore(ctx, client, completed, "2020-03-02")

				By("searching the active collection case-insensitively")
				page, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: strings.ToUpper(scope),
					Status: "active",
					Limit:  20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(page.Collection.More).To(BeFalse())
				Expect(page.Collection.Items).To(gstruct.MatchAllElementsWithIndex(
					gstruct.IndexIdentity,
					gstruct.Elements{
						"0": Equal(earlyFirst.Chore),
						"1": Equal(earlySecond.Chore),
						"2": Equal(later.Chore),
					},
				))
				Expect(page.Collection.Links).To(gstruct.MatchAllKeys(gstruct.Keys{
					shiftbellapi.RelationSelf: gstruct.MatchAllFields(gstruct.Fields{
						"Href": Not(BeEmpty()),
					}),
				}))
				Expect(page.Collection.Actions).To(gstruct.MatchAllKeys(gstruct.Keys{
					shiftbellapi.ActionCreateChore: Equal(
						collection.Actions[shiftbellapi.ActionCreateChore],
					),
				}))
			},
		)

		It(
			"searches completed chores ordered by completion date and ID descending",
			func(ctx SpecContext) {
				collection := discoverChoreCollection(ctx, client)
				form := getManualOneOffChoreForm(ctx, client, collection)
				scope := uniqueChoreName("Browse completed")
				createManualOneOffChore(
					ctx,
					client,
					form,
					scope+" Active",
					"Must be excluded by status.",
					"2020-04-01",
				)
				oldest := createManualOneOffChore(
					ctx,
					client,
					form,
					scope+" Oldest",
					"Matches by name.",
					"2020-04-01",
				)
				sameDateFirst := createManualOneOffChore(
					ctx,
					client,
					form,
					scope+" Same date first",
					"Matches by name.",
					"2020-04-02",
				)
				sameDateSecond := createManualOneOffChore(
					ctx,
					client,
					form,
					uniqueChoreName("Completed description match"),
					scope+" matches by description.",
					"2020-04-03",
				)
				unrelated := createManualOneOffChore(
					ctx,
					client,
					form,
					uniqueChoreName("Unrelated completed"),
					"Must not match the scoped search.",
					"2020-04-01",
				)
				completedOldest := completeOneOffChore(
					ctx,
					client,
					oldest,
					"2020-04-01",
				)
				completedSameDateFirst := completeOneOffChore(
					ctx,
					client,
					sameDateFirst,
					"2020-04-02",
				)
				completedSameDateSecond := completeOneOffChore(
					ctx,
					client,
					sameDateSecond,
					"2020-04-02",
				)
				completeOneOffChore(ctx, client, unrelated, "2020-04-03")

				By("searching completed history case-insensitively")
				page, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: strings.ToUpper(scope),
					Status: "completed",
					Limit:  20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(page.Collection.More).To(BeFalse())
				Expect(page.Collection.Items).To(gstruct.MatchAllElementsWithIndex(
					gstruct.IndexIdentity,
					gstruct.Elements{
						"0": Equal(completedSameDateSecond.Chore),
						"1": Equal(completedSameDateFirst.Chore),
						"2": Equal(completedOldest.Chore),
					},
				))
				Expect(page.Collection.Links).To(gstruct.MatchAllKeys(gstruct.Keys{
					shiftbellapi.RelationSelf: gstruct.MatchAllFields(gstruct.Fields{
						"Href": Not(BeEmpty()),
					}),
				}))
				Expect(page.Collection.Actions).To(gstruct.MatchAllKeys(gstruct.Keys{
					shiftbellapi.ActionCreateChore: Equal(
						collection.Actions[shiftbellapi.ActionCreateChore],
					),
				}))
			},
		)
	})

	When("editing an active one-off chore", func() {
		It(
			"updates its normalized name, description, and deadline",
			func(ctx SpecContext) {
				collection := discoverChoreCollection(ctx, client)
				form := getManualOneOffChoreForm(ctx, client, collection)
				scope := uniqueChoreName("Edit active one-off")
				created := createManualOneOffChore(
					ctx,
					client,
					form,
					scope+" Original",
					"Original description.",
					"2020-05-01",
				)
				editAction := created.Actions[shiftbellapi.ActionEditChore]

				By("editing through the advertised action")
				result, err := client.EditChore(
					ctx,
					shiftbellapi.RequestParams{
						Method:      editAction.Method,
						Href:        editAction.Href,
						ContentType: editAction.ContentType,
					},
					shiftbellapi.EditChoreParams{
						Name:        "  " + scope + " Cafe\u0301  ",
						Description: "  First line\r\nCafe\u0301  ",
						Deadline:    "2020-05-03",
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.StatusCode).To(Equal(http.StatusOK))
				Expect(result.ErrorResponse).To(BeNil())
				Expect(result.SuccessResponse).NotTo(BeNil())
				edited := result.SuccessResponse
				Expect(edited.Chore).To(gstruct.MatchAllFields(gstruct.Fields{
					"Id":          Equal(created.Chore.Id),
					"ScheduleId":  BeNil(),
					"Status":      Equal("active"),
					"Name":        Equal(scope + " Caf\u00e9"),
					"Description": Equal("First line\nCaf\u00e9"),
					"Deadline":    Equal("2020-05-03"),
					"CompletedOn": BeNil(),
					"Links":       Equal(created.Chore.Links),
				}))
				Expect(edited.Actions).To(Equal(created.Actions))

				By("retrieving the edited chore")
				retrieved, err := client.GetChore(ctx, shiftbellapi.GetChoreParams{
					Link: edited.Chore.Links[shiftbellapi.RelationSelf],
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved.StatusCode).To(Equal(http.StatusOK))
				Expect(retrieved.ErrorResponse).To(BeNil())
				Expect(retrieved.SuccessResponse).NotTo(BeNil())
				Expect(retrieved.SuccessResponse.Chore).To(Equal(edited.Chore))
				Expect(retrieved.SuccessResponse.Actions).To(Equal(edited.Actions))
			},
		)
	})

	When("completing an active one-off chore", func() {
		It(
			"moves the chore from the active collection to completed history",
			func(ctx SpecContext) {
				collection := discoverChoreCollection(ctx, client)
				form := getManualOneOffChoreForm(ctx, client, collection)
				scope := uniqueChoreName("Complete one-off")
				created := createManualOneOffChore(
					ctx,
					client,
					form,
					scope,
					"Complete this chore.",
					"2020-06-01",
				)

				By("proving the chore starts in the active collection")
				active, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: scope,
					Status: "active",
					Limit:  20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(active.Collection.Items).To(gstruct.MatchAllElementsWithIndex(
					gstruct.IndexIdentity,
					gstruct.Elements{"0": Equal(created.Chore)},
				))

				By("completing through the advertised action")
				completed := completeOneOffChore(ctx, client, created, "2020-06-02")
				expectCompletedOneOffActions(
					completed.Actions,
					created.Chore.Links[shiftbellapi.RelationSelf].Href,
				)

				By("proving the chore moved between status collections")
				active, err = client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: scope,
					Status: "active",
					Limit:  20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(active.Collection.Items).To(BeEmpty())
				history, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: scope,
					Status: "completed",
					Limit:  20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(history.Collection.Items).To(gstruct.MatchAllElementsWithIndex(
					gstruct.IndexIdentity,
					gstruct.Elements{"0": Equal(completed.Chore)},
				))
			},
		)

		It(
			"treats repeated completion requests as successful no-ops",
			func(ctx SpecContext) {
				collection := discoverChoreCollection(ctx, client)
				form := getManualOneOffChoreForm(ctx, client, collection)
				scope := uniqueChoreName("Repeat completion")
				created := createManualOneOffChore(
					ctx,
					client,
					form,
					scope,
					"Complete repeatedly.",
					"2020-06-01",
				)

				By("reusing the captured completion action")
				action := created.Actions[shiftbellapi.ActionCompleteChore]
				requestParams := shiftbellapi.RequestParams{
					Method:      action.Method,
					Href:        action.Href,
					ContentType: action.ContentType,
				}
				first := completeOneOffChore(ctx, client, created, "2020-06-02")
				secondResult, err := client.CompleteChore(
					ctx,
					requestParams,
					shiftbellapi.CompleteChoreParams{CompletedOn: "2020-06-03"},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(secondResult.StatusCode).To(Equal(http.StatusOK))
				Expect(secondResult.ErrorResponse).To(BeNil())
				Expect(secondResult.SuccessResponse).To(Equal(first))
				thirdResult, err := client.CompleteChore(
					ctx,
					requestParams,
					shiftbellapi.CompleteChoreParams{CompletedOn: "2020-06-04"},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(thirdResult.StatusCode).To(Equal(http.StatusOK))
				Expect(thirdResult.ErrorResponse).To(BeNil())
				Expect(thirdResult.SuccessResponse).To(Equal(first))
				expectCompletedOneOffActions(
					first.Actions,
					created.Chore.Links[shiftbellapi.RelationSelf].Href,
				)

				By("retrieving the stable completed chore")
				retrieved, err := client.GetChore(ctx, shiftbellapi.GetChoreParams{
					Link: created.Chore.Links[shiftbellapi.RelationSelf],
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved.StatusCode).To(Equal(http.StatusOK))
				Expect(retrieved.ErrorResponse).To(BeNil())
				Expect(retrieved.SuccessResponse).NotTo(BeNil())
				Expect(retrieved.SuccessResponse.Chore).To(Equal(first.Chore))
				Expect(retrieved.SuccessResponse.Actions).To(Equal(first.Actions))

				By("proving completed history contains one stable resource")
				history, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: scope,
					Status: "completed",
					Limit:  20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(history.Collection.Items).To(gstruct.MatchAllElementsWithIndex(
					gstruct.IndexIdentity,
					gstruct.Elements{"0": Equal(first.Chore)},
				))
			},
		)

		It(
			"rejects a completion date after the application-local current date",
			func(ctx SpecContext) {
				collection := discoverChoreCollection(ctx, client)
				form := getManualOneOffChoreForm(ctx, client, collection)
				scope := uniqueChoreName("Future completion")
				created := createManualOneOffChore(
					ctx,
					client,
					form,
					scope,
					"Remain active.",
					"2020-06-01",
				)
				action := created.Actions[shiftbellapi.ActionCompleteChore]
				futureDate := time.Now().UTC().AddDate(0, 0, 2).Format(time.DateOnly)

				By("submitting a future completion date")
				result, err := client.CompleteChore(
					ctx,
					shiftbellapi.RequestParams{
						Method:      action.Method,
						Href:        action.Href,
						ContentType: action.ContentType,
					},
					shiftbellapi.CompleteChoreParams{CompletedOn: futureDate},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.StatusCode).To(Equal(http.StatusUnprocessableEntity))
				Expect(result.SuccessResponse).To(BeNil())
				Expect(result.ErrorResponse).NotTo(BeNil())
				Expect(result.ErrorResponse.Error).To(Equal("invalid completion date"))

				By("proving the chore remains active")
				retrieved, err := client.GetChore(ctx, shiftbellapi.GetChoreParams{
					Link: created.Chore.Links[shiftbellapi.RelationSelf],
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved.StatusCode).To(Equal(http.StatusOK))
				Expect(retrieved.ErrorResponse).To(BeNil())
				Expect(retrieved.SuccessResponse).NotTo(BeNil())
				Expect(retrieved.SuccessResponse.Chore).To(Equal(created.Chore))
				Expect(retrieved.SuccessResponse.Actions).To(Equal(created.Actions))
				active, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: scope,
					Status: "active",
					Limit:  20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(active.Collection.Items).To(gstruct.MatchAllElementsWithIndex(
					gstruct.IndexIdentity,
					gstruct.Elements{"0": Equal(created.Chore)},
				))
				history, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: scope,
					Status: "completed",
					Limit:  20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(history.Collection.Items).To(BeEmpty())
			},
		)
	})

	When("correcting a one-off chore completion", func() {
		It("updates its completion date", func(ctx SpecContext) {
			collection := discoverChoreCollection(ctx, client)
			form := getManualOneOffChoreForm(ctx, client, collection)
			scope := uniqueChoreName("Correct completion")
			created := createManualOneOffChore(
				ctx,
				client,
				form,
				scope,
				"Correct this completion.",
				"2020-07-01",
			)
			completed := completeOneOffChore(ctx, client, created, "2020-07-02")

			By("retrieving the completed chore and its correction action")
			beforeCorrection, err := client.GetChore(ctx, shiftbellapi.GetChoreParams{
				Link: completed.Chore.Links[shiftbellapi.RelationSelf],
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(beforeCorrection.StatusCode).To(Equal(http.StatusOK))
			Expect(beforeCorrection.ErrorResponse).To(BeNil())
			Expect(beforeCorrection.SuccessResponse).NotTo(BeNil())
			Expect(beforeCorrection.SuccessResponse.Chore).To(Equal(completed.Chore))
			Expect(beforeCorrection.SuccessResponse.Actions).To(Equal(completed.Actions))
			action := beforeCorrection.SuccessResponse.Actions[shiftbellapi.ActionCorrectCompletion]

			By("correcting the completion date through the advertised action")
			result, err := client.CorrectChoreCompletion(
				ctx,
				shiftbellapi.RequestParams{
					Method:      action.Method,
					Href:        action.Href,
					ContentType: action.ContentType,
				},
				shiftbellapi.CorrectChoreCompletionParams{CompletedOn: "2020-07-03"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.StatusCode).To(Equal(http.StatusOK))
			Expect(result.ErrorResponse).To(BeNil())
			Expect(result.SuccessResponse).NotTo(BeNil())
			corrected := result.SuccessResponse
			correctedOn := "2020-07-03"
			Expect(corrected.Chore).To(gstruct.MatchAllFields(gstruct.Fields{
				"Id":          Equal(completed.Chore.Id),
				"ScheduleId":  Equal(completed.Chore.ScheduleId),
				"Status":      Equal(completed.Chore.Status),
				"Name":        Equal(completed.Chore.Name),
				"Description": Equal(completed.Chore.Description),
				"Deadline":    Equal(completed.Chore.Deadline),
				"CompletedOn": Equal(&correctedOn),
				"Links":       Equal(completed.Chore.Links),
			}))
			Expect(corrected.Actions).To(Equal(completed.Actions))

			By("retrieving the corrected completion date")
			retrieved, err := client.GetChore(ctx, shiftbellapi.GetChoreParams{
				Link: corrected.Chore.Links[shiftbellapi.RelationSelf],
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(retrieved.StatusCode).To(Equal(http.StatusOK))
			Expect(retrieved.ErrorResponse).To(BeNil())
			Expect(retrieved.SuccessResponse).NotTo(BeNil())
			Expect(retrieved.SuccessResponse.Chore).To(Equal(corrected.Chore))
			Expect(retrieved.SuccessResponse.Actions).To(Equal(corrected.Actions))
		})

		It(
			"rejects a completion date after the application-local current date",
			func(ctx SpecContext) {
				collection := discoverChoreCollection(ctx, client)
				form := getManualOneOffChoreForm(ctx, client, collection)
				scope := uniqueChoreName("Reject completion correction")
				created := createManualOneOffChore(
					ctx,
					client,
					form,
					scope,
					"Keep the original completion date.",
					"2020-07-01",
				)
				completed := completeOneOffChore(ctx, client, created, "2020-07-02")
				action := completed.Actions[shiftbellapi.ActionCorrectCompletion]
				futureDate := time.Now().UTC().AddDate(0, 0, 2).Format(time.DateOnly)

				By("submitting a future corrected completion date")
				result, err := client.CorrectChoreCompletion(
					ctx,
					shiftbellapi.RequestParams{
						Method:      action.Method,
						Href:        action.Href,
						ContentType: action.ContentType,
					},
					shiftbellapi.CorrectChoreCompletionParams{CompletedOn: futureDate},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.StatusCode).To(Equal(http.StatusUnprocessableEntity))
				Expect(result.SuccessResponse).To(BeNil())
				Expect(result.ErrorResponse).NotTo(BeNil())
				Expect(result.ErrorResponse.Error).To(Equal("invalid completion date"))

				By("proving the original completion date remains unchanged")
				retrieved, err := client.GetChore(ctx, shiftbellapi.GetChoreParams{
					Link: completed.Chore.Links[shiftbellapi.RelationSelf],
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved.StatusCode).To(Equal(http.StatusOK))
				Expect(retrieved.ErrorResponse).To(BeNil())
				Expect(retrieved.SuccessResponse).NotTo(BeNil())
				Expect(retrieved.SuccessResponse.Chore).To(Equal(completed.Chore))
				Expect(retrieved.SuccessResponse.Actions).To(Equal(completed.Actions))
			},
		)
	})

	When("permanently deleting a one-off chore", func() {
		DescribeTable("deletes the chore",
			func(ctx SpecContext, prepare prepareOneOffChoreForDeletion) {
				collection := discoverChoreCollection(ctx, client)
				form := getManualOneOffChoreForm(ctx, client, collection)
				scope := uniqueChoreName("Delete one-off")
				created := createManualOneOffChore(
					ctx,
					client,
					form,
					scope,
					"Delete this chore permanently.",
					"2020-08-01",
				)
				chore, actions := prepare(ctx, client, created)
				deleteAction := actions[shiftbellapi.ActionDeleteChore]

				By("deleting through the state-appropriate advertised action")
				deleted, err := client.DeleteChore(
					ctx,
					shiftbellapi.RequestParams{
						Method:      deleteAction.Method,
						Href:        deleteAction.Href,
						ContentType: deleteAction.ContentType,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(deleted.StatusCode).To(Equal(http.StatusNoContent))
				Expect(deleted.ErrorResponse).To(BeNil())

				By("proving the former resource is missing")
				missing, err := client.GetChore(ctx, shiftbellapi.GetChoreParams{
					Link: chore.Links[shiftbellapi.RelationSelf],
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(missing.StatusCode).To(Equal(http.StatusNotFound))
				Expect(missing.SuccessResponse).To(BeNil())
				Expect(missing.ErrorResponse).NotTo(BeNil())
				Expect(missing.ErrorResponse.Error).To(Equal("chore not found"))
				Expect(missing.ErrorResponse.Links).To(gstruct.MatchAllKeys(gstruct.Keys{
					shiftbellapi.RelationCollection: gstruct.MatchAllFields(
						gstruct.Fields{
							"Href": Equal(
								collection.Links[shiftbellapi.RelationSelf].Href,
							),
						},
					),
				}))
				Expect(missing.ErrorResponse.Actions).To(BeEmpty())

				By("proving the chore is absent from both status collections")
				active, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: scope,
					Status: "active",
					Limit:  20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(active.Collection.Items).To(BeEmpty())
				history, err := client.BrowseChores(ctx, shiftbellapi.BrowseChoresParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: scope,
					Status: "completed",
					Limit:  20,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(history.Collection.Items).To(BeEmpty())
			},
			Entry("active one-off chore", prepareActiveOneOffChoreForDeletion),
			Entry("completed one-off chore", prepareCompletedOneOffChoreForDeletion),
		)
	})
})

func prepareActiveOneOffChoreForDeletion(
	_ context.Context,
	_ *shiftbellapi.APIClient,
	created *shiftbellapi.CreateChoreResponse,
) (shiftbellapi.Chore, shiftbellapi.Actions) {
	return created.Chore, created.Actions
}

func prepareCompletedOneOffChoreForDeletion(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	created *shiftbellapi.CreateChoreResponse,
) (shiftbellapi.Chore, shiftbellapi.Actions) {
	GinkgoHelper()
	completed := completeOneOffChore(ctx, client, created, "2020-08-02")
	return completed.Chore, completed.Actions
}

func discoverChoreCollection(
	ctx context.Context,
	client *shiftbellapi.APIClient,
) shiftbellapi.ChoreCollection {
	GinkgoHelper()
	home, err := client.GetHome(ctx)
	Expect(err).NotTo(HaveOccurred())
	collection, err := client.GetChores(
		ctx,
		home.Home.Links[shiftbellapi.RelationChores].Href,
	)
	Expect(err).NotTo(HaveOccurred())
	return collection.Collection
}

func uniqueChoreName(prefix string) string {
	GinkgoHelper()
	return fmt.Sprintf("%s %d", prefix, time.Now().UnixNano())
}

func createManualOneOffChore(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	form shiftbellapi.ChoreCreationStep,
	name string,
	description string,
	deadline string,
) *shiftbellapi.CreateChoreResponse {
	GinkgoHelper()
	result, err := client.CreateChore(
		ctx,
		shiftbellapi.RequestParams{
			Method:      form.Action.Method,
			Href:        form.Action.Href,
			ContentType: form.Action.ContentType,
		},
		shiftbellapi.CreateChoreParams{
			Name:        name,
			Description: description,
			Deadline:    deadline,
		},
	)
	Expect(err).NotTo(HaveOccurred())
	Expect(result.StatusCode).To(Equal(http.StatusCreated))
	Expect(result.ErrorResponse).To(BeNil())
	Expect(result.SuccessResponse).NotTo(BeNil())
	return result.SuccessResponse
}

func completeOneOffChore(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	created *shiftbellapi.CreateChoreResponse,
	completedOn string,
) *shiftbellapi.CompleteChoreResponse {
	GinkgoHelper()
	action := created.Actions[shiftbellapi.ActionCompleteChore]
	result, err := client.CompleteChore(
		ctx,
		shiftbellapi.RequestParams{
			Method:      action.Method,
			Href:        action.Href,
			ContentType: action.ContentType,
		},
		shiftbellapi.CompleteChoreParams{CompletedOn: completedOn},
	)
	Expect(err).NotTo(HaveOccurred())
	Expect(result.StatusCode).To(Equal(http.StatusOK))
	Expect(result.ErrorResponse).To(BeNil())
	Expect(result.SuccessResponse).NotTo(BeNil())
	completed := result.SuccessResponse
	Expect(completed.Chore).To(gstruct.MatchAllFields(gstruct.Fields{
		"Id":          Equal(created.Chore.Id),
		"ScheduleId":  BeNil(),
		"Status":      Equal("completed"),
		"Name":        Equal(created.Chore.Name),
		"Description": Equal(created.Chore.Description),
		"Deadline":    Equal(created.Chore.Deadline),
		"CompletedOn": Equal(&completedOn),
		"Links":       Equal(created.Chore.Links),
	}))
	return completed
}

func expectCompletedOneOffActions(actions shiftbellapi.Actions, selfHref string) {
	GinkgoHelper()
	Expect(actions).To(gstruct.MatchAllKeys(gstruct.Keys{
		shiftbellapi.ActionCorrectCompletion: gstruct.MatchAllFields(gstruct.Fields{
			"Href":        Equal(selfHref + "/completion"),
			"Method":      Equal(http.MethodPatch),
			"ContentType": Equal("application/json"),
			"Fields": Equal([]shiftbellapi.ActionField{
				{Name: "completed_on", Type: "date", Required: true},
			}),
		}),
		shiftbellapi.ActionDeleteChore: gstruct.MatchAllFields(gstruct.Fields{
			"Href":        Equal(selfHref),
			"Method":      Equal(http.MethodDelete),
			"ContentType": BeEmpty(),
			"Fields":      BeEmpty(),
		}),
	}))
}

func getManualOneOffChoreForm(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	collection shiftbellapi.ChoreCollection,
) shiftbellapi.ChoreCreationStep {
	GinkgoHelper()
	recurrence := getManualChoreRecurrence(ctx, client, collection)

	By("choosing one-off and retrieving the final form")
	form, err := client.GetChoreCreationStep(ctx, recurrence.Choices[0].Href)
	Expect(err).NotTo(HaveOccurred())
	Expect(form.StatusCode).To(Equal(http.StatusOK))
	Expect(form.ErrorResponse).To(BeNil())
	Expect(form.SuccessResponse).NotTo(BeNil())
	Expect(*form.SuccessResponse).To(gstruct.MatchAllFields(gstruct.Fields{
		"Step":     Equal("form"),
		"Template": BeNil(),
		"Choices":  BeEmpty(),
		"Fields": Equal([]shiftbellapi.ActionField{
			{Name: "name", Type: "string", Required: true},
			{Name: "description", Type: "string", Required: false},
			{Name: "deadline", Type: "date", Required: true},
			{Name: "save_as_chore_template", Type: "boolean", Required: false},
		}),
		"Action": Equal(&shiftbellapi.Action{
			Href:        "/chores",
			Method:      http.MethodPost,
			ContentType: "application/json",
		}),
	}))
	return *form.SuccessResponse
}

func getManualChoreRecurrence(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	collection shiftbellapi.ChoreCollection,
) shiftbellapi.ChoreCreationStep {
	GinkgoHelper()
	By("choosing to specify a new chore")
	source, err := client.GetChoreCreationStep(
		ctx,
		collection.Actions[shiftbellapi.ActionCreateChore].Href,
	)
	Expect(err).NotTo(HaveOccurred())
	Expect(source.StatusCode).To(Equal(http.StatusOK))
	Expect(source.ErrorResponse).To(BeNil())
	Expect(source.SuccessResponse).NotTo(BeNil())
	Expect(*source.SuccessResponse).To(gstruct.MatchAllFields(gstruct.Fields{
		"Step":     Equal("source"),
		"Template": BeNil(),
		"Choices": Equal([]shiftbellapi.ChoreCreationChoice{
			{Label: "Specify new", Href: "/chores/new?source=manual"},
			{Label: "Select template", Href: "/chore-templates?picker=1"},
		}),
		"Fields": BeEmpty(),
		"Action": BeNil(),
	}))

	recurrence, err := client.GetChoreCreationStep(
		ctx,
		source.SuccessResponse.Choices[0].Href,
	)
	Expect(err).NotTo(HaveOccurred())
	Expect(recurrence.StatusCode).To(Equal(http.StatusOK))
	Expect(recurrence.ErrorResponse).To(BeNil())
	Expect(recurrence.SuccessResponse).NotTo(BeNil())
	Expect(*recurrence.SuccessResponse).To(gstruct.MatchAllFields(gstruct.Fields{
		"Step":     Equal("recurrence"),
		"Template": BeNil(),
		"Choices": Equal([]shiftbellapi.ChoreCreationChoice{
			{
				Label: "One-off",
				Href:  "/chores/new?source=manual&recurrence=one-off",
			},
			{
				Label: "Scheduled",
				Href:  "/chores/new?source=manual&recurrence=scheduled",
			},
		}),
		"Fields": BeEmpty(),
		"Action": BeNil(),
	}))
	return *recurrence.SuccessResponse
}

func getTemplateOneOffChoreForm(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	collection shiftbellapi.ChoreCollection,
	template shiftbellapi.ChoreTemplate,
) shiftbellapi.ChoreCreationStep {
	GinkgoHelper()
	recurrence := getTemplateChoreRecurrence(ctx, client, collection, template)

	By("choosing one-off and retrieving the template-based form")
	form, err := client.GetChoreCreationStep(ctx, recurrence.Choices[0].Href)
	Expect(err).NotTo(HaveOccurred())
	Expect(form.StatusCode).To(Equal(http.StatusOK))
	Expect(form.ErrorResponse).To(BeNil())
	Expect(form.SuccessResponse).NotTo(BeNil())
	Expect(*form.SuccessResponse).To(gstruct.MatchAllFields(gstruct.Fields{
		"Step": Equal("form"),
		"Template": Equal(&shiftbellapi.ChoreCreationTemplate{
			Id:   template.Id,
			Name: template.Name,
		}),
		"Choices": BeEmpty(),
		"Fields": Equal([]shiftbellapi.ActionField{
			{Name: "deadline", Type: "date", Required: true},
		}),
		"Action": Equal(&shiftbellapi.Action{
			Href:        "/chores",
			Method:      http.MethodPost,
			ContentType: "application/json",
		}),
	}))
	return *form.SuccessResponse
}

func getTemplateChoreRecurrence(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	collection shiftbellapi.ChoreCollection,
	template shiftbellapi.ChoreTemplate,
) shiftbellapi.ChoreCreationStep {
	GinkgoHelper()
	By("choosing a chore template source")
	source, err := client.GetChoreCreationStep(
		ctx,
		collection.Actions[shiftbellapi.ActionCreateChore].Href,
	)
	Expect(err).NotTo(HaveOccurred())
	Expect(source.StatusCode).To(Equal(http.StatusOK))
	Expect(source.ErrorResponse).To(BeNil())
	Expect(source.SuccessResponse).NotTo(BeNil())
	Expect(*source.SuccessResponse).To(gstruct.MatchAllFields(gstruct.Fields{
		"Step":     Equal("source"),
		"Template": BeNil(),
		"Choices": Equal([]shiftbellapi.ChoreCreationChoice{
			{Label: "Specify new", Href: "/chores/new?source=manual"},
			{Label: "Select template", Href: "/chore-templates?picker=1"},
		}),
		"Fields": BeEmpty(),
		"Action": BeNil(),
	}))

	By("selecting an active chore template")
	picker, err := client.BrowseChoreTemplatePicker(
		ctx,
		shiftbellapi.BrowseChoreTemplatePickerParams{
			Href:   source.SuccessResponse.Choices[1].Href,
			Search: template.Name,
			Limit:  20,
		},
	)
	Expect(err).NotTo(HaveOccurred())
	selectHref := fmt.Sprintf("/chores/new?template_id=%d", template.Id)
	Expect(picker.Collection.Items).To(gstruct.MatchAllElementsWithIndex(
		gstruct.IndexIdentity,
		gstruct.Elements{
			"0": gstruct.MatchAllFields(gstruct.Fields{
				"Id":   Equal(template.Id),
				"Name": Equal(template.Name),
				"Select": gstruct.MatchAllFields(gstruct.Fields{
					"Href": Equal(selectHref),
				}),
			}),
		},
	))
	Expect(picker.Collection.More).To(BeFalse())

	recurrence, err := client.GetChoreCreationStep(ctx, selectHref)
	Expect(err).NotTo(HaveOccurred())
	Expect(recurrence.StatusCode).To(Equal(http.StatusOK))
	Expect(recurrence.ErrorResponse).To(BeNil())
	Expect(recurrence.SuccessResponse).NotTo(BeNil())
	Expect(*recurrence.SuccessResponse).To(gstruct.MatchAllFields(gstruct.Fields{
		"Step": Equal("recurrence"),
		"Template": Equal(&shiftbellapi.ChoreCreationTemplate{
			Id:          template.Id,
			Name:        template.Name,
			Description: template.Description,
		}),
		"Choices": Equal([]shiftbellapi.ChoreCreationChoice{
			{
				Label: "One-off",
				Href: fmt.Sprintf(
					"/chores/new?template_id=%d&recurrence=one-off",
					template.Id,
				),
			},
			{
				Label: "Scheduled",
				Href: fmt.Sprintf(
					"/chores/new?template_id=%d&recurrence=scheduled",
					template.Id,
				),
			},
		}),
		"Fields": BeEmpty(),
		"Action": BeNil(),
	}))
	return *recurrence.SuccessResponse
}
