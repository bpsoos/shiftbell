package chores_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
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
			"_links": {},
			"_actions": {
				"create": {
					"href": "/chores",
					"method": "POST",
					"content_type": "application/json",
					"fields": null
				}
			}
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
			"_links": {
				"collection": {"href": "/chores"}
			},
			"_actions": {
				"create": {
					"href": "/chores",
					"method": "POST",
					"content_type": "application/json",
					"fields": null
				}
			}
		}`))
		},
	)
})
