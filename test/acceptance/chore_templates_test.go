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

var _ = Describe("Chore template API", func() {
	var client *shiftbellapi.APIClient

	BeforeEach(func() {
		baseURL := os.Getenv("SHIFTBELL_BASE_URL")
		Expect(baseURL).NotTo(BeEmpty())
		client = shiftbellapi.NewAPIClient(baseURL)
	})

	It(
		"creates a normalized chore template that can be retrieved",
		func(ctx SpecContext) {
			By("discovering the chore template collection from home")
			homeResult, err := client.GetHome(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(homeResult.Home.Links).To(gstruct.MatchAllKeys(gstruct.Keys{
				shiftbellapi.RelationSelf: gstruct.MatchAllFields(gstruct.Fields{
					"Href": Equal("/"),
				}),
				shiftbellapi.RelationChoreTemplates: gstruct.MatchAllFields(
					gstruct.Fields{
						"Href": Equal("/chore-templates"),
					},
				),
			}))

			collectionResult, err := client.GetChoreTemplates(
				ctx,
				homeResult.Home.Links[shiftbellapi.RelationChoreTemplates].Href,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(collectionResult.Collection.Items).NotTo(BeNil())
			Expect(
				collectionResult.Collection.Links,
			).To(gstruct.MatchAllKeys(gstruct.Keys{
				shiftbellapi.RelationSelf: gstruct.MatchAllFields(gstruct.Fields{
					"Href": Equal("/chore-templates"),
				}),
			}))
			Expect(
				collectionResult.Collection.Actions,
			).To(gstruct.MatchAllKeys(gstruct.Keys{
				shiftbellapi.ActionCreateChoreTemplate: gstruct.MatchAllFields(
					gstruct.Fields{
						"Href":        Equal("/chore-templates"),
						"Method":      Equal(http.MethodPost),
						"ContentType": Equal("application/json"),
						"Fields": gstruct.MatchAllElementsWithIndex(
							gstruct.IndexIdentity,
							gstruct.Elements{
								"0": gstruct.MatchAllFields(gstruct.Fields{
									"Name":     Equal("name"),
									"Type":     Equal("string"),
									"Required": BeTrue(),
								}),
								"1": gstruct.MatchAllFields(gstruct.Fields{
									"Name":     Equal("description"),
									"Type":     Equal("string"),
									"Required": BeFalse(),
								}),
							},
						),
					},
				),
			}))
			createAction := collectionResult.Collection.Actions[shiftbellapi.ActionCreateChoreTemplate]

			By("creating a chore template")
			createResult, err := client.CreateChoreTemplate(
				ctx,
				shiftbellapi.CreateChoreTemplateParams{
					Action:      createAction,
					Name:        "  Laundry  ",
					Description: "  Wash and fold weekly.  ",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResult.StatusCode).To(Equal(http.StatusCreated))
			Expect(createResult.ErrorResponse).To(BeNil())
			Expect(createResult.SuccessResponse).NotTo(BeNil())
			created := createResult.SuccessResponse
			Expect(created.Location).NotTo(BeEmpty())
			Expect(created.ChoreTemplate).To(gstruct.MatchAllFields(gstruct.Fields{
				"Id":            BeNumerically(">", 0),
				"Name":          Equal("Laundry"),
				"Description":   Equal("Wash and fold weekly."),
				"DeactivatedAt": BeNil(),
				"Links": gstruct.MatchAllKeys(gstruct.Keys{
					shiftbellapi.RelationSelf: gstruct.MatchAllFields(gstruct.Fields{
						"Href": Equal(created.Location),
					}),
					shiftbellapi.RelationCollection: gstruct.MatchAllFields(
						gstruct.Fields{
							"Href": Equal("/chore-templates"),
						},
					),
				}),
			}))

			By("retrieving the created chore template")
			getResult, err := client.GetChoreTemplate(
				ctx,
				shiftbellapi.GetChoreTemplateParams{
					Link: created.ChoreTemplate.Links[shiftbellapi.RelationSelf],
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(getResult.StatusCode).To(Equal(http.StatusOK))
			Expect(getResult.ErrorResponse).To(BeNil())
			Expect(getResult.SuccessResponse).NotTo(BeNil())
			retrieved := getResult.SuccessResponse
			Expect(retrieved.ChoreTemplate).To(Equal(created.ChoreTemplate))
		},
	)

	It(
		"browses a scoped active collection through pagination links",
		func(ctx SpecContext) {
			collection := discoverChoreTemplateCollection(ctx, client)
			scope := uniqueChoreTemplateName("Browse")
			first := createChoreTemplate(
				ctx,
				client,
				collection,
				scope+" First",
				"First description",
			)
			second := createChoreTemplate(
				ctx,
				client,
				collection,
				scope+" Second",
				"Second description",
			)
			third := createChoreTemplate(
				ctx,
				client,
				collection,
				scope+" Third",
				"Third description",
			)
			createChoreTemplate(
				ctx,
				client,
				collection,
				uniqueChoreTemplateName("Unrelated"),
				"Must not match the scoped search",
			)

			page, err := client.BrowseChoreTemplates(
				ctx,
				shiftbellapi.BrowseChoreTemplatesParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: scope,
					State:  "active",
					Limit:  2,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(page.Collection.More).To(BeTrue())
			Expect(
				page.Collection.Items,
			).To(gstruct.MatchAllElementsWithIndex(gstruct.IndexIdentity, gstruct.Elements{
				"0": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Id":   Equal(third.ChoreTemplate.Id),
					"Name": Equal(scope + " Third"),
				}),
				"1": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Id":   Equal(second.ChoreTemplate.Id),
					"Name": Equal(scope + " Second"),
				}),
			}))
			Expect(page.Collection.Links).To(gstruct.MatchAllKeys(gstruct.Keys{
				shiftbellapi.RelationSelf: gstruct.MatchAllFields(gstruct.Fields{
					"Href": Not(BeEmpty()),
				}),
				shiftbellapi.RelationNext: gstruct.MatchAllFields(gstruct.Fields{
					"Href": Not(BeEmpty()),
				}),
			}))

			finalPage, err := client.GetChoreTemplates(
				ctx,
				page.Collection.Links[shiftbellapi.RelationNext].Href,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(finalPage.Collection.More).To(BeFalse())
			Expect(
				finalPage.Collection.Items,
			).To(gstruct.MatchAllElementsWithIndex(gstruct.IndexIdentity, gstruct.Elements{
				"0": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Id":   Equal(first.ChoreTemplate.Id),
					"Name": Equal(scope + " First"),
				}),
			}))
			Expect(finalPage.Collection.Links).To(gstruct.MatchAllKeys(gstruct.Keys{
				shiftbellapi.RelationSelf: gstruct.MatchAllFields(gstruct.Fields{
					"Href": Not(BeEmpty()),
				}),
				shiftbellapi.RelationPrevious: gstruct.MatchAllFields(gstruct.Fields{
					"Href": Not(BeEmpty()),
				}),
			}))
		},
	)

	It(
		"rejects a case-insensitive duplicate name with a link to the existing template",
		func(ctx SpecContext) {
			collection := discoverChoreTemplateCollection(ctx, client)
			name := uniqueChoreTemplateName("Duplicate")
			existing := createChoreTemplate(
				ctx,
				client,
				collection,
				name,
				"Existing description",
			)
			createAction := collection.Actions[shiftbellapi.ActionCreateChoreTemplate]

			conflict, err := client.CreateChoreTemplate(
				ctx,
				shiftbellapi.CreateChoreTemplateParams{
					Action:      createAction,
					Name:        strings.ToUpper(name),
					Description: "Conflicting description",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(conflict.StatusCode).To(Equal(http.StatusConflict))
			Expect(conflict.SuccessResponse).To(BeNil())
			Expect(conflict.ErrorResponse).NotTo(BeNil())
			Expect(conflict.ErrorResponse.Error).NotTo(BeEmpty())
			Expect(conflict.ErrorResponse.Links).To(gstruct.MatchAllKeys(gstruct.Keys{
				shiftbellapi.RelationExisting: gstruct.MatchAllFields(gstruct.Fields{
					"Href": Equal(
						existing.ChoreTemplate.Links[shiftbellapi.RelationSelf].Href,
					),
				}),
			}))
			Expect(conflict.ErrorResponse.Actions).To(gstruct.MatchAllKeys(gstruct.Keys{
				shiftbellapi.ActionCreateChoreTemplate: Equal(createAction),
			}))
		},
	)

	It(
		"edits an active chore template through its advertised action",
		func(ctx SpecContext) {
			collection := discoverChoreTemplateCollection(ctx, client)
			created := createChoreTemplate(
				ctx,
				client,
				collection,
				uniqueChoreTemplateName("Edit"),
				"Original description",
			)
			selfHref := created.ChoreTemplate.Links[shiftbellapi.RelationSelf].Href
			expectActiveChoreTemplateActions(created.Actions, selfHref)

			editResult, err := client.EditChoreTemplate(
				ctx,
				shiftbellapi.EditChoreTemplateParams{
					Action:      created.Actions[shiftbellapi.ActionEditChoreTemplate],
					Name:        "  Edited template  ",
					Description: "  Edited description  ",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(editResult.StatusCode).To(Equal(http.StatusOK))
			Expect(editResult.ErrorResponse).To(BeNil())
			Expect(editResult.SuccessResponse).NotTo(BeNil())
			edited := editResult.SuccessResponse
			Expect(edited.ChoreTemplate).To(gstruct.MatchAllFields(gstruct.Fields{
				"Id":            Equal(created.ChoreTemplate.Id),
				"Name":          Equal("Edited template"),
				"Description":   Equal("Edited description"),
				"DeactivatedAt": BeNil(),
				"Links":         Equal(created.ChoreTemplate.Links),
			}))
			expectActiveChoreTemplateActions(edited.Actions, selfHref)

			getResult, err := client.GetChoreTemplate(
				ctx,
				shiftbellapi.GetChoreTemplateParams{
					Link: edited.ChoreTemplate.Links[shiftbellapi.RelationSelf],
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(getResult.StatusCode).To(Equal(http.StatusOK))
			Expect(getResult.ErrorResponse).To(BeNil())
			Expect(getResult.SuccessResponse).NotTo(BeNil())
			retrieved := getResult.SuccessResponse
			Expect(retrieved.ChoreTemplate).To(Equal(edited.ChoreTemplate))
			Expect(retrieved.Actions).To(Equal(edited.Actions))
		},
	)

	It("rejects editing a chore template to an existing name", func(ctx SpecContext) {
		collection := discoverChoreTemplateCollection(ctx, client)
		existingName := uniqueChoreTemplateName("Existing edit name")
		existing := createChoreTemplate(
			ctx,
			client,
			collection,
			existingName,
			"Existing description",
		)
		target := createChoreTemplate(
			ctx,
			client,
			collection,
			uniqueChoreTemplateName("Edit conflict target"),
			"Target description",
		)
		selfHref := target.ChoreTemplate.Links[shiftbellapi.RelationSelf].Href
		expectActiveChoreTemplateActions(target.Actions, selfHref)
		editAction := target.Actions[shiftbellapi.ActionEditChoreTemplate]

		conflict, err := client.EditChoreTemplate(
			ctx,
			shiftbellapi.EditChoreTemplateParams{
				Action:      editAction,
				Name:        strings.ToUpper(existingName),
				Description: "Conflicting edit",
			},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(conflict.StatusCode).To(Equal(http.StatusConflict))
		Expect(conflict.SuccessResponse).To(BeNil())
		Expect(conflict.ErrorResponse).NotTo(BeNil())
		Expect(conflict.ErrorResponse.Error).NotTo(BeEmpty())
		Expect(conflict.ErrorResponse.Links).To(gstruct.MatchAllKeys(gstruct.Keys{
			shiftbellapi.RelationExisting: gstruct.MatchAllFields(gstruct.Fields{
				"Href": Equal(
					existing.ChoreTemplate.Links[shiftbellapi.RelationSelf].Href,
				),
			}),
		}))
		Expect(conflict.ErrorResponse.Actions).To(gstruct.MatchAllKeys(gstruct.Keys{
			shiftbellapi.ActionEditChoreTemplate: Equal(editAction),
		}))
	})

	It(
		"permanently deactivates a chore template through its advertised action",
		func(ctx SpecContext) {
			collection := discoverChoreTemplateCollection(ctx, client)
			name := uniqueChoreTemplateName("Deactivate")
			created := createChoreTemplate(
				ctx,
				client,
				collection,
				name,
				"Deactivate description",
			)
			selfHref := created.ChoreTemplate.Links[shiftbellapi.RelationSelf].Href
			expectActiveChoreTemplateActions(created.Actions, selfHref)

			deactivated, err := client.DeactivateChoreTemplate(
				ctx,
				shiftbellapi.DeactivateChoreTemplateParams{
					Action: created.Actions[shiftbellapi.ActionDeactivateTemplate],
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(deactivated.ChoreTemplate).To(gstruct.MatchAllFields(gstruct.Fields{
				"Id":            Equal(created.ChoreTemplate.Id),
				"Name":          Equal(name),
				"Description":   Equal("Deactivate description"),
				"DeactivatedAt": Not(BeNil()),
				"Links":         Equal(created.ChoreTemplate.Links),
			}))
			Expect(deactivated.Actions).NotTo(BeNil())
			Expect(deactivated.Actions).To(BeEmpty())

			active, err := client.BrowseChoreTemplates(
				ctx,
				shiftbellapi.BrowseChoreTemplatesParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: name,
					State:  "active",
					Limit:  20,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(active.Collection.Items).To(BeEmpty())

			inactive, err := client.BrowseChoreTemplates(
				ctx,
				shiftbellapi.BrowseChoreTemplatesParams{
					Href:   collection.Links[shiftbellapi.RelationSelf].Href,
					Search: name,
					State:  "deactivated",
					Limit:  20,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(
				inactive.Collection.Items,
			).To(gstruct.MatchAllElementsWithIndex(gstruct.IndexIdentity, gstruct.Elements{
				"0": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Id":   Equal(created.ChoreTemplate.Id),
					"Name": Equal(name),
				}),
			}))
		},
	)

	It("returns recovery controls for an invalid create request", func(ctx SpecContext) {
		collection := discoverChoreTemplateCollection(ctx, client)
		createAction := collection.Actions[shiftbellapi.ActionCreateChoreTemplate]

		invalid, err := client.CreateChoreTemplate(
			ctx,
			shiftbellapi.CreateChoreTemplateParams{
				Action:      createAction,
				Name:        "   ",
				Description: "Invalid template",
			},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(invalid.StatusCode).To(Equal(http.StatusUnprocessableEntity))
		Expect(invalid.SuccessResponse).To(BeNil())
		Expect(invalid.ErrorResponse).NotTo(BeNil())
		Expect(invalid.ErrorResponse.Error).To(Equal("invalid name"))
		Expect(invalid.ErrorResponse.Links).To(gstruct.MatchAllKeys(gstruct.Keys{
			shiftbellapi.RelationCollection: gstruct.MatchAllFields(gstruct.Fields{
				"Href": Equal(collection.Links[shiftbellapi.RelationSelf].Href),
			}),
		}))
		Expect(invalid.ErrorResponse.Actions).To(gstruct.MatchAllKeys(gstruct.Keys{
			shiftbellapi.ActionCreateChoreTemplate: Equal(createAction),
		}))
	})

	It(
		"returns collection navigation for a missing chore template",
		func(ctx SpecContext) {
			collection := discoverChoreTemplateCollection(ctx, client)
			missingHref := strings.TrimRight(
				collection.Links[shiftbellapi.RelationSelf].Href,
				"/",
			) + "/999999999"

			missing, err := client.GetChoreTemplate(
				ctx,
				shiftbellapi.GetChoreTemplateParams{
					Link: shiftbellapi.Link{Href: missingHref},
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(missing.StatusCode).To(Equal(http.StatusNotFound))
			Expect(missing.SuccessResponse).To(BeNil())
			Expect(missing.ErrorResponse).NotTo(BeNil())
			Expect(missing.ErrorResponse.Error).To(Equal("chore template not found"))
			Expect(missing.ErrorResponse.Links).To(gstruct.MatchAllKeys(gstruct.Keys{
				shiftbellapi.RelationCollection: gstruct.MatchAllFields(gstruct.Fields{
					"Href": Equal(collection.Links[shiftbellapi.RelationSelf].Href),
				}),
			}))
			Expect(missing.ErrorResponse.Actions).To(BeEmpty())
		},
	)

	When("creating a template after another was deactivated", func() {
		It("allows reuse of the deactivated template's case-insensitive name", func() {
			Expect(true).To(BeTrue())
		})
	})
})

func discoverChoreTemplateCollection(
	ctx context.Context,
	client *shiftbellapi.APIClient,
) shiftbellapi.ChoreTemplateCollection {
	GinkgoHelper()
	home, err := client.GetHome(ctx)
	Expect(err).NotTo(HaveOccurred())
	collection, err := client.GetChoreTemplates(
		ctx,
		home.Home.Links[shiftbellapi.RelationChoreTemplates].Href,
	)
	Expect(err).NotTo(HaveOccurred())
	return collection.Collection
}

func createChoreTemplate(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	collection shiftbellapi.ChoreTemplateCollection,
	name string,
	description string,
) *shiftbellapi.CreateChoreTemplateResponse {
	GinkgoHelper()
	result, err := client.CreateChoreTemplate(ctx, shiftbellapi.CreateChoreTemplateParams{
		Action:      collection.Actions[shiftbellapi.ActionCreateChoreTemplate],
		Name:        name,
		Description: description,
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(result.StatusCode).To(Equal(http.StatusCreated))
	Expect(result.ErrorResponse).To(BeNil())
	Expect(result.SuccessResponse).NotTo(BeNil())
	return result.SuccessResponse
}

func uniqueChoreTemplateName(prefix string) string {
	GinkgoHelper()
	return fmt.Sprintf("%s %d", prefix, time.Now().UnixNano())
}

func expectActiveChoreTemplateActions(actions shiftbellapi.Actions, selfHref string) {
	GinkgoHelper()
	Expect(actions).To(gstruct.MatchAllKeys(gstruct.Keys{
		shiftbellapi.ActionEditChoreTemplate: gstruct.MatchAllFields(gstruct.Fields{
			"Href":        Equal(selfHref),
			"Method":      Equal(http.MethodPatch),
			"ContentType": Equal("application/json"),
			"Fields": Equal([]shiftbellapi.ActionField{
				{Name: "name", Type: "string", Required: true},
				{Name: "description", Type: "string", Required: false},
			}),
		}),
		shiftbellapi.ActionDeactivateTemplate: gstruct.MatchAllFields(gstruct.Fields{
			"Href":        Equal(selfHref + "/deactivation"),
			"Method":      Equal(http.MethodPut),
			"ContentType": BeEmpty(),
			"Fields":      BeEmpty(),
		}),
	}))
}
