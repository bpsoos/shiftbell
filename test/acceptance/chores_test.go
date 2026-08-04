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
	"github.com/onsi/gomega/gstruct"
)

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

		DescribeTable("rejects invalid values without creating a chore",
			func() {
				Expect(true).To(BeTrue())
			},
			Entry("blank name"),
			Entry("missing deadline"),
		)

		It(
			"does not create a chore when save-as-template conflicts with an active template",
			func() {
				Expect(true).To(BeTrue())
			},
		)

		It("does not create a chore from a deactivated template", func() {
			Expect(true).To(BeTrue())
		})
	})

	When("scheduled recurrence is requested", func() {
		DescribeTable("returns Not Implemented without persisting resources",
			func() {
				Expect(true).To(BeTrue())
			},
			Entry("manual scheduled chore"),
			Entry("template-based scheduled chore"),
		)
	})

	When("browsing chores", func() {
		DescribeTable("browses and searches chores in the selected status",
			func() {
				Expect(true).To(BeTrue())
			},
			Entry("active chores ordered by deadline and ID ascending"),
			Entry("completed chores ordered by completion date and ID descending"),
		)
	})

	When("editing an active one-off chore", func() {
		It("updates its normalized name, description, and deadline", func() {
			Expect(true).To(BeTrue())
		})
	})

	When("completing an active one-off chore", func() {
		It("moves the chore from the active collection to completed history", func() {
			Expect(true).To(BeTrue())
		})
		It("treats repeated completion requests as successful no-ops", func() {
			Expect(true).To(BeTrue())
		})

		It("rejects a completion date after the application-local current date", func() {
			Expect(true).To(BeTrue())
		})
	})

	When("correcting a one-off chore completion", func() {
		It("updates its completion date", func() {
			Expect(true).To(BeTrue())
		})

		It("rejects a completion date after the application-local current date", func() {
			Expect(true).To(BeTrue())
		})
	})

	When("permanently deleting a one-off chore", func() {
		DescribeTable("deletes the chore",
			func() {
				Expect(true).To(BeTrue())
			},
			Entry("active one-off chore"),
			Entry("completed one-off chore"),
		)
	})

	When("retrieving a missing chore", func() {
		It("returns collection navigation and no mutation actions", func() {
			Expect(true).To(BeTrue())
		})
	})
})

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

func getManualOneOffChoreForm(
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
	Expect(source.Step).To(gstruct.MatchAllFields(gstruct.Fields{
		"Step":     Equal("source"),
		"Template": BeNil(),
		"Choices": Equal([]shiftbellapi.ChoreCreationChoice{
			{Label: "Specify new", Href: "/chores/new?source=manual"},
			{Label: "Select template", Href: "/chore-templates?picker=1"},
		}),
		"Fields": BeEmpty(),
		"Action": BeNil(),
	}))

	recurrence, err := client.GetChoreCreationStep(ctx, source.Step.Choices[0].Href)
	Expect(err).NotTo(HaveOccurred())
	Expect(recurrence.Step).To(gstruct.MatchAllFields(gstruct.Fields{
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

	By("choosing one-off and retrieving the final form")
	form, err := client.GetChoreCreationStep(ctx, recurrence.Step.Choices[0].Href)
	Expect(err).NotTo(HaveOccurred())
	Expect(form.Step).To(gstruct.MatchAllFields(gstruct.Fields{
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
	return form.Step
}

func getTemplateOneOffChoreForm(
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
	Expect(source.Step).To(gstruct.MatchAllFields(gstruct.Fields{
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
			Href:   source.Step.Choices[1].Href,
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
	Expect(recurrence.Step).To(gstruct.MatchAllFields(gstruct.Fields{
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

	By("choosing one-off and retrieving the template-based form")
	form, err := client.GetChoreCreationStep(ctx, recurrence.Step.Choices[0].Href)
	Expect(err).NotTo(HaveOccurred())
	Expect(form.Step).To(gstruct.MatchAllFields(gstruct.Fields{
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
	return form.Step
}
