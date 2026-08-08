package chores_test

import (
	"net/http"
	"net/http/httptest"

	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Start chore creation", func() {
	It("offers manual and template sources", func(ctx SpecContext) {
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{})
		e := echo.New()
		e.GET("/chores/new", handler.New)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores/new", nil)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"step": "source",
			"choices": [
				{"label": "Specify new", "href": "/chores/new?source=manual"},
				{"label": "Select template", "href": "/chore-templates?picker=1"}
			]
		}`))
	})

	It(
		"offers one-off and scheduled recurrence for a manual source",
		func(ctx SpecContext) {
			handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{})
			e := echo.New()
			e.GET("/chores/new", handler.New)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/new?source=manual",
				nil,
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"step": "recurrence",
			"choices": [
				{
					"label": "One-off",
					"href": "/chores/new?source=manual&recurrence=one-off"
				},
				{
					"label": "Scheduled",
					"href": "/chores/new?source=manual&recurrence=scheduled"
				}
			]
		}`))
		},
	)

	It("returns Not Implemented for manual scheduled recurrence", func(ctx SpecContext) {
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{})
		e := echo.New()
		e.GET("/chores/new", handler.New)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/new?source=manual&recurrence=scheduled",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotImplemented))
		Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"error": "scheduled recurrence is not implemented",
			"_links": null,
			"_actions": null
		}`))
	})

	It("returns the manual one-off form", func(ctx SpecContext) {
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{})
		e := echo.New()
		e.GET("/chores/new", handler.New)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/new?source=manual&recurrence=one-off",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"step": "form",
			"fields": [
				{"name": "name", "type": "string", "required": true},
				{"name": "description", "type": "string", "required": false},
				{"name": "deadline", "type": "date", "required": true},
				{
					"name": "save_as_chore_template",
					"type": "boolean",
					"required": false
				}
			],
			"action": {
				"href": "/chores",
				"method": "POST",
				"content_type": "application/json",
				"fields": null
			}
		}`))
	})

	It("offers recurrence choices for a selected template", func(ctx SpecContext) {
		choreTemplateService := NewMockChoreTemplateService(GinkgoT())
		choreTemplateService.EXPECT().
			Get(ctx, 42).
			Return(&choretemplatemodels.ChoreTemplateDetails{
				ChoreTemplate: choretemplatemodels.ChoreTemplate{
					Id:          42,
					Name:        "Kitchen",
					Description: "Reusable template steps.",
				},
			}, nil).
			Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			ChoreTemplateService: choreTemplateService,
		})
		e := echo.New()
		e.GET("/chores/new", handler.New)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/new?template_id=42",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"step": "recurrence",
			"template": {
				"id": 42,
				"name": "Kitchen",
				"description": "Reusable template steps."
			},
			"choices": [
				{
					"label": "One-off",
					"href": "/chores/new?template_id=42&recurrence=one-off"
				},
				{
					"label": "Scheduled",
					"href": "/chores/new?template_id=42&recurrence=scheduled"
				}
			]
		}`))
	})

	It(
		"returns Not Implemented for template-based scheduled recurrence",
		func(ctx SpecContext) {
			choreTemplateService := NewMockChoreTemplateService(GinkgoT())
			choreTemplateService.EXPECT().
				Get(ctx, 42).
				Return(&choretemplatemodels.ChoreTemplateDetails{
					ChoreTemplate: choretemplatemodels.ChoreTemplate{
						Id:          42,
						Name:        "Kitchen",
						Description: "Reusable template steps.",
					},
				}, nil).
				Once()
			handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
				ChoreTemplateService: choreTemplateService,
			})
			e := echo.New()
			e.GET("/chores/new", handler.New)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/new?template_id=42&recurrence=scheduled",
				nil,
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusNotImplemented))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
				"error": "scheduled recurrence is not implemented",
				"_links": null,
				"_actions": null
			}`))
		},
	)

	It("returns the template-based one-off form", func(ctx SpecContext) {
		choreTemplateService := NewMockChoreTemplateService(GinkgoT())
		choreTemplateService.EXPECT().
			Get(ctx, 42).
			Return(&choretemplatemodels.ChoreTemplateDetails{
				ChoreTemplate: choretemplatemodels.ChoreTemplate{
					Id:          42,
					Name:        "Kitchen",
					Description: "Reusable template steps.",
				},
			}, nil).
			Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			ChoreTemplateService: choreTemplateService,
		})
		e := echo.New()
		e.GET("/chores/new", handler.New)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/new?template_id=42&recurrence=one-off",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"step": "form",
			"template": {
				"id": 42,
				"name": "Kitchen",
				"description": "Reusable template steps."
			},
			"fields": [
				{"name": "deadline", "type": "date", "required": true}
			],
			"action": {
				"href": "/chores",
				"method": "POST",
				"content_type": "application/json",
				"fields": null
			}
		}`))
	})
})
