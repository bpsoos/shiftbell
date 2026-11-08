package chores_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Get chore", func() {
	DescribeTable(
		"preserves a local collection return link",
		func(ctx SpecContext, returnTo, expected string) {
			service := NewMockService(GinkgoT())
			service.EXPECT().
				Get(ctx, 42).
				Return(&choremodels.Chore{Id: 42, Status: choremodels.ChoreStatusActive}, nil).
				Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().
				Detail(mock.Anything, true).
				Run(func(model choreviewmodels.Detail, fullPage bool) {
					Expect(model.BackHref).To(Equal(expected))
				}).
				Return(templ.Raw("detail")).
				Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service, View: view},
			)
			app := echo.New()
			app.GET("/chores/:id", handler.Get)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/42?return_to="+url.QueryEscape(returnTo),
				nil,
			)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
		},
		Entry(
			"mine with search and pagination",
			"/chores?scope=mine&search=Kitchen&offset=20",
			"/chores?scope=mine&search=Kitchen&offset=20",
		),
		Entry(
			"completed chores",
			"/chores?scope=mine&status=completed",
			"/chores?scope=mine&status=completed",
		),
		Entry("external URL", "https://example.net/chores?scope=mine", "/chores"),
		Entry("another resource", "/profiles", "/chores"),
	)

	DescribeTable(
		"includes the owner in HTML details without listing profiles",
		func(ctx SpecContext, status choremodels.ChoreStatus) {
			service := NewMockService(GinkgoT())
			view := NewMockView(GinkgoT())
			service.EXPECT().
				Get(ctx, 42).
				Return(&choremodels.Chore{Id: 42, Status: status, OwnerProfileId: new(7), OwnerProfileName: "Alex"}, nil).
				Once()
			view.EXPECT().
				Detail(mock.Anything, true).
				Run(func(model choreviewmodels.Detail, fullPage bool) {
					Expect(model.Chore.OwnerProfileId).To(Equal(new(7)))
					Expect(model.Chore.OwnerProfileName).To(Equal("Alex"))
					Expect(
						model.Chore.Actions.Href("set_assignment"),
					).To(Equal("/chores/42/assignment"))
				}).
				Return(templ.Raw("owned chore")).
				Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{
					Service:           service,
					View:              view,
					ProfileService:    NewMockProfileService(GinkgoT()),
					ProfilesUIEnabled: true,
				},
			)
			app := echo.New()
			app.GET("/chores/:id", handler.Get)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/42",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("owned chore"))
		},
		Entry("active", choremodels.ChoreStatusActive),
		Entry("completed", choremodels.ChoreStatusCompleted),
	)

	It("rejects a non-numeric chore id", func(ctx SpecContext) {
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: NewMockService(GinkgoT())},
		)
		e := echo.New()
		e.GET("/chores/:id", handler.Get)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/invalid",
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
			"{\"error\":\"invalid chore id\",\"_links\":[],\"_actions\":[]}\n",
		}))
	})

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
			"owner_profile_id": null,
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
				{"rel": "set_assignment", "href": "/chores/42/assignment"},
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
			"owner_profile_id": null,
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
				{"rel": "set_assignment", "href": "/chores/42/assignment"},
				{"rel": "correct_completion", "href": "/chores/42/completion"},
				{"rel": "delete", "href": "/chores/42"}
			]
		}`))
	})

	It(
		"returns an active scheduled chore without a delete action",
		func(ctx SpecContext) {
			scheduleId := 7
			service := NewMockService(GinkgoT())
			service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
				Id:          42,
				ScheduleId:  &scheduleId,
				Status:      choremodels.ChoreStatusActive,
				Name:        "Kitchen",
				Description: "Wash and fold.",
				Deadline:    time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC),
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
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"id": 42,
			"schedule_id": 7,
			"owner_profile_id": null,
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
				{"rel": "set_assignment", "href": "/chores/42/assignment"},
				{"rel": "edit", "href": "/chores/42"},
				{"rel": "complete", "href": "/chores/42/completion"}
			]
		}`))
		},
	)

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
			Id:          42,
			Status:      choremodels.ChoreStatusActive,
			Name:        "Kitchen",
			Description: "Wash and fold",
			Deadline:    "2026-08-15",
			Links: api.Relations{
				{Rel: "self", Href: "/chores/42"},
				{Rel: "collection", Href: "/chores"},
			},
			Actions: api.Relations{
				{Rel: "set_assignment", Href: "/chores/42/assignment"},
				{Rel: "edit", Href: "/chores/42"},
				{Rel: "complete", Href: "/chores/42/completion"},
				{Rel: "delete", Href: "/chores/42"},
			},
		}
		view := NewMockView(GinkgoT())
		view.EXPECT().Detail(choreviewmodels.Detail{
			Chore:      chore,
			BackHref:   "/chores",
			EditHref:   "/chores/42/edit",
			DeleteHref: "/chores/42/deletion",
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
