package acceptance_test

import (
	"os"

	"github.com/bpsoos/shiftbell/internal/testsupport/shiftbellapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
)

var _ = Describe("Profile API", func() {
	It(
		"creates a passwordless profile and lists its stable identity",
		func(ctx SpecContext) {
			baseURL := os.Getenv("SHIFTBELL_BASE_URL")
			Expect(baseURL).NotTo(BeEmpty())
			client := shiftbellapi.NewAPIClient(baseURL)
			home, err := client.GetHome(ctx)
			Expect(err).NotTo(HaveOccurred())
			href := home.Home.Links.Href(shiftbellapi.RelationProfiles)
			Expect(href).NotTo(BeEmpty())
			collection, err := client.GetProfiles(ctx, href)
			Expect(err).NotTo(HaveOccurred())
			created, err := client.CreateProfile(
				ctx,
				collection.Actions.Href(shiftbellapi.RelationCreate),
				"Alex",
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(created.Name).To(Equal("Alex"))
			Expect(created.Id).To(BeNumerically(">", 0))
			listed, err := client.GetProfiles(
				ctx,
				created.Links.Href(shiftbellapi.RelationCollection),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(listed.Items).To(ContainElements(
				gstruct.MatchAllFields(gstruct.Fields{
					"Id":   gstruct.Ignore(),
					"Name": Equal("Developer"),
					"Links": ConsistOf(shiftbellapi.Relation{
						Rel: shiftbellapi.RelationCollection, Href: "/profiles",
					}),
				}),
				gstruct.MatchAllFields(gstruct.Fields{
					"Id":   Equal(created.Id),
					"Name": Equal("Alex"),
					"Links": ConsistOf(shiftbellapi.Relation{
						Rel: shiftbellapi.RelationCollection, Href: "/profiles",
					}),
				}),
			))
		},
	)

	It("lists the initialized profile and reports setup complete", func(ctx SpecContext) {
		baseURL := os.Getenv("SHIFTBELL_BASE_URL")
		Expect(baseURL).NotTo(BeEmpty())
		client := shiftbellapi.NewAPIClient(baseURL)
		home, err := client.GetHome(ctx)
		Expect(err).NotTo(HaveOccurred())
		collection, err := client.GetProfiles(
			ctx,
			home.Home.Links.Href(shiftbellapi.RelationProfiles),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(
			collection.Items,
		).To(ContainElements(gstruct.MatchAllFields(gstruct.Fields{
			"Name": Equal("Developer"),
			"Id":   gstruct.Ignore(),
			"Links": ConsistOf(shiftbellapi.Relation{
				Rel: shiftbellapi.RelationCollection, Href: "/profiles",
			}),
		})))
		setup, err := client.GetProfileSetup(
			ctx,
			home.Home.Links.Href(shiftbellapi.RelationSetup),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(setup.Required).To(BeFalse())
		Expect(setup.Actions).To(BeEmpty())
	})
})
