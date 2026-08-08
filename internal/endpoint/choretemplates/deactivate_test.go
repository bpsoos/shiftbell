package choretemplates_test

import (
	"net/http"
	"net/http/httptest"

	choretemplatesendpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deactivate chore template", func() {
	It("rejects an HTML response", func(ctx SpecContext) {
		handler := choretemplatesendpoint.NewHandler(&choretemplatesendpoint.HandlerDeps{
			Service: NewMockService(GinkgoT()),
		})
		e := echo.New()
		e.PUT("/chore-templates/:id/deactivation", handler.Deactivate)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPut,
			"/chore-templates/8/deactivation",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Body.String()).To(BeEmpty())
	})
})
