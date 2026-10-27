package acceptance_test

import (
	"context"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/testsupport/shiftbellapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
)

func getManualScheduledChoreForm(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	collection shiftbellapi.ChoreCollection,
) shiftbellapi.ChoreCreationStep {
	GinkgoHelper()
	recurrence := getManualChoreRecurrence(ctx, client, collection)

	By("choosing scheduled recurrence and retrieving the manual form")
	form, err := client.GetChoreCreationStep(ctx, recurrence.Choices[1].Href)
	Expect(err).NotTo(HaveOccurred())
	Expect(form.StatusCode).To(Equal(http.StatusOK))
	Expect(form.ErrorResponse).To(BeNil())
	Expect(form.SuccessResponse).NotTo(BeNil())
	Expect(*form.SuccessResponse).To(gstruct.MatchAllFields(gstruct.Fields{
		"Step":     Equal("form"),
		"Template": BeNil(),
		"Choices":  BeEmpty(),
		"Actions": ConsistOf(
			shiftbellapi.Relation{Rel: shiftbellapi.RelationCreate, Href: "/chores"},
		),
	}))
	return *form.SuccessResponse
}

func getTemplateScheduledChoreForm(
	ctx context.Context,
	client *shiftbellapi.APIClient,
	collection shiftbellapi.ChoreCollection,
	template shiftbellapi.ChoreTemplate,
) shiftbellapi.ChoreCreationStep {
	GinkgoHelper()
	recurrence := getTemplateChoreRecurrence(ctx, client, collection, template)

	By("choosing scheduled recurrence and retrieving the template-based form")
	form, err := client.GetChoreCreationStep(ctx, recurrence.Choices[1].Href)
	Expect(err).NotTo(HaveOccurred())
	Expect(form.StatusCode).To(Equal(http.StatusOK))
	Expect(form.ErrorResponse).To(BeNil())
	Expect(form.SuccessResponse).NotTo(BeNil())
	Expect(*form.SuccessResponse).To(gstruct.MatchAllFields(gstruct.Fields{
		"Step": Equal("form"),
		"Template": Equal(&shiftbellapi.ChoreCreationTemplate{
			Id: template.Id, Name: template.Name, Description: template.Description,
		}),
		"Choices": BeEmpty(),
		"Actions": ConsistOf(
			shiftbellapi.Relation{Rel: shiftbellapi.RelationCreate, Href: "/chores"},
		),
	}))
	return *form.SuccessResponse
}
