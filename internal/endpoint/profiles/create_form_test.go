package profiles_test

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateForm", func() {
	DescribeTable(
		"renders the creation form",
		func(ctx SpecContext, hx string, fullPage bool) {
			handler, _, view := newHandler()
			view.On("Form", viewmodels.Form{ActionHref: "/profiles", CancelHref: "/profiles"}, fullPage).
				Return(templ.Raw("new profile")).
				Once()
			response := perform(
				ctx,
				handler.CreateForm,
				http.MethodGet,
				"/profiles/new",
				"text/html",
				"",
				"",
				hx,
			)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("new profile"))
		},
		Entry("page", "", true),
		Entry("fragment", "true", false),
	)

	It("rejects JSON for the HTML-only form route", func(ctx SpecContext) {
		handler, _, _ := newHandler()
		response := perform(
			ctx,
			handler.CreateForm,
			http.MethodGet,
			"/profiles/new",
			hypermedia.MediaType,
			"",
			"",
			"",
		)
		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
	})
})
