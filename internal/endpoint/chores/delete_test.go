package chores_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete chore", func() {
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
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNoContent))
		Expect(response.Header().Get("Content-Type")).To(BeEmpty())
		Expect(response.Body.String()).To(BeEmpty())
	})
})
