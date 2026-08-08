package hypermedia_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Response negotiation", func() {
	It("selects the Shiftbell JSON representation when explicitly requested", func() {
		request := httptest.NewRequest(http.MethodGet, "/chores", nil)
		request.Header.Set("Accept", hypermedia.MediaType)

		Expect(hypermedia.Negotiate(request)).To(Equal(hypermedia.RepresentationJSON))
	})

	It("selects HTML when no representation is requested", func() {
		request := httptest.NewRequest(http.MethodGet, "/chores", nil)

		Expect(hypermedia.Negotiate(request)).To(Equal(hypermedia.RepresentationHTML))
	})

	It("selects HTML for a browser Accept header", func() {
		request := httptest.NewRequest(http.MethodGet, "/chores", nil)
		request.Header.Set(
			"Accept",
			"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		)

		Expect(hypermedia.Negotiate(request)).To(Equal(hypermedia.RepresentationHTML))
	})

	It("selects HTML when any representation is acceptable", func() {
		request := httptest.NewRequest(http.MethodGet, "/chores", nil)
		request.Header.Set("Accept", "*/*")

		Expect(hypermedia.Negotiate(request)).To(Equal(hypermedia.RepresentationHTML))
	})

	It("selects the acceptable representation with the higher quality", func() {
		request := httptest.NewRequest(http.MethodGet, "/chores", nil)
		request.Header.Set(
			"Accept",
			"text/html;q=0.2, "+hypermedia.MediaType+";q=0.8",
		)

		Expect(hypermedia.Negotiate(request)).To(Equal(hypermedia.RepresentationJSON))
	})

	It("selects HTML when it has the higher quality", func() {
		request := httptest.NewRequest(http.MethodGet, "/chores", nil)
		request.Header.Set(
			"Accept",
			hypermedia.MediaType+";q=0.4, text/html;q=0.9",
		)

		Expect(hypermedia.Negotiate(request)).To(Equal(hypermedia.RepresentationHTML))
	})

	It("rejects supported representations that have zero quality", func() {
		request := httptest.NewRequest(http.MethodGet, "/chores", nil)
		request.Header.Set(
			"Accept",
			hypermedia.MediaType+";q=0, text/html;q=0",
		)

		Expect(
			hypermedia.Negotiate(request),
		).To(Equal(hypermedia.RepresentationUnsupported))
	})

	It("reports unsupported media types", func() {
		request := httptest.NewRequest(http.MethodGet, "/chores", nil)
		request.Header.Set("Accept", "application/json")

		Expect(
			hypermedia.Negotiate(request),
		).To(Equal(hypermedia.RepresentationUnsupported))
	})
})
