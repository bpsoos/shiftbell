package chores_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/a-h/templ"
	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choreapimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	templatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Edit chore", func() {
	DescribeTable(
		"edits a scheduled chore through JSON while preserving its deadline",
		func(ctx SpecContext, body string, updateTemplate bool) {
			scheduleId := 7
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
				Id:         42,
				ScheduleId: &scheduleId,
				Status:     choremodels.ChoreStatusActive,
				Deadline:   deadline,
			}, nil).Once()
			service.EXPECT().Edit(ctx, &choremodels.EditChoreParams{
				Id:                      42,
				ScheduleId:              &scheduleId,
				Name:                    "  Kitchen edited  ",
				Description:             "  Edited steps.  ",
				Deadline:                deadline,
				AlsoUpdateChoreTemplate: updateTemplate,
			}).Return(&choremodels.ChoreDetails{
				Id: 42, ScheduleId: &scheduleId, Status: choremodels.ChoreStatusActive,
				Name: "Kitchen edited", Description: "Edited steps.", Deadline: deadline,
			}, nil).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			app := echo.New()
			app.PATCH("/chores/:id", handler.Patch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/chores/42",
				strings.NewReader(body),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"id":42,"schedule_id":7,"status":"active","name":"Kitchen edited","description":"Edited steps.",
			"deadline":"2020-02-03","completed_on":null,
			"_links":[{"rel":"self","href":"/chores/42"},{"rel":"collection","href":"/chores"}],
			"_actions":[{"rel":"edit","href":"/chores/42"},{"rel":"complete","href":"/chores/42/completion"}]
		}`))
		},
		Entry(
			"default leaves the template unchanged",
			`{"name":"  Kitchen edited  ","description":"  Edited steps.  "}`,
			false,
		),
		Entry(
			"explicit template update",
			`{"name":"  Kitchen edited  ","description":"  Edited steps.  ","also_update_chore_template":true}`,
			true,
		),
		Entry(
			"changed deadline ignored",
			`{"name":"  Kitchen edited  ","description":"  Edited steps.  ","deadline":"2030-01-01"}`,
			false,
		),
		Entry(
			"invalid deadline ignored",
			`{"name":"  Kitchen edited  ","description":"  Edited steps.  ","deadline":"invalid"}`,
			false,
		),
	)

	It(
		"returns a conflict with an edit retry action when the linked template name is taken",
		func(ctx SpecContext) {
			scheduleId := 7
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
				Id:         42,
				ScheduleId: &scheduleId,
				Status:     choremodels.ChoreStatusActive,
				Deadline:   deadline,
			}, nil).Once()
			service.EXPECT().Edit(ctx, &choremodels.EditChoreParams{
				Id:                      42,
				ScheduleId:              &scheduleId,
				Name:                    "Laundry",
				Description:             "Edited steps.",
				Deadline:                deadline,
				AlsoUpdateChoreTemplate: true,
			}).Return(nil, fmt.Errorf("edit scheduled chore: %w", templatemodels.ErrNameConflict)).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			app := echo.New()
			app.PATCH("/chores/:id", handler.Patch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/chores/42",
				strings.NewReader(`{
			"name":"Laundry","description":"Edited steps.","also_update_chore_template":true
		}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusConflict))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"error":"chore template name conflicts with an active chore template","_links":[],
			"_actions":[{"rel":"edit","href":"/chores/42"}]
		}`))
		},
	)

	DescribeTable(
		"returns scheduled edit errors through JSON",
		func(ctx SpecContext, editErr error, status int, message string) {
			scheduleId := 7
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
				Id:         42,
				ScheduleId: &scheduleId,
				Status:     choremodels.ChoreStatusActive,
				Deadline:   deadline,
			}, nil).Once()
			service.EXPECT().Edit(ctx, &choremodels.EditChoreParams{
				Id:                      42,
				ScheduleId:              &scheduleId,
				Name:                    "Kitchen",
				Deadline:                deadline,
				AlsoUpdateChoreTemplate: true,
			}).Return(nil, fmt.Errorf("edit scheduled chore: %w", editErr)).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			app := echo.New()
			app.PATCH("/chores/:id", handler.Patch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/chores/42",
				strings.NewReader(`{
			"name":"Kitchen","also_update_chore_template":true
		}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(status))
			Expect(response.Body.String()).To(ContainSubstring(message))
		},
		Entry(
			"completed or missing chore",
			choremodels.ErrNotFound,
			http.StatusNotFound,
			"chore not found",
		),
		Entry(
			"inactive template",
			templatemodels.ErrInactive,
			http.StatusUnprocessableEntity,
			"chore template inactive",
		),
		Entry(
			"missing template",
			templatemodels.ErrNotFound,
			http.StatusNotFound,
			"chore template not found",
		),
		Entry(
			"invalid name",
			validationerrors.ErrInvalidName,
			http.StatusUnprocessableEntity,
			"invalid name",
		),
		Entry(
			"invalid description",
			validationerrors.ErrInvalidDescription,
			http.StatusUnprocessableEntity,
			"invalid description",
		),
	)

	It(
		"continues requiring a valid deadline for one-off JSON edits",
		func(ctx SpecContext) {
			service := NewMockService(GinkgoT())
			service.EXPECT().
				Get(ctx, 42).
				Return(&choremodels.ChoreDetails{Id: 42, Status: choremodels.ChoreStatusActive}, nil).
				Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			app := echo.New()
			app.PATCH("/chores/:id", handler.Patch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/chores/42",
				strings.NewReader(`{"name":"Kitchen"}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
			Expect(
				response.Body.Bytes(),
			).To(MatchJSON(`{"error":"invalid deadline","_links":[],"_actions":[]}`))
		},
	)

	It("rejects an HTML representation", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.PATCH("/chores/:id", handler.Patch)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPatch,
			"/chores/42",
			strings.NewReader(`{
				"name": "Kitchen",
				"deadline": "2020-02-03"
			}`),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Body.String()).To(BeEmpty())
	})

	It(
		"edits an active one-off chore through its advertised action",
		func(ctx SpecContext) {
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
				Id:     42,
				Status: choremodels.ChoreStatusActive,
			}, nil).Once()
			service.EXPECT().Edit(ctx, &choremodels.EditChoreParams{
				Id:          42,
				Name:        "  Kitchen edited  ",
				Description: "  Clean all counters.  ",
				Deadline:    deadline,
			}).Return(&choremodels.ChoreDetails{
				Id:          42,
				Status:      choremodels.ChoreStatusActive,
				Name:        "Kitchen edited",
				Description: "Clean all counters.",
				Deadline:    deadline,
			}, nil).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			e := echo.New()
			e.PATCH("/chores/:id", handler.Patch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/chores/42",
				strings.NewReader(`{
				"name": "  Kitchen edited  ",
				"description": "  Clean all counters.  ",
				"deadline": "2020-02-03"
			}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"id": 42,
			"schedule_id": null,
			"status": "active",
			"name": "Kitchen edited",
			"description": "Clean all counters.",
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

	It("edits an active one-off chore from an HTMX form", func(ctx SpecContext) {
		originalDeadline := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
		deadline := time.Date(2026, time.August, 16, 0, 0, 0, 0, time.UTC)
		service := NewMockService(GinkgoT())
		service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
			Id:       42,
			Status:   choremodels.ChoreStatusActive,
			Deadline: originalDeadline,
		}, nil).Once()
		service.EXPECT().Edit(ctx, &choremodels.EditChoreParams{
			Id:          42,
			Name:        "Kitchen edited",
			Description: "Clean all counters.",
			Deadline:    deadline,
		}).Return(&choremodels.ChoreDetails{
			Id:          42,
			Status:      choremodels.ChoreStatusActive,
			Name:        "Kitchen edited",
			Description: "Clean all counters.",
			Deadline:    deadline,
		}, nil).Once()
		chore := choreapimodels.Representation{
			Id:          42,
			Status:      choremodels.ChoreStatusActive,
			Name:        "Kitchen edited",
			Description: "Clean all counters.",
			Deadline:    "2026-08-16",
			Links: api.Relations{
				{Rel: "self", Href: "/chores/42"},
				{Rel: "collection", Href: "/chores"},
			},
			Actions: api.Relations{
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
			Notice:     "Chore updated.",
		}, false).Return(templ.Raw("updated detail sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.PATCH("/chores/:id", handler.Patch)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPatch,
			"/chores/42",
			strings.NewReader(
				"name=Kitchen+edited&description=Clean+all+counters.&deadline=2026-08-16",
			),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Header().Get("HX-Replace-Url")).To(Equal("/chores/42"))
		Expect(response.Body.String()).To(Equal("updated detail sentinel"))
	})

	It(
		"edits a scheduled chore without accepting a changed deadline",
		func(ctx SpecContext) {
			scheduleId := 7
			deadline := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
				Id:         42,
				ScheduleId: &scheduleId,
				Status:     choremodels.ChoreStatusActive,
				Deadline:   deadline,
			}, nil).Once()
			service.EXPECT().Edit(ctx, &choremodels.EditChoreParams{
				Id:                      42,
				ScheduleId:              &scheduleId,
				Name:                    "Kitchen edited",
				Description:             "Clean all counters.",
				Deadline:                deadline,
				AlsoUpdateChoreTemplate: true,
			}).Return(&choremodels.ChoreDetails{
				Id:          42,
				ScheduleId:  &scheduleId,
				Status:      choremodels.ChoreStatusActive,
				Name:        "Kitchen edited",
				Description: "Clean all counters.",
				Deadline:    deadline,
			}, nil).Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().Detail(choreviewmodels.Detail{
				Chore: choreapimodels.Representation{
					Id:          42,
					ScheduleId:  &scheduleId,
					Status:      choremodels.ChoreStatusActive,
					Name:        "Kitchen edited",
					Description: "Clean all counters.",
					Deadline:    "2026-08-15",
					Links: api.Relations{
						{Rel: "self", Href: "/chores/42"},
						{Rel: "collection", Href: "/chores"},
					},
					Actions: api.Relations{
						{Rel: "edit", Href: "/chores/42"},
						{Rel: "complete", Href: "/chores/42/completion"},
					},
				},
				BackHref:   "/chores",
				EditHref:   "/chores/42/edit",
				DeleteHref: "",
				Notice:     "Chore updated.",
			}, false).Return(templ.Raw("updated scheduled detail sentinel")).Once()
			handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
				Service: service,
				View:    view,
			})
			e := echo.New()
			e.PATCH("/chores/:id", handler.Patch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/chores/42",
				strings.NewReader(
					"name=Kitchen+edited&description=Clean+all+counters.&deadline=2030-01-01&also_update_chore_template=true",
				),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Header().Get("HX-Replace-Url")).To(Equal("/chores/42"))
			Expect(response.Body.String()).To(Equal("updated scheduled detail sentinel"))
		},
	)

	It(
		"preserves the template-update choice when an HTMX edit has a template name conflict",
		func(ctx SpecContext) {
			scheduleId := 7
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
				Id:         42,
				ScheduleId: &scheduleId,
				Status:     choremodels.ChoreStatusActive,
				Deadline:   deadline,
			}, nil).Once()
			service.EXPECT().Edit(ctx, &choremodels.EditChoreParams{
				Id:                      42,
				ScheduleId:              &scheduleId,
				Name:                    "Laundry",
				Description:             "Edited steps.",
				Deadline:                deadline,
				AlsoUpdateChoreTemplate: true,
			}).Return(nil, fmt.Errorf("edit scheduled chore: %w", templatemodels.ErrNameConflict)).Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().EditForm(choreviewmodels.EditForm{
				ActionHref: "/chores/42",
				CancelHref: "/chores/42",
				Scheduled:  true,
				Name: choreviewmodels.Field{
					Value: "Laundry",
					Error: templatemodels.ErrNameConflict.Error(),
				},
				Description:             choreviewmodels.Field{Value: "Edited steps."},
				Deadline:                choreviewmodels.Field{Value: "2020-02-03"},
				AlsoUpdateChoreTemplate: true,
				SummaryError:            templatemodels.ErrNameConflict.Error(),
			}, false).Return(templ.Raw("conflict form")).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service, View: view},
			)
			app := echo.New()
			app.PATCH("/chores/:id", handler.Patch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/chores/42",
				strings.NewReader(
					"name=Laundry&description=Edited+steps.&also_update_chore_template=true",
				),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusConflict))
			Expect(response.Body.String()).To(Equal("conflict form"))
		},
	)

	It("preserves submitted values and renders field feedback", func(ctx SpecContext) {
		deadline := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
		service := NewMockService(GinkgoT())
		service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
			Id:       42,
			Status:   choremodels.ChoreStatusActive,
			Deadline: deadline,
		}, nil).Once()
		service.EXPECT().Edit(ctx, &choremodels.EditChoreParams{
			Id:          42,
			Name:        "",
			Description: "Submitted description",
			Deadline:    deadline,
		}).Return(nil, errors.Join(
			errors.New("edit chore"),
			validationerrors.ErrInvalidName,
		)).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().EditForm(choreviewmodels.EditForm{
			ActionHref: "/chores/42",
			CancelHref: "/chores/42",
			Name: choreviewmodels.Field{
				Error: "invalid name",
			},
			Description: choreviewmodels.Field{
				Value: "Submitted description",
			},
			Deadline: choreviewmodels.Field{
				Value: "2026-08-15",
			},
			SummaryError: "invalid name",
		}, false).Return(templ.Raw("edit validation sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.PATCH("/chores/:id", handler.Patch)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPatch,
			"/chores/42",
			strings.NewReader(
				"name=&description=Submitted+description&deadline=2026-08-15",
			),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.String()).To(Equal("edit validation sentinel"))
	})
})
