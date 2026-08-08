package chores_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/a-h/templ"
	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choreapimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get chore", func() {
	It("returns recovery controls when the chore is missing", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Get(ctx, 42).Return(
			nil,
			errors.Join(errors.New("get chore"), choremodels.ErrNotFound),
		).Once()
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.GET("/chores/:id", handler.Get)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/42",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotFound))
		Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"error": "chore not found",
			"_links": {
				"collection": {"href": "/chores"}
			},
			"_actions": {}
		}`))
	})

	It(
		"returns an active manual one-off with its mutation controls",
		func(ctx SpecContext) {
			service := NewMockService(GinkgoT())
			service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
				Id:          42,
				Status:      choremodels.ChoreStatusActive,
				Name:        "Kitchen",
				Description: "Wash and fold.",
				Deadline: time.Date(
					2020,
					time.February,
					3,
					0,
					0,
					0,
					0,
					time.UTC,
				),
			}, nil).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			e := echo.New()
			e.GET("/chores/:id", handler.Get)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/42",
				nil,
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"id": 42,
			"schedule_id": null,
			"status": "active",
			"name": "Kitchen",
			"description": "Wash and fold.",
			"deadline": "2020-02-03",
			"completed_on": null,
			"_links": {
				"self": {"href": "/chores/42"},
				"collection": {"href": "/chores"}
			},
			"_actions": {
				"edit": {
					"href": "/chores/42",
					"method": "PATCH",
					"content_type": "application/json",
					"fields": [
						{"name": "name", "type": "string", "required": true},
						{"name": "description", "type": "string", "required": false},
						{"name": "deadline", "type": "date", "required": true}
					]
				},
				"complete": {
					"href": "/chores/42/completion",
					"method": "PUT",
					"content_type": "application/json",
					"fields": [
						{"name": "completed_on", "type": "date", "required": true}
					]
				},
				"delete": {
					"href": "/chores/42",
					"method": "DELETE",
					"content_type": "",
					"fields": null
				}
			}
		}`))
		},
	)

	It("returns a completed one-off with its mutation controls", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
			Id:          42,
			Status:      choremodels.ChoreStatusCompleted,
			Name:        "Kitchen",
			Description: "Wash and fold.",
			Deadline:    time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC),
			CompletedOn: time.Date(2020, time.February, 4, 0, 0, 0, 0, time.UTC),
		}, nil).Once()
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.GET("/chores/:id", handler.Get)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/42",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"id": 42,
			"schedule_id": null,
			"status": "completed",
			"name": "Kitchen",
			"description": "Wash and fold.",
			"deadline": "2020-02-03",
			"completed_on": "2020-02-04",
			"_links": {
				"self": {"href": "/chores/42"},
				"collection": {"href": "/chores"}
			},
			"_actions": {
				"correct_completion": {
					"href": "/chores/42/completion",
					"method": "PATCH",
					"content_type": "application/json",
					"fields": [
						{"name": "completed_on", "type": "date", "required": true}
					]
				},
				"delete": {
					"href": "/chores/42",
					"method": "DELETE",
					"content_type": "",
					"fields": null
				}
			}
		}`))
	})

	It("renders an active chore as read-only HTML", func(ctx SpecContext) {
		deadline := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
		service := NewMockService(GinkgoT())
		service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
			Id:          42,
			Status:      choremodels.ChoreStatusActive,
			Name:        "Kitchen",
			Description: "Wash and fold",
			Deadline:    deadline,
		}, nil).Once()
		chore := choreapimodels.Representation{
			Response: choreapimodels.Response{
				Id:          42,
				Status:      choremodels.ChoreStatusActive,
				Name:        "Kitchen",
				Description: "Wash and fold",
				Deadline:    "2026-08-15",
				Links: map[string]api.Link{
					"self":       {Href: "/chores/42"},
					"collection": {Href: "/chores"},
				},
			},
			Actions: map[string]api.Action{
				"edit": {
					Href:        "/chores/42",
					Method:      http.MethodPatch,
					ContentType: "application/json",
					Fields: []api.ActionField{
						{Name: "name", Type: "string", Required: true},
						{Name: "description", Type: "string"},
						{Name: "deadline", Type: "date", Required: true},
					},
				},
				"complete": {
					Href:        "/chores/42/completion",
					Method:      http.MethodPut,
					ContentType: "application/json",
					Fields: []api.ActionField{
						{Name: "completed_on", Type: "date", Required: true},
					},
				},
				"delete": {Href: "/chores/42", Method: http.MethodDelete},
			},
		}
		view := NewMockView(GinkgoT())
		view.EXPECT().Detail(choreviewmodels.Detail{
			Chore:    chore,
			BackHref: "/chores",
		}, true).Return(templ.Raw("detail sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.GET("/chores/:id", handler.Get)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores/42", nil)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("detail sentinel"))
	})
})
