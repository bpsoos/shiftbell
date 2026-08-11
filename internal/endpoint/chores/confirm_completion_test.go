package chores_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/a-h/templ"
	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Confirm chore completion", func() {
	It("rejects an ordinary HTML request", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.GET("/chores/:id/completion", handler.ConfirmCompletion)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/42/completion",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Body.String()).To(BeEmpty())
		Expect(response.Header().Get("Vary")).To(Equal("Accept, HX-Request"))
	})

	It("rejects a vendor representation from HTMX", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.GET("/chores/:id/completion", handler.ConfirmCompletion)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/42/completion",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Body.String()).To(BeEmpty())
	})

	It("loads the chore and renders only its dialog fragment", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
			Id:     42,
			Status: choremodels.ChoreStatusActive,
			Name:   "Kitchen",
		}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().CompletionDialog(choreviewmodels.CompletionDialog{
			Name:       "Kitchen",
			ActionHref: "/chores/42/completion",
		}).Return(templ.Raw("dialog sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.GET("/chores/:id/completion", handler.ConfirmCompletion)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/42/completion",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(
			response.Header().Get("Content-Type"),
		).To(Equal("text/html; charset=UTF-8"))
		Expect(response.Body.String()).To(Equal("dialog sentinel"))
	})
})
