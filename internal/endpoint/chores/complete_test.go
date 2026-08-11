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
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Complete chore", func() {
	It("rejects an ordinary HTML request", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.PUT("/chores/:id/completion", handler.Complete)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPut,
			"/chores/42/completion",
			strings.NewReader(`{"completed_on":"2020-02-03"}`),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Body.String()).To(BeEmpty())
	})

	It("completes a one-off chore through its advertised action", func(ctx SpecContext) {
		completedOn := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
		deadline := time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC)
		service := NewMockService(GinkgoT())
		service.EXPECT().Complete(ctx, &choremodels.CompleteChoreParams{
			Id:          42,
			CompletedOn: completedOn,
		}).Return(&choremodels.CompleteChoreResult{
			Chore: &choremodels.Chore{
				Id:          42,
				Status:      choremodels.ChoreStatusCompleted,
				Name:        "Kitchen",
				Description: "Clean counters",
				Deadline:    deadline,
				CompletedOn: completedOn,
			},
		}, nil).Once()
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.PUT("/chores/:id/completion", handler.Complete)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPut,
			"/chores/42/completion",
			strings.NewReader(`{"completed_on":"2020-02-03"}`),
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
			"status": "completed",
			"name": "Kitchen",
			"description": "Clean counters",
			"deadline": "2020-02-01",
			"completed_on": "2020-02-03",
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

	It("completes a chore from an HTMX form", func(ctx SpecContext) {
		completedOn := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
		service := NewMockService(GinkgoT())
		service.EXPECT().Complete(ctx, &choremodels.CompleteChoreParams{
			Id:          42,
			CompletedOn: completedOn,
		}).Return(&choremodels.CompleteChoreResult{
			Chore: &choremodels.Chore{Id: 42},
		}, nil).Once()
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.PUT("/chores/:id/completion", handler.Complete)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPut,
			"/chores/42/completion",
			strings.NewReader("completed_on=2020-02-03"),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNoContent))
		Expect(response.Body.String()).To(BeEmpty())
		Expect(response.Header().Get("HX-Trigger")).To(Equal("choreCompleted"))
		Expect(response.Header().Get("Location")).To(BeEmpty())
		Expect(response.Header().Get("HX-Redirect")).To(BeEmpty())
		cookies := response.Result().Cookies()
		Expect(cookies).To(HaveLen(1))
		Expect(cookies[0].Name).To(Equal("shiftbell_flash"))
		Expect(cookies[0].Value).To(Equal("chore-completed"))
		Expect(cookies[0].Path).To(Equal("/chores"))
	})

	It(
		"rerenders the dialog with the submitted date when validation fails",
		func(ctx SpecContext) {
			completedOn := time.Date(2030, time.February, 3, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Complete(ctx, &choremodels.CompleteChoreParams{
				Id:          42,
				CompletedOn: completedOn,
			}).Return(nil, errors.Join(
				errors.New("complete chore"),
				validationerrors.ErrInvalidCompletionDate,
			)).Once()
			service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
				Id:     42,
				Status: choremodels.ChoreStatusActive,
				Name:   "Kitchen",
			}, nil).Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().CompletionDialog(choreviewmodels.CompletionDialog{
				Name:       "Kitchen",
				ActionHref: "/chores/42/completion",
				CompletedOn: choreviewmodels.Field{
					Value: "2030-02-03",
					Error: "invalid completion date",
				},
			}).Return(templ.Raw("dialog sentinel")).Once()
			handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
				Service: service,
				View:    view,
			})
			e := echo.New()
			e.PUT("/chores/:id/completion", handler.Complete)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPut,
				"/chores/42/completion",
				strings.NewReader("completed_on=2030-02-03"),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
			Expect(
				response.Header().Get("Content-Type"),
			).To(Equal("text/html; charset=UTF-8"))
			Expect(response.Body.String()).To(Equal("dialog sentinel"))
		},
	)
})
