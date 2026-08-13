package chores_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/a-h/templ"
	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	confirmationviewmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Confirm chore deletion", func() {
	It("rejects an ordinary HTML request", func(ctx SpecContext) {
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: NewMockService(GinkgoT())},
		)
		e := echo.New()
		e.GET("/chores/:id/deletion", handler.ConfirmDeletion)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/42/deletion",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Body.String()).To(BeEmpty())
		Expect(response.Header().Get("Vary")).To(Equal("Accept, HX-Request"))
	})

	It("loads a deletable chore and renders its dialog", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
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
			Icon:         "trash",
		}).Return(templ.Raw("deletion dialog sentinel")).Once()
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service, View: view},
		)
		e := echo.New()
		e.GET("/chores/:id/deletion", handler.ConfirmDeletion)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/42/deletion",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("deletion dialog sentinel"))
	})

	It("rejects an active scheduled chore", func(ctx SpecContext) {
		scheduleId := 7
		service := NewMockService(GinkgoT())
		service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
			Id:         42,
			ScheduleId: &scheduleId,
			Status:     choremodels.ChoreStatusActive,
		}, nil).Once()
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.GET("/chores/:id/deletion", handler.ConfirmDeletion)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/42/deletion",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.String()).To(BeEmpty())
	})
})
