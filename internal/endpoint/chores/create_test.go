package chores_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/a-h/templ"
	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create chore", func() {
	It(
		"creates a manual one-off through the advertised form action",
		func(ctx SpecContext) {
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Create(ctx, &choremodels.CreateChoreParams{
				Name:        "  Kitchen  ",
				Description: "  Wash and fold.  ",
				Deadline:    deadline,
			}).Return(&choremodels.CreateChoreResult{
				Chore: &choremodels.Chore{
					Id:          42,
					Status:      choremodels.ChoreStatusActive,
					Name:        "Kitchen",
					Description: "Wash and fold.",
					Deadline:    deadline,
				},
			}, nil).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			e := echo.New()
			e.POST("/chores", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/chores",
				strings.NewReader(`{
				"name": "  Kitchen  ",
				"description": "  Wash and fold.  ",
				"deadline": "2020-02-03",
				"save_as_chore_template": false
			}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusCreated))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Header().Get("Location")).To(Equal("/chores/42"))
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

	It(
		"returns a retry action when saving conflicts with an active template",
		func(ctx SpecContext) {
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Create(ctx, &choremodels.CreateChoreParams{
				Name:                "KITCHEN",
				Description:         "Conflicting description",
				Deadline:            deadline,
				SaveAsChoreTemplate: true,
			}).Return(nil, fmt.Errorf(
				"create manual one-off chore: %w",
				choretemplatemodels.ErrNameConflict,
			)).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			e := echo.New()
			e.POST("/chores", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/chores",
				strings.NewReader(`{
				"name": "KITCHEN",
				"description": "Conflicting description",
				"deadline": "2020-02-03",
				"save_as_chore_template": true
			}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusConflict))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"error": "chore template name conflicts with an active chore template",
			"_links": [],
			"_actions": [{"rel": "create", "href": "/chores"}]
		}`))
		},
	)

	It(
		"returns recovery controls for an inactive source template",
		func(ctx SpecContext) {
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			templateId := 42
			service := NewMockService(GinkgoT())
			service.EXPECT().Create(ctx, &choremodels.CreateChoreParams{
				Deadline:        deadline,
				ChoreTemplateId: &templateId,
			}).Return(nil, fmt.Errorf(
				"create template one-off chore: %w",
				choretemplatemodels.ErrInactive,
			)).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			e := echo.New()
			e.POST("/chores", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/chores",
				strings.NewReader(`{
				"chore_template_id": 42,
				"deadline": "2020-02-03"
			}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"error": "chore template inactive",
			"_links": [{"rel": "collection", "href": "/chores"}],
			"_actions": [{"rel": "create", "href": "/chores"}]
		}`))
		},
	)

	It(
		"rejects direct scheduled creation without calling the service",
		func(ctx SpecContext) {
			service := NewMockService(GinkgoT())
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			e := echo.New()
			e.POST("/chores", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/chores",
				strings.NewReader(`{
				"name": "Kitchen",
				"deadline": "2020-02-03",
				"schedule_name": "Weekly",
				"interval_days": 7
			}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusNotImplemented))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"error": "scheduled recurrence is not implemented",
			"_links": [],
			"_actions": []
		}`))
		},
	)

	It("maps a failed manual HTML submission to the typed form", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		view := NewMockView(GinkgoT())
		view.EXPECT().ManualOneOffForm(choreviewmodels.ManualOneOffForm{
			ActionHref:   "/chores",
			CancelHref:   "/chores",
			SummaryError: "invalid deadline",
			Submitted:    true,
			Name:         choreviewmodels.Field{Value: "Kitchen"},
			Description:  choreviewmodels.Field{Value: "Wash and fold"},
			Deadline: choreviewmodels.Field{
				Value: "invalid",
				Error: "invalid deadline",
			},
		}, false).Return(templ.Raw("manual form sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.POST("/chores", handler.Create)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPost,
			"/chores",
			strings.NewReader("name=Kitchen&description=Wash+and+fold&deadline=invalid"),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.String()).To(Equal("manual form sentinel"))
	})

	It("ignores save-as-template in an HTML submission", func(ctx SpecContext) {
		deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
		service := NewMockService(GinkgoT())
		service.EXPECT().Create(ctx, &choremodels.CreateChoreParams{
			Name:        "Kitchen",
			Description: "Wash and fold",
			Deadline:    deadline,
		}).Return(&choremodels.CreateChoreResult{
			Chore: &choremodels.Chore{
				Id:          42,
				Status:      choremodels.ChoreStatusActive,
				Name:        "Kitchen",
				Description: "Wash and fold",
				Deadline:    deadline,
			},
		}, nil).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
		})
		e := echo.New()
		e.POST("/chores", handler.Create)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPost,
			"/chores",
			strings.NewReader(
				"name=Kitchen&description=Wash+and+fold&deadline=2020-02-03&save_as_chore_template=true",
			),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusSeeOther))
		Expect(response.Header().Get("Location")).To(Equal("/chores"))
	})

	It(
		"ignores save-as-template in an HTML-negotiated JSON submission",
		func(ctx SpecContext) {
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Create(ctx, &choremodels.CreateChoreParams{
				Name:        "Kitchen",
				Description: "Wash and fold",
				Deadline:    deadline,
			}).Return(&choremodels.CreateChoreResult{
				Chore: &choremodels.Chore{
					Id:          42,
					Status:      choremodels.ChoreStatusActive,
					Name:        "Kitchen",
					Description: "Wash and fold",
					Deadline:    deadline,
				},
			}, nil).Once()
			handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
				Service: service,
			})
			e := echo.New()
			e.POST("/chores", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/chores",
				strings.NewReader(`{
				"name": "Kitchen",
				"description": "Wash and fold",
				"deadline": "2020-02-03",
				"save_as_chore_template": true
			}`),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusSeeOther))
			Expect(response.Header().Get("Location")).To(Equal("/chores"))
		},
	)

	It("maps a failed template HTML submission to the typed form", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		view := NewMockView(GinkgoT())
		view.EXPECT().TemplateOneOffForm(choreviewmodels.TemplateOneOffForm{
			ActionHref:      "/chores",
			CancelHref:      "/chores",
			SummaryError:    "invalid deadline",
			Submitted:       true,
			ChoreTemplateId: 42,
			Deadline: choreviewmodels.Field{
				Value: "invalid",
				Error: "invalid deadline",
			},
		}, true).Return(templ.Raw("template form sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.POST("/chores", handler.Create)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPost,
			"/chores",
			strings.NewReader("chore_template_id=42&deadline=invalid"),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.String()).To(Equal("template form sentinel"))
	})
})
