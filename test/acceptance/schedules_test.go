package acceptance_test

import (
	"os"

	"github.com/bpsoos/shiftbell/internal/testsupport/shiftbellapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Schedule API", func() {
	It("discovers an empty schedules collection from home", func(ctx SpecContext) {
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
		Expect(collection.Items).To(BeEmpty())
		Expect(collection.More).To(BeFalse())
		Expect(collection.Links).To(ConsistOf(shiftbellapi.Relation{
			Rel: shiftbellapi.RelationSelf, Href: href,
		}))
	})
})
