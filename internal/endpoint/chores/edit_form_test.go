package chores_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/a-h/templ"
	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Edit chore form", func() {
	It("rejects an ordinary HTML request", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.GET("/chores/:id/edit", handler.Edit)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/42/edit",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Body.String()).To(BeEmpty())
	})

	It("renders an active one-off chore into an edit form", func(ctx SpecContext) {
		deadline := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
		service := NewMockService(GinkgoT())
		service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
			Id:          42,
			Status:      choremodels.ChoreStatusActive,
			Name:        "Kitchen",
			Description: "Wash and fold",
			Deadline:    deadline,
		}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().EditForm(choreviewmodels.EditForm{
			ActionHref: "/chores/42",
			CancelHref: "/chores/42",
			Name: choreviewmodels.Field{
				Value: "Kitchen",
			},
			Description: choreviewmodels.Field{
				Value: "Wash and fold",
			},
			Deadline: choreviewmodels.Field{
				Value: "2026-08-15",
			},
		}, false).Return(templ.Raw("edit form sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.GET("/chores/:id/edit", handler.Edit)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/42/edit",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("edit form sentinel"))
	})

	It(
		"prepares a scheduled chore form with a read-only deadline",
		func(ctx SpecContext) {
			scheduleId := 7
			deadline := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
				Id:          42,
				ScheduleId:  &scheduleId,
				Status:      choremodels.ChoreStatusActive,
				Name:        "Kitchen",
				Description: "Wash and fold",
				Deadline:    deadline,
			}, nil).Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().EditForm(choreviewmodels.EditForm{
				ActionHref: "/chores/42",
				CancelHref: "/chores/42",
				Scheduled:  true,
				Name: choreviewmodels.Field{
					Value: "Kitchen",
				},
				Description: choreviewmodels.Field{
					Value: "Wash and fold",
				},
				Deadline: choreviewmodels.Field{
					Value: "2026-08-15",
				},
			}, false).Return(templ.Raw("scheduled edit form sentinel")).Once()
			handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
				Service: service,
				View:    view,
			})
			e := echo.New()
			e.GET("/chores/:id/edit", handler.Edit)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/42/edit",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("scheduled edit form sentinel"))
		},
	)

	It("rejects a completed chore", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
			Id:     42,
			Status: choremodels.ChoreStatusCompleted,
		}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().Error(choreviewmodels.Error{
			Message: "completed chore cannot be edited",
			Links: []choreviewmodels.Link{
				{Label: "Back to chore", Href: "/chores/42"},
				{Label: "Back to chores", Href: "/chores"},
			},
		}, false).Return(templ.Raw("completed error sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.GET("/chores/:id/edit", handler.Edit)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/42/edit",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.String()).To(Equal("completed error sentinel"))
	})

	It("returns not found for a missing chore", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Get(ctx, 42).Return(nil, errors.Join(
			errors.New("get chore"),
			choremodels.ErrNotFound,
		)).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().Error(choreviewmodels.Error{
			Message: "chore not found",
			Links: []choreviewmodels.Link{
				{Label: "Back to chores", Href: "/chores"},
			},
		}, false).Return(templ.Raw("missing error sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.GET("/chores/:id/edit", handler.Edit)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/42/edit",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotFound))
		Expect(response.Body.String()).To(Equal("missing error sentinel"))
	})
})
