package chores_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/a-h/templ"
	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	choreapimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Start chore creation", func() {
	DescribeTable(
		"renders the manual scheduled HTML form",
		func(ctx SpecContext, hxRequest string, fullPage bool) {
			view := NewMockView(GinkgoT())
			view.EXPECT().
				ScheduledForm(choreviewmodels.ScheduledForm{ActionHref: "/chores", CancelHref: "/chores"}, fullPage).
				Return(templ.Raw("manual scheduled form")).
				Once()
			handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{View: view})
			app := echo.New()
			app.GET("/chores/new", handler.New)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/new?source=manual&recurrence=scheduled",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", hxRequest)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(
				response.Header().Get("Content-Type"),
			).To(Equal(echo.MIMETextHTMLCharsetUTF8))
			Expect(response.Body.String()).To(Equal("manual scheduled form"))
		},
		Entry("full page", "", true),
		Entry("HTMX fragment", "true", false),
	)

	DescribeTable(
		"renders the selected template in the scheduled HTML form",
		func(ctx SpecContext, hxRequest string, fullPage bool) {
			templateService := NewMockChoreTemplateService(GinkgoT())
			templateService.EXPECT().
				Get(ctx, 42).
				Return(&choretemplatemodels.ChoreTemplateDetails{
					Id: 42, Name: "Kitchen", Description: "Reusable steps.",
				}, nil).
				Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().ScheduledForm(choreviewmodels.ScheduledForm{
				ActionHref:   "/chores",
				CancelHref:   "/chores",
				BackHref:     "/chores/new?template_id=42",
				ScheduleName: choreviewmodels.Field{Value: "Kitchen schedule"},
				Template: &choreapimodels.CreationTemplate{
					Id:          42,
					Name:        "Kitchen",
					Description: "Reusable steps.",
				},
			}, fullPage).Return(templ.Raw("template scheduled form")).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{
					View:                 view,
					ChoreTemplateService: templateService,
				},
			)
			app := echo.New()
			app.GET("/chores/new", handler.New)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/new?template_id=42&recurrence=scheduled",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", hxRequest)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(
				response.Header().Get("Content-Type"),
			).To(Equal(echo.MIMETextHTMLCharsetUTF8))
			Expect(response.Body.String()).To(Equal("template scheduled form"))
		},
		Entry("full page", "", true),
		Entry("HTMX fragment", "true", false),
	)

	It("rejects a non-numeric template id", func(ctx SpecContext) {
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{})
		e := echo.New()
		e.GET("/chores/new", handler.New)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/new?template_id=invalid",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(struct {
			Status int
			Body   string
		}{response.Code, response.Body.String()}).To(Equal(struct {
			Status int
			Body   string
		}{
			http.StatusUnprocessableEntity,
			"{\"error\":\"invalid chore template id\",\"_links\":[],\"_actions\":[]}\n",
		}))
	})

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
			],
			"_actions": []
		}`))
	})

	It("maps required source actions into the HTML view", func(ctx SpecContext) {
		view := NewMockView(GinkgoT())
		view.EXPECT().Creation(choreviewmodels.Creation{
			Creation: choreapimodels.CreationResponse{
				Step: "source",
				Choices: []choreapimodels.CreationChoice{
					{
						Label: "Specify new",
						Href:  "/chores/new?source=manual",
					},
					{Label: "Select template", Href: "/chore-templates?picker=1"},
				},
			},
			BackHref:              "/chores",
			CreateFromScratchHref: "/chores/new?recurrence=one-off&source=manual",
			SelectTemplateHref:    "/chore-templates?picker=1",
		}, false).Return(templ.Raw("creation sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{View: view})
		e := echo.New()
		e.GET("/chores/new", handler.New)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores/new", nil)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("creation sentinel"))
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
					"href": "/chores/new?recurrence=one-off&source=manual"
				},
				{
					"label": "Scheduled",
					"href": "/chores/new?recurrence=scheduled&source=manual"
				}
			],
			"_actions": []
		}`))
		},
	)

	It("returns the manual scheduled form through the JSON API", func(ctx SpecContext) {
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

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"step": "form",
			"_actions": [{"rel": "create", "href": "/chores"}]
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
			"_actions": [{"rel": "create", "href": "/chores"}]
		}`))
	})

	It("offers recurrence choices for a selected template", func(ctx SpecContext) {
		choreTemplateService := NewMockChoreTemplateService(GinkgoT())
		choreTemplateService.EXPECT().
			Get(ctx, 42).
			Return(&choretemplatemodels.ChoreTemplateDetails{
				Id:          42,
				Name:        "Kitchen",
				Description: "Reusable template steps.",
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
					"href": "/chores/new?recurrence=one-off&template_id=42"
				},
				{
					"label": "Scheduled",
					"href": "/chores/new?recurrence=scheduled&template_id=42"
				}
			],
			"_actions": []
		}`))
	})

	It(
		"returns the selected template and scheduled form action through the JSON API",
		func(ctx SpecContext) {
			choreTemplateService := NewMockChoreTemplateService(GinkgoT())
			choreTemplateService.EXPECT().
				Get(ctx, 42).
				Return(&choretemplatemodels.ChoreTemplateDetails{
					Id:          42,
					Name:        "Kitchen",
					Description: "Reusable template steps.",
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

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
				"step": "form",
				"template": {
					"id": 42,
					"name": "Kitchen",
					"description": "Reusable template steps."
				},
				"_actions": [{"rel": "create", "href": "/chores"}]
			}`))
		},
	)

	It("returns the template-based one-off form", func(ctx SpecContext) {
		choreTemplateService := NewMockChoreTemplateService(GinkgoT())
		choreTemplateService.EXPECT().
			Get(ctx, 42).
			Return(&choretemplatemodels.ChoreTemplateDetails{
				Id:          42,
				Name:        "Kitchen",
				Description: "Reusable template steps.",
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
			"_actions": [{"rel": "create", "href": "/chores"}]
		}`))
	})

	DescribeTable(
		"rejects a missing source template",
		func(ctx SpecContext, recurrence string) {
			templateService := NewMockChoreTemplateService(GinkgoT())
			templateService.EXPECT().
				Get(ctx, 42).
				Return(nil, choretemplatemodels.ErrNotFound).
				Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{ChoreTemplateService: templateService},
			)
			app := echo.New()
			app.GET("/chores/new", handler.New)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/new?template_id=42&recurrence="+recurrence,
				nil,
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusNotFound))
			Expect(
				response.Body.Bytes(),
			).To(MatchJSON(`{"error":"chore template not found","_links":[],"_actions":[]}`))
		},
		Entry("one-off", "one-off"),
		Entry("scheduled", "scheduled"),
	)

	DescribeTable(
		"rejects a deactivated source template",
		func(ctx SpecContext, recurrence string) {
			deactivatedAt := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			templateService := NewMockChoreTemplateService(GinkgoT())
			templateService.EXPECT().
				Get(ctx, 42).
				Return(&choretemplatemodels.ChoreTemplateDetails{
					Id: 42, Name: "Kitchen", DeactivatedAt: &deactivatedAt,
				}, nil).
				Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{ChoreTemplateService: templateService},
			)
			app := echo.New()
			app.GET("/chores/new", handler.New)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/new?template_id=42&recurrence="+recurrence,
				nil,
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
			Expect(
				response.Body.Bytes(),
			).To(MatchJSON(`{"error":"chore template inactive","_links":[],"_actions":[]}`))
		},
		Entry("one-off", "one-off"),
		Entry("scheduled", "scheduled"),
	)
})
