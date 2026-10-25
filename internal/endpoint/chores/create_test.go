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
	schedulemodels "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create chore", func() {
	It(
		"returns a conflict and a retry action when an active schedule name is taken",
		func(ctx SpecContext) {
			templateID, interval := 7, 14
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Create(ctx, &choremodels.CreateChoreParams{
				ChoreTemplateId: &templateID,
				ScheduleName:    "Fortnightly",
				IntervalDays:    &interval,
				Deadline:        deadline,
			}).Return(nil, fmt.Errorf("create template scheduled chore: %w", schedulemodels.ErrNameConflict)).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			app := echo.New()
			app.POST("/chores", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/chores",
				strings.NewReader(`{
			"chore_template_id":7,"schedule_name":"Fortnightly","interval_days":14,"deadline":"2020-02-03"
		}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusConflict))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Header().Get("Location")).To(BeEmpty())
			Expect(response.Header().Get("Set-Cookie")).To(BeEmpty())
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"error":"schedule name conflicts with an active schedule",
			"_links":[],
			"_actions":[{"rel":"create","href":"/chores"}]
		}`))
		},
	)

	It(
		"creates a template-based scheduled chore through the JSON API",
		func(ctx SpecContext) {
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			templateID, interval, scheduleID := 7, 14, 3
			service := NewMockService(GinkgoT())
			service.EXPECT().Create(ctx, &choremodels.CreateChoreParams{
				ChoreTemplateId: &templateID,
				ScheduleName:    "  Fortnightly  ",
				IntervalDays:    &interval,
				Deadline:        deadline,
			}).Return(&choremodels.CreateChoreResult{Chore: &choremodels.Chore{
				Id:          42,
				ScheduleId:  &scheduleID,
				Name:        "Kitchen",
				Description: "Reusable steps.",
				Status:      choremodels.ChoreStatusActive,
				Deadline:    deadline,
			}}, nil).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			app := echo.New()
			app.POST("/chores", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/chores",
				strings.NewReader(`{
			"chore_template_id":7,"schedule_name":"  Fortnightly  ","interval_days":14,"deadline":"2020-02-03"
		}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusCreated))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Header().Get("Location")).To(Equal("/chores/42"))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"id":42,"schedule_id":3,"status":"active","name":"Kitchen","description":"Reusable steps.",
			"deadline":"2020-02-03","completed_on":null,
			"_links":[{"rel":"self","href":"/chores/42"},{"rel":"collection","href":"/chores"}],
			"_actions":[{"rel":"edit","href":"/chores/42"},{"rel":"complete","href":"/chores/42/completion"}]
		}`))
		},
	)

	DescribeTable(
		"rejects invalid template-based scheduled creation",
		func(ctx SpecContext, createErr error, status int) {
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			templateID, interval := 7, 14
			service := NewMockService(GinkgoT())
			service.EXPECT().Create(ctx, &choremodels.CreateChoreParams{
				ChoreTemplateId: &templateID,
				ScheduleName:    "Fortnightly",
				IntervalDays:    &interval,
				Deadline:        deadline,
			}).Return(nil, fmt.Errorf("create template scheduled chore: %w", createErr)).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			app := echo.New()
			app.POST("/chores", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/chores",
				strings.NewReader(`{
			"chore_template_id":7,"schedule_name":"Fortnightly","interval_days":14,"deadline":"2020-02-03"
		}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
			Expect(response.Header().Get("Location")).To(BeEmpty())
			Expect(response.Body.String()).To(ContainSubstring(createErr.Error()))
		},
		Entry(
			"inactive template",
			choretemplatemodels.ErrInactive,
			http.StatusUnprocessableEntity,
		),
		Entry("missing template", choretemplatemodels.ErrNotFound, http.StatusNotFound),
		Entry(
			"invalid schedule name",
			validationerrors.ErrInvalidName,
			http.StatusUnprocessableEntity,
		),
		Entry(
			"invalid interval",
			validationerrors.ErrInvalidInterval,
			http.StatusUnprocessableEntity,
		),
	)

	DescribeTable(
		"keeps scheduled HTML submissions unsupported",
		func(ctx SpecContext, source string) {
			service := NewMockService(GinkgoT())
			view := NewMockView(GinkgoT())
			view.EXPECT().Error(choreviewmodels.Error{
				Message: "scheduled recurrence is not implemented",
			}, true).Return(templ.Raw("not implemented")).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service, View: view},
			)
			app := echo.New()
			app.POST("/chores", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/chores",
				strings.NewReader(
					source+"&schedule_name=Weekly&interval_days=7&deadline=2020-02-03",
				),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("Content-Type", echo.MIMEApplicationForm)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusNotImplemented))
		},
		Entry("from a template", "chore_template_id=7"),
		Entry("manual", "name=Kitchen"),
	)

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
		"creates a manual scheduled chore and returns only the chore",
		func(ctx SpecContext) {
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			interval, scheduleID := 7, 3
			service := NewMockService(GinkgoT())
			service.EXPECT().Create(ctx, &choremodels.CreateChoreParams{
				Name: "  Kitchen  ", Description: "  Reusable steps.  ",
				Deadline: deadline, ScheduleName: "  Weekly  ", IntervalDays: &interval,
			}).Return(&choremodels.CreateChoreResult{
				Chore: &choremodels.Chore{
					Id:          42,
					ScheduleId:  &scheduleID,
					Status:      choremodels.ChoreStatusActive,
					Name:        "Kitchen",
					Description: "Reusable steps.",
					Deadline:    deadline,
				},
				ChoreTemplate: &choretemplatemodels.ChoreTemplate{Id: 7, Name: "Kitchen"},
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
				"description": "  Reusable steps.  ",
				"deadline": "2020-02-03",
				"schedule_name": "  Weekly  ",
				"interval_days": 7
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
			"id":42,"schedule_id":3,"status":"active","name":"Kitchen","description":"Reusable steps.",
			"deadline":"2020-02-03","completed_on":null,
			"_links":[{"rel":"self","href":"/chores/42"},{"rel":"collection","href":"/chores"}],
			"_actions":[{"rel":"edit","href":"/chores/42"},{"rel":"complete","href":"/chores/42/completion"}]
		}`))
		},
	)

	DescribeTable(
		"returns manual scheduled creation feedback",
		func(ctx SpecContext, createErr error, status int) {
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			interval := 7
			service := NewMockService(GinkgoT())
			service.EXPECT().Create(ctx, &choremodels.CreateChoreParams{
				Name:         "Kitchen",
				Deadline:     deadline,
				ScheduleName: "Weekly",
				IntervalDays: &interval,
			}).Return(nil, fmt.Errorf("create manual scheduled chore: %w", createErr)).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			app := echo.New()
			app.POST("/chores", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/chores",
				strings.NewReader(`{
			"name":"Kitchen","deadline":"2020-02-03","schedule_name":"Weekly","interval_days":7
		}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(status))
			Expect(response.Header().Get("Location")).To(BeEmpty())
			Expect(response.Body.String()).To(ContainSubstring(createErr.Error()))
		},
		Entry(
			"template name conflict",
			choretemplatemodels.ErrNameConflict,
			http.StatusConflict,
		),
		Entry(
			"schedule name conflict",
			schedulemodels.ErrNameConflict,
			http.StatusConflict,
		),
		Entry(
			"invalid name",
			validationerrors.ErrInvalidName,
			http.StatusUnprocessableEntity,
		),
		Entry(
			"invalid description",
			validationerrors.ErrInvalidDescription,
			http.StatusUnprocessableEntity,
		),
		Entry(
			"invalid interval",
			validationerrors.ErrInvalidInterval,
			http.StatusUnprocessableEntity,
		),
	)

	It(
		"preserves save-as-template in a failed manual HTML submission",
		func(ctx SpecContext) {
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
				SaveAsTemplate: true,
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
				strings.NewReader(
					"name=Kitchen&description=Wash+and+fold&deadline=invalid&save_as_chore_template=true",
				),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
			Expect(response.Body.String()).To(Equal("manual form sentinel"))
		},
	)

	It("saves a template from a manual HTML submission", func(ctx SpecContext) {
		deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
		service := NewMockService(GinkgoT())
		service.EXPECT().Create(ctx, &choremodels.CreateChoreParams{
			Name:                "Kitchen",
			Description:         "Wash and fold",
			Deadline:            deadline,
			SaveAsChoreTemplate: true,
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
		cookies := response.Result().Cookies()
		Expect(cookies).To(HaveLen(1))
		Expect(cookies[0].Name).To(Equal("shiftbell_flash"))
		Expect(cookies[0].Value).To(Equal("chore-and-template-created"))
		Expect(cookies[0].Path).To(Equal("/chores"))
		Expect(cookies[0].HttpOnly).To(BeTrue())
		Expect(cookies[0].SameSite).To(Equal(http.SameSiteLaxMode))
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
