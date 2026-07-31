package acceptance_test

import (
	"net/http"
	"os"

	"github.com/bpsoos/shiftbell/internal/testsupport/shiftbellapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Chore template API", func() {
	var (
		baseUrl string
		client  *shiftbellapi.APIClient
	)

	BeforeEach(func() {
		baseUrl = os.Getenv("SHIFTBELL_BASE_URL")
		Expect(baseUrl).NotTo(BeEmpty())
		client = shiftbellapi.NewAPIClient(baseUrl)
	})

	It("creates a normalized chore template that can be retrieved", func(ctx SpecContext) {
		By("discovering the chore template collection from home")
		homeResult, err := client.GetHome(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(homeResult.Home.Links).To(MatchAllKeys(Keys{
			shiftbellapi.RelationSelf: MatchAllFields(Fields{
				"Href": Equal("/"),
			}),
			shiftbellapi.RelationChoreTemplates: MatchAllFields(Fields{
				"Href": Equal("/chore-templates"),
			}),
		}))

		collectionResult, err := client.GetChoreTemplates(
			ctx,
			homeResult.Home.Links[shiftbellapi.RelationChoreTemplates].Href,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(collectionResult.Collection.Items).NotTo(BeNil())
		Expect(collectionResult.Collection.Links).To(MatchAllKeys(Keys{
			shiftbellapi.RelationSelf: MatchAllFields(Fields{
				"Href": Equal("/chore-templates"),
			}),
		}))
		Expect(collectionResult.Collection.Actions).To(MatchAllKeys(Keys{
			shiftbellapi.ActionCreateChoreTemplate: MatchAllFields(Fields{
				"Href":        Equal("/chore-templates"),
				"Method":      Equal(http.MethodPost),
				"ContentType": Equal("application/json"),
				"Fields": MatchAllElementsWithIndex(IndexIdentity, Elements{
					"0": MatchAllFields(Fields{
						"Name":      Equal("name"),
						"Type":      Equal("string"),
						"Required":  BeTrue(),
						"MaxLength": Equal(200),
					}),
					"1": MatchAllFields(Fields{
						"Name":      Equal("description"),
						"Type":      Equal("string"),
						"Required":  BeFalse(),
						"MaxLength": Equal(2000),
					}),
				}),
			}),
		}))
		createAction := collectionResult.Collection.Actions[shiftbellapi.ActionCreateChoreTemplate]

		By("creating a chore template")
		created, err := client.CreateChoreTemplate(ctx, shiftbellapi.CreateChoreTemplateParams{
			Action:      createAction,
			Name:        "  Laundry  ",
			Description: "  Wash and fold weekly.  ",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(created.Location).NotTo(BeEmpty())
		Expect(created.ChoreTemplate).To(MatchAllFields(Fields{
			"Id":            BeNumerically(">", 0),
			"Name":          Equal("Laundry"),
			"Description":   Equal("Wash and fold weekly."),
			"DeactivatedAt": BeNil(),
			"Links": MatchAllKeys(Keys{
				shiftbellapi.RelationSelf: MatchAllFields(Fields{
					"Href": Equal(created.Location),
				}),
				shiftbellapi.RelationCollection: MatchAllFields(Fields{
					"Href": Equal("/chore-templates"),
				}),
			}),
		}))

		By("retrieving the created chore template")
		retrieved, err := client.GetChoreTemplate(ctx, shiftbellapi.GetChoreTemplateParams{
			Link: created.ChoreTemplate.Links[shiftbellapi.RelationSelf],
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(retrieved.ChoreTemplate).To(Equal(created.ChoreTemplate))
	})
})
