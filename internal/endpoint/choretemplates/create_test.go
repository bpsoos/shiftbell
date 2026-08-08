package choretemplates_test

import (
	"net/http"
	"net/http/httptest"

	choretemplatesendpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create chore template", func() {
	It("rejects an HTML response", func(ctx SpecContext) {
		handler := choretemplatesendpoint.NewHandler(&choretemplatesendpoint.HandlerDeps{
			Service: NewMockService(GinkgoT()),
		})
		e := echo.New()
		e.POST("/chore-templates", handler.Create)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPost,
			"/chore-templates",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Body.String()).To(BeEmpty())
	})
})
