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
	apimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Correct chore completion", func() {
	DescribeTable(
		"renders the corrected chore as HTML",
		func(ctx SpecContext, hxRequest string, fullPage bool) {
			completedOn := time.Date(2020, 2, 4, 0, 0, 0, 0, time.UTC)
			deadline := time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().
				CorrectCompletion(ctx, &choremodels.CorrectCompletionParams{Id: 42, CompletedOn: completedOn}).
				Return(&choremodels.Chore{
					Id:          42,
					Name:        "Kitchen",
					Status:      choremodels.ChoreStatusCompleted,
					Deadline:    deadline,
					CompletedOn: completedOn,
				}, nil).
				Once()
			view := NewMockView(GinkgoT())
			completedDate := "2020-02-04"
			view.EXPECT().Detail(viewmodels.Detail{
				Chore: apimodels.Representation{Response: apimodels.Response{
					Id:          42,
					Name:        "Kitchen",
					Status:      choremodels.ChoreStatusCompleted,
					Deadline:    "2020-02-01",
					CompletedOn: &completedDate,
					Links: api.Relations{
						{Rel: "self", Href: "/chores/42"},
						{Rel: "collection", Href: "/chores"},
					},
				}, Actions: api.Relations{{Rel: "correct_completion", Href: "/chores/42/completion"}, {Rel: "delete", Href: "/chores/42"}}},
				BackHref:   "/chores",
				DeleteHref: "/chores/42/deletion",
				Notice:     "Completion date corrected.",
			}, fullPage).Return(templ.Raw("corrected chore")).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service, View: view},
			)
			app := echo.New()
			app.PATCH("/chores/:id/completion", handler.CorrectCompletion)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/chores/42/completion",
				strings.NewReader(`{"completed_on":"2020-02-04"}`),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", hxRequest)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(
				response.Header().Get("Content-Type"),
			).To(Equal(echo.MIMETextHTMLCharsetUTF8))
			Expect(response.Body.String()).To(Equal("corrected chore"))
		},
		Entry("full page", "", true),
		Entry("HTMX fragment", "true", false),
	)

	DescribeTable(
		"renders completion correction failures as HTML",
		func(ctx SpecContext, correctionErr error, status int, message string) {
			completedOn := time.Date(2020, 2, 4, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().
				CorrectCompletion(ctx, &choremodels.CorrectCompletionParams{Id: 42, CompletedOn: completedOn}).
				Return(nil, correctionErr).
				Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().
				Error(viewmodels.Error{Message: message}, true).
				Return(templ.Raw("correction error")).
				Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service, View: view},
			)
			app := echo.New()
			app.PATCH("/chores/:id/completion", handler.CorrectCompletion)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/chores/42/completion",
				strings.NewReader(`{"completed_on":"2020-02-04"}`),
			)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(status))
			Expect(response.Body.String()).To(Equal("correction error"))
		},
		Entry(
			"invalid date",
			validationerrors.ErrInvalidCompletionDate,
			http.StatusUnprocessableEntity,
			"invalid completion date",
		),
		Entry(
			"missing chore",
			choremodels.ErrNotFound,
			http.StatusNotFound,
			"chore not found",
		),
		Entry(
			"storage failure",
			errors.New("storage failed"),
			http.StatusInternalServerError,
			"something went wrong",
		),
	)
	It(
		"corrects a one-off completion through its advertised action",
		func(ctx SpecContext) {
			completedOn := time.Date(2020, time.February, 4, 0, 0, 0, 0, time.UTC)
			deadline := time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().CorrectCompletion(ctx, &choremodels.CorrectCompletionParams{
				Id:          42,
				CompletedOn: completedOn,
			}).Return(&choremodels.ChoreDetails{
				Id:          42,
				Status:      choremodels.ChoreStatusCompleted,
				Name:        "Kitchen",
				Description: "Clean counters",
				Deadline:    deadline,
				CompletedOn: completedOn,
			}, nil).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			e := echo.New()
			e.PATCH("/chores/:id/completion", handler.CorrectCompletion)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/chores/42/completion",
				strings.NewReader(`{"completed_on":"2020-02-04"}`),
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
		},
	)
})
