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
	"github.com/onsi/gomega/types"
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
			Expect(homeResult.Home.Links).To(ConsistOf(
				shiftbellapi.Relation{Rel: shiftbellapi.RelationSelf, Href: "/"},
				shiftbellapi.Relation{Rel: shiftbellapi.RelationChores, Href: "/chores"},
				shiftbellapi.Relation{
					Rel:  shiftbellapi.RelationChoreTemplates,
					Href: "/chore-templates",
				},
			))

			collectionResult, err := client.GetChoreTemplates(
				ctx,
				homeResult.Home.Links.Href(shiftbellapi.RelationChoreTemplates),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(collectionResult.Collection.Items).NotTo(BeNil())
			Expect(collectionResult.Collection.Links).To(ConsistOf(
				shiftbellapi.Relation{
					Rel:  shiftbellapi.RelationSelf,
					Href: "/chore-templates",
				},
			))
			Expect(collectionResult.Collection.Actions).To(ConsistOf(
				shiftbellapi.Relation{
					Rel:  shiftbellapi.RelationCreate,
					Href: "/chore-templates",
				},
			))
			createHref := collectionResult.Collection.Actions.Href(
				shiftbellapi.RelationCreate,
			)

			By("creating a chore template")
			name := uniqueChoreTemplateName("Laundry")
			createResult, err := client.CreateChoreTemplate(
				ctx,
				createHref,
				shiftbellapi.CreateChoreTemplateParams{
					Name:        "  " + name + "  ",
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
				"Name":          Equal(name),
				"Description":   Equal("Wash and fold weekly."),
				"DeactivatedAt": BeNil(),
				"Links": ConsistOf(
					shiftbellapi.Relation{
						Rel:  shiftbellapi.RelationSelf,
						Href: created.Location,
					},
					shiftbellapi.Relation{
						Rel:  shiftbellapi.RelationCollection,
						Href: "/chore-templates",
					},
				),
			}))

			By("retrieving the created chore template")
			getResult, err := client.GetChoreTemplate(
				ctx,
				created.ChoreTemplate.Links.Href(shiftbellapi.RelationSelf),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(getResult.StatusCode).To(Equal(http.StatusOK))
			Expect(getResult.ErrorResponse).To(BeNil())
			Expect(getResult.SuccessResponse).NotTo(BeNil())
			retrieved := getResult.SuccessResponse
			Expect(retrieved.ChoreTemplate).To(matchChoreTemplate(created.ChoreTemplate))
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
					Href:   collection.Links.Href(shiftbellapi.RelationSelf),
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
			Expect(page.Collection.Links).To(ConsistOf(
				gstruct.MatchAllFields(gstruct.Fields{
					"Rel":  Equal(shiftbellapi.RelationSelf),
					"Href": Not(BeEmpty()),
				}),
				gstruct.MatchAllFields(gstruct.Fields{
					"Rel":  Equal(shiftbellapi.RelationNext),
					"Href": Not(BeEmpty()),
				}),
			))

			finalPage, err := client.GetChoreTemplates(
				ctx,
				page.Collection.Links.Href(shiftbellapi.RelationNext),
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
			Expect(finalPage.Collection.Links).To(ConsistOf(
				gstruct.MatchAllFields(gstruct.Fields{
					"Rel":  Equal(shiftbellapi.RelationSelf),
					"Href": Not(BeEmpty()),
				}),
				gstruct.MatchAllFields(gstruct.Fields{
					"Rel":  Equal(shiftbellapi.RelationPrevious),
					"Href": Not(BeEmpty()),
				}),
			))
		},
	)

	It(
		"rejects a case-insensitive duplicate name",
		func(ctx SpecContext) {
			collection := discoverChoreTemplateCollection(ctx, client)
			name := uniqueChoreTemplateName("Duplicate")
			createChoreTemplate(
				ctx,
				client,
				collection,
				name,
				"Existing description",
			)
			createHref := collection.Actions.Href(shiftbellapi.RelationCreate)

			conflict, err := client.CreateChoreTemplate(
				ctx,
				createHref,
				shiftbellapi.CreateChoreTemplateParams{
					Name:        strings.ToUpper(name),
					Description: "Conflicting description",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(conflict.StatusCode).To(Equal(http.StatusConflict))
			Expect(conflict.SuccessResponse).To(BeNil())
			Expect(conflict.ErrorResponse).NotTo(BeNil())
			Expect(conflict.ErrorResponse.Error).NotTo(BeEmpty())
			Expect(conflict.ErrorResponse.Links).To(BeEmpty())
			Expect(conflict.ErrorResponse.Actions).To(ConsistOf(
				shiftbellapi.Relation{Rel: shiftbellapi.RelationCreate, Href: createHref},
			))
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
			selfHref := created.ChoreTemplate.Links.Href(shiftbellapi.RelationSelf)
			expectActiveChoreTemplateActions(created.Actions, selfHref)
			editHref := created.Actions.Href(shiftbellapi.RelationEdit)
			editedName := uniqueChoreTemplateName("Edited template")

			editResult, err := client.EditChoreTemplate(
				ctx,
				editHref,
				shiftbellapi.EditChoreTemplateParams{
					Name:        "  " + editedName + "  ",
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
				"Name":          Equal(editedName),
				"Description":   Equal("Edited description"),
				"DeactivatedAt": BeNil(),
				"Links":         ConsistOf(created.ChoreTemplate.Links),
			}))
			expectActiveChoreTemplateActions(edited.Actions, selfHref)

			getResult, err := client.GetChoreTemplate(
				ctx,
				edited.ChoreTemplate.Links.Href(shiftbellapi.RelationSelf),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(getResult.StatusCode).To(Equal(http.StatusOK))
			Expect(getResult.ErrorResponse).To(BeNil())
			Expect(getResult.SuccessResponse).NotTo(BeNil())
			retrieved := getResult.SuccessResponse
			Expect(retrieved.ChoreTemplate).To(matchChoreTemplate(edited.ChoreTemplate))
			Expect(retrieved.Actions).To(ConsistOf(edited.Actions))
		},
	)

	It("rejects editing a chore template to an existing name", func(ctx SpecContext) {
		collection := discoverChoreTemplateCollection(ctx, client)
		existingName := uniqueChoreTemplateName("Existing edit name")
		createChoreTemplate(
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
		selfHref := target.ChoreTemplate.Links.Href(shiftbellapi.RelationSelf)
		expectActiveChoreTemplateActions(target.Actions, selfHref)
		editHref := target.Actions.Href(shiftbellapi.RelationEdit)

		conflict, err := client.EditChoreTemplate(
			ctx,
			editHref,
			shiftbellapi.EditChoreTemplateParams{
				Name:        strings.ToUpper(existingName),
				Description: "Conflicting edit",
			},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(conflict.StatusCode).To(Equal(http.StatusConflict))
		Expect(conflict.SuccessResponse).To(BeNil())
		Expect(conflict.ErrorResponse).NotTo(BeNil())
		Expect(conflict.ErrorResponse.Error).NotTo(BeEmpty())
		Expect(conflict.ErrorResponse.Links).To(BeEmpty())
		Expect(conflict.ErrorResponse.Actions).To(ConsistOf(
			shiftbellapi.Relation{Rel: shiftbellapi.RelationEdit, Href: editHref},
		))
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
			selfHref := created.ChoreTemplate.Links.Href(shiftbellapi.RelationSelf)
			expectActiveChoreTemplateActions(created.Actions, selfHref)
			deactivateHref := created.Actions.Href(shiftbellapi.RelationDeactivate)

			deactivated, err := client.DeactivateChoreTemplate(ctx, deactivateHref)
			Expect(err).NotTo(HaveOccurred())
			Expect(deactivated.ChoreTemplate).To(gstruct.MatchAllFields(gstruct.Fields{
				"Id":            Equal(created.ChoreTemplate.Id),
				"Name":          Equal(name),
				"Description":   Equal("Deactivate description"),
				"DeactivatedAt": Not(BeNil()),
				"Links":         ConsistOf(created.ChoreTemplate.Links),
			}))
			Expect(deactivated.Actions).NotTo(BeNil())
			Expect(deactivated.Actions).To(BeEmpty())

			active, err := client.BrowseChoreTemplates(
				ctx,
				shiftbellapi.BrowseChoreTemplatesParams{
					Href:   collection.Links.Href(shiftbellapi.RelationSelf),
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
					Href:   collection.Links.Href(shiftbellapi.RelationSelf),
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
		createHref := collection.Actions.Href(shiftbellapi.RelationCreate)

		invalid, err := client.CreateChoreTemplate(
			ctx,
			createHref,
			shiftbellapi.CreateChoreTemplateParams{
				Name:        "   ",
				Description: "Invalid template",
			},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(invalid.StatusCode).To(Equal(http.StatusUnprocessableEntity))
		Expect(invalid.SuccessResponse).To(BeNil())
		Expect(invalid.ErrorResponse).NotTo(BeNil())
		Expect(invalid.ErrorResponse.Error).To(Equal("invalid name"))
		Expect(invalid.ErrorResponse.Links).To(ConsistOf(
			shiftbellapi.Relation{
				Rel:  shiftbellapi.RelationCollection,
				Href: collection.Links.Href(shiftbellapi.RelationSelf),
			},
		))
		Expect(invalid.ErrorResponse.Actions).To(ConsistOf(
			shiftbellapi.Relation{Rel: shiftbellapi.RelationCreate, Href: createHref},
		))
	})

	It(
		"returns collection navigation for a missing chore template",
		func(ctx SpecContext) {
			collection := discoverChoreTemplateCollection(ctx, client)
			missingHref := strings.TrimRight(
				collection.Links.Href(shiftbellapi.RelationSelf),
				"/",
			) + "/999999999"

			missing, err := client.GetChoreTemplate(ctx, missingHref)
			Expect(err).NotTo(HaveOccurred())
			Expect(missing.StatusCode).To(Equal(http.StatusNotFound))
			Expect(missing.SuccessResponse).To(BeNil())
			Expect(missing.ErrorResponse).NotTo(BeNil())
			Expect(missing.ErrorResponse.Error).To(Equal("chore template not found"))
			Expect(missing.ErrorResponse.Links).To(ConsistOf(
				shiftbellapi.Relation{
					Rel:  shiftbellapi.RelationCollection,
					Href: collection.Links.Href(shiftbellapi.RelationSelf),
				},
			))
			Expect(missing.ErrorResponse.Actions).To(BeEmpty())
		},
	)
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
		home.Home.Links.Href(shiftbellapi.RelationChoreTemplates),
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
	result, err := client.CreateChoreTemplate(
		ctx,
		collection.Actions.Href(shiftbellapi.RelationCreate),
		shiftbellapi.CreateChoreTemplateParams{
			Name:        name,
			Description: description,
		},
	)
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

func expectActiveChoreTemplateActions(actions shiftbellapi.Relations, selfHref string) {
	GinkgoHelper()
	Expect(actions).To(ConsistOf(
		shiftbellapi.Relation{Rel: shiftbellapi.RelationEdit, Href: selfHref},
		shiftbellapi.Relation{
			Rel:  shiftbellapi.RelationDeactivate,
			Href: selfHref + "/deactivation",
		},
	))
}

func matchChoreTemplate(expected shiftbellapi.ChoreTemplate) types.GomegaMatcher {
	return gstruct.MatchAllFields(gstruct.Fields{
		"Id":            Equal(expected.Id),
		"Name":          Equal(expected.Name),
		"Description":   Equal(expected.Description),
		"DeactivatedAt": Equal(expected.DeactivatedAt),
		"Links":         ConsistOf(expected.Links),
	})
}
