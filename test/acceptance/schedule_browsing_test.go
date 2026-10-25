package acceptance_test

import (
	"context"

	"github.com/bpsoos/shiftbell/internal/testsupport/shiftbellapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func browseAllSchedules(
	ctx context.Context,
	client *shiftbellapi.APIClient,
) []shiftbellapi.Schedule {
	GinkgoHelper()
	home, err := client.GetHome(ctx)
	Expect(err).NotTo(HaveOccurred())
	href := home.Home.Links.Href(shiftbellapi.RelationSchedules)
	Expect(href).NotTo(BeEmpty())
	return browseAllSchedulesFrom(ctx, client, href)
}

func browseAllSchedulesFrom(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	href string,
) []shiftbellapi.Schedule {
	GinkgoHelper()
	items := make([]shiftbellapi.Schedule, 0)
	for href != "" {
		collection, err := client.GetSchedules(ctx, href)
		Expect(err).NotTo(HaveOccurred())
		items = append(items, collection.Items...)
		href = collection.Links.Href(shiftbellapi.RelationNext)
	}
	return items
}
