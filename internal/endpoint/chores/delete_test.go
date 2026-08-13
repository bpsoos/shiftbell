package chores_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/a-h/templ"
	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	confirmationviewmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete chore", func() {
	It("rejects an HTML representation", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.DELETE("/chores/:id", handler.Delete)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodDelete,
			"/chores/42",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Body.String()).To(BeEmpty())
	})

	It("permanently deletes the chore", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Delete(ctx, 42).Return(nil).Once()
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.DELETE("/chores/:id", handler.Delete)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodDelete,
			"/chores/42",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNoContent))
		Expect(response.Header().Get("Content-Type")).To(BeEmpty())
		Expect(response.Body.String()).To(BeEmpty())
	})

	It("returns no content when the chore is already missing", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Delete(ctx, 42).Return(
			errors.Join(errors.New("delete chore"), choremodels.ErrNotFound),
		).Once()
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.DELETE("/chores/:id", handler.Delete)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodDelete,
			"/chores/42",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNoContent))
		Expect(response.Header().Get("Content-Type")).To(BeEmpty())
		Expect(response.Body.String()).To(BeEmpty())
	})

	It("deletes a chore from an HTMX dialog and redirects", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Delete(ctx, 42).Return(nil).Once()
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.DELETE("/chores/:id", handler.Delete)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodDelete,
			"/chores/42",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Header().Get("HX-Redirect")).To(Equal("/chores"))
		Expect(response.Header().Get("HX-Trigger")).To(Equal("choreDeleted"))
		cookies := response.Result().Cookies()
		Expect(cookies).To(HaveLen(1))
		Expect(cookies[0].Name).To(Equal("shiftbell_flash"))
		Expect(cookies[0].Value).To(Equal("chore-deleted"))
	})

	It("keeps HTMX deletion idempotent when the chore is missing", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Delete(ctx, 42).Return(
			errors.Join(errors.New("delete chore"), choremodels.ErrNotFound),
		).Once()
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.DELETE("/chores/:id", handler.Delete)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodDelete,
			"/chores/42",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Header().Get("HX-Redirect")).To(Equal("/chores"))
	})

	It("keeps the dialog open when HTMX deletion fails", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Delete(ctx, 42).Return(errors.New("database unavailable")).Once()
		service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
			Id:     42,
			Status: choremodels.ChoreStatusActive,
			Name:   "Kitchen",
		}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().ConfirmationDialog(confirmationviewmodels.Dialog{
			Heading:      "Delete chore?",
			Prompt:       "Delete",
			Name:         "Kitchen",
			Supporting:   []string{"This cannot be undone."},
			ActionHref:   "/chores/42",
			ActionMethod: "delete",
			ActionLabel:  "Delete permanently",
			Error:        "The chore could not be deleted. Try again.",
			Icon:         "trash",
		}).Return(templ.Raw("deletion error sentinel")).Once()
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service, View: view},
		)
		e := echo.New()
		e.DELETE("/chores/:id", handler.Delete)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodDelete,
			"/chores/42",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusInternalServerError))
		Expect(response.Header().Get("HX-Redirect")).To(BeEmpty())
		Expect(response.Body.String()).To(Equal("deletion error sentinel"))
	})
})
