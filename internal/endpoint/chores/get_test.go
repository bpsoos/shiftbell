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
			"_links": [{"rel": "collection", "href": "/chores"}],
			"_actions": []
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
			"_links": [
				{"rel": "self", "href": "/chores/42"},
				{"rel": "collection", "href": "/chores"}
			],
			"_actions": [
				{"rel": "edit", "href": "/chores/42"},
				{"rel": "complete", "href": "/chores/42/completion"},
				{"rel": "delete", "href": "/chores/42"}
			]
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
			"_links": [
				{"rel": "self", "href": "/chores/42"},
				{"rel": "collection", "href": "/chores"}
			],
			"_actions": [
				{"rel": "correct_completion", "href": "/chores/42/completion"},
				{"rel": "delete", "href": "/chores/42"}
			]
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
				Links: api.Relations{
					{Rel: "self", Href: "/chores/42"},
					{Rel: "collection", Href: "/chores"},
				},
			},
			Actions: api.Relations{
				{Rel: "edit", Href: "/chores/42"},
				{Rel: "complete", Href: "/chores/42/completion"},
				{Rel: "delete", Href: "/chores/42"},
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
