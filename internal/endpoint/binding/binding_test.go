package binding_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/bpsoos/shiftbell/internal/endpoint/binding"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type requestBody struct {
	Name    string `form:"name"    json:"name"`
	Enabled bool   `form:"enabled" json:"enabled"`
	Count   *int   `form:"count"   json:"count"`
}

var _ = Describe("Request binding", func() {
	It("binds a JSON request", func() {
		var body requestBody
		response := serveBindingRequest(
			http.MethodPost,
			`{"name":"Kitchen","enabled":true,"count":7}`,
			echo.MIMEApplicationJSON,
			&body,
		)

		Expect(response.Code).To(Equal(http.StatusNoContent))
		Expect(body.Name).To(Equal("Kitchen"))
		Expect(body.Enabled).To(BeTrue())
		Expect(body.Count).NotTo(BeNil())
		Expect(*body.Count).To(Equal(7))
	})

	It("binds a vendor JSON request", func() {
		var body requestBody
		response := serveBindingRequest(
			http.MethodPost,
			`{"name":"Kitchen"}`,
			hypermedia.MediaType,
			&body,
		)

		Expect(response.Code).To(Equal(http.StatusNoContent))
		Expect(body.Name).To(Equal("Kitchen"))
	})

	It("binds a form request", func() {
		var body requestBody
		response := serveBindingRequest(
			http.MethodPatch,
			"name=Kitchen&enabled=true&count=7",
			echo.MIMEApplicationForm,
			&body,
		)

		Expect(response.Code).To(Equal(http.StatusNoContent))
		Expect(body.Name).To(Equal("Kitchen"))
		Expect(body.Enabled).To(BeTrue())
		Expect(body.Count).NotTo(BeNil())
		Expect(*body.Count).To(Equal(7))
	})

	It("rejects an unsupported content type", func() {
		var body requestBody
		e := echo.New()
		e.POST("/", func(ctx *echo.Context) error {
			err := binding.Bind(ctx, &body)
			Expect(errors.Is(err, binding.ErrUnsupportedMediaType)).To(BeTrue())
			return ctx.NoContent(http.StatusUnsupportedMediaType)
		})
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("Kitchen"))
		request.Header.Set(echo.HeaderContentType, "application/problem+json")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnsupportedMediaType))
	})
})

func serveBindingRequest(
	method string,
	payload string,
	contentType string,
	target any,
) *httptest.ResponseRecorder {
	e := echo.New()
	e.Add(method, "/", func(ctx *echo.Context) error {
		Expect(binding.Bind(ctx, target)).To(Succeed())
		return ctx.NoContent(http.StatusNoContent)
	})
	request := httptest.NewRequest(method, "/", strings.NewReader(payload))
	request.Header.Set(echo.HeaderContentType, contentType)
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)
	return response
}
