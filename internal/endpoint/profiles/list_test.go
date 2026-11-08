package profiles_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/profiles"
	"github.com/bpsoos/shiftbell/internal/models/api"
	apimodels "github.com/bpsoos/shiftbell/internal/models/api/profiles"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("List", func() {
	It("returns profiles and discoverable actions as JSON", func(ctx SpecContext) {
		handler, service, _ := newHandler()
		service.On("List", ctx).
			Return([]models.Profile{{Id: 7, Name: "Alex"}}, nil).
			Once()
		e := echo.New()
		e.GET("/profiles", handler.List)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/profiles", nil)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()
		e.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
		Expect(response.Header().Get("Vary")).To(Equal("Accept"))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"items":[{"id":7,"name":"Alex","_links":[{"rel":"collection","href":"/profiles"}],"_actions":[{"rel":"delete","href":"/profiles/7"}]}],
			"_links":[{"rel":"self","href":"/profiles"},{"rel":"setup","href":"/setup"}],
			"_actions":[{"rel":"create","href":"/profiles"}]
		}`))
	})

	It("returns an empty array for an empty installation", func(ctx SpecContext) {
		handler, service, _ := newHandler()
		service.On("List", ctx).Return(nil, nil).Once()
		e := echo.New()
		e.GET("/profiles", handler.List)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/profiles", nil)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()
		e.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(ContainSubstring(`"items":[]`))
	})

	DescribeTable(
		"renders an HTML collection",
		func(ctx SpecContext, hx string, fullPage bool) {
			handler, service, view := newHandler()
			service.On("List", ctx).
				Return([]models.Profile{{Id: 7, Name: "Alex"}}, nil).
				Once()
			view.On("Collection", viewmodels.Collection{
				Profiles: []apimodels.Response{
					{
						Id:      7,
						Name:    "Alex",
						Links:   api.Relations{{Rel: "collection", Href: "/profiles"}},
						Actions: api.Relations{{Rel: "delete", Href: "/profiles/7"}},
					},
				},
				CreateHref: "/profiles/new",
			}, fullPage).Return(templ.Raw("profile list")).Once()
			e := echo.New()
			e.GET("/profiles", handler.List)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/profiles",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", hx)
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(
				response.Header().Get("Content-Type"),
			).To(Equal("text/html; charset=UTF-8"))
			Expect(response.Header().Get("Vary")).To(Equal("Accept, HX-Request"))
			Expect(response.Body.String()).To(Equal("profile list"))
		},
		Entry("page", "", true),
		Entry("fragment", "true", false),
	)

	It("hides storage errors in JSON", func(ctx SpecContext) {
		handler, service, _ := newHandler()
		service.On("List", ctx).Return(nil, errors.New("private database path")).Once()
		e := echo.New()
		e.GET("/profiles", handler.List)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/profiles", nil)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()
		e.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusInternalServerError))
		Expect(response.Body.String()).NotTo(ContainSubstring("private database path"))
	})

	It("renders storage errors as HTML", func(ctx SpecContext) {
		handler, service, view := newHandler()
		service.On("List", ctx).Return(nil, errors.New("storage failed")).Once()
		view.On("Error", viewmodels.Error{Message: "Something went wrong. Try again.", CollectionHref: "/profiles"}, true).
			Return(templ.Raw("error")).
			Once()
		e := echo.New()
		e.GET("/profiles", handler.List)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/profiles", nil)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()
		e.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusInternalServerError))
		Expect(response.Body.String()).To(Equal("error"))
	})

	It(
		"rejects unsupported representations before accessing storage",
		func(ctx SpecContext) {
			handler, _, _ := newHandler()
			e := echo.New()
			e.GET("/profiles", handler.List)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/profiles",
				nil,
			)
			request.Header.Set("Accept", "application/xml")
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		},
	)

	It(
		"does not advertise HTML support until a renderer is supplied",
		func(ctx SpecContext) {
			handler := profiles.NewHandler(
				&profiles.HandlerDeps{Service: &MockService{}},
				&profiles.HandlerConfig{UIEnabled: true},
			)
			e := echo.New()
			e.GET("/profiles", handler.List)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/profiles",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		},
	)
})
