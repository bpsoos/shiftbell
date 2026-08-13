package chores_test

import (
	"errors"
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
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Edit chore", func() {
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
			Response: choreapimodels.Response{
				Id:          42,
				Status:      choremodels.ChoreStatusActive,
				Name:        "Kitchen edited",
				Description: "Clean all counters.",
				Deadline:    "2026-08-16",
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
			EditHref: "/chores/42/edit",
			Notice:   "Chore updated.",
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
					Response: choreapimodels.Response{
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
					},
					Actions: api.Relations{
						{Rel: "edit", Href: "/chores/42"},
						{Rel: "complete", Href: "/chores/42/completion"},
					},
				},
				BackHref: "/chores",
				EditHref: "/chores/42/edit",
				Notice:   "Chore updated.",
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
