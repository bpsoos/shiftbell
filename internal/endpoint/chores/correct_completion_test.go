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
		func(ctx SpecContext, hxRequest string, fullPage bool, contentType, body string) {
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
				strings.NewReader(body),
			)
			request.Header.Set("Content-Type", contentType)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", hxRequest)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(
				response.Header().Get("Content-Type"),
			).To(Equal(echo.MIMETextHTMLCharsetUTF8))
			Expect(response.Body.String()).To(Equal("corrected chore"))
			Expect(response.Header().Get("HX-Replace-Url")).To(Equal("/chores/42"))
		},
		Entry("full page JSON", "", true, "", `{"completed_on":"2020-02-04"}`),
		Entry(
			"HTMX JSON",
			"true",
			false,
			"application/json",
			`{"completed_on":"2020-02-04"}`,
		),
		Entry(
			"HTMX form",
			"true",
			false,
			echo.MIMEApplicationForm,
			"completed_on=2020-02-04",
		),
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
			"missing chore",
			choremodels.ErrNotFound,
			http.StatusNotFound,
			"chore not found",
		),
	)

	DescribeTable(
		"preserves the entered HTML date when correction fails",
		func(ctx SpecContext, serviceErr error, status int, fieldError, summaryError string) {
			service := NewMockService(GinkgoT())
			service.EXPECT().
				CorrectCompletion(ctx, &choremodels.CorrectCompletionParams{Id: 42, CompletedOn: time.Date(2020, 2, 4, 0, 0, 0, 0, time.UTC)}).
				Return(nil, serviceErr).
				Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().CompletionCorrectionForm(viewmodels.CompletionCorrectionForm{
				ActionHref: "/chores/42/completion",
				CancelHref: "/chores/42",
				CompletedOn: viewmodels.Field{
					Value: "2020-02-04",
					Error: fieldError,
				},
				SummaryError: summaryError,
			}, false).Return(templ.Raw("feedback")).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service, View: view},
			)
			app := echo.New()
			app.PATCH("/chores/:id/completion", handler.CorrectCompletion)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/chores/42/completion",
				strings.NewReader("completed_on=2020-02-04"),
			)
			request.Header.Set("Content-Type", echo.MIMEApplicationForm)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
			Expect(response.Header().Get("HX-Replace-Url")).To(BeEmpty())
			Expect(response.Body.String()).To(Equal("feedback"))
		},
		Entry(
			"invalid date",
			validationerrors.ErrInvalidCompletionDate,
			422,
			"Enter a valid date no later than today.",
			"",
		),
		Entry(
			"storage error",
			errors.New("storage failed"),
			500,
			"",
			"The completion date could not be updated. Try again.",
		),
	)

	DescribeTable(
		"rejects malformed HTML dates without calling the service",
		func(ctx SpecContext, date string) {
			view := NewMockView(GinkgoT())
			view.EXPECT().CompletionCorrectionForm(viewmodels.CompletionCorrectionForm{
				ActionHref: "/chores/42/completion",
				CancelHref: "/chores/42",
				CompletedOn: viewmodels.Field{
					Value: date,
					Error: "Enter a valid date no later than today.",
				},
			}, true).Return(templ.Raw("date error")).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{
					Service: NewMockService(GinkgoT()),
					View:    view,
				},
			)
			app := echo.New()
			app.PATCH("/chores/:id/completion", handler.CorrectCompletion)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/chores/42/completion",
				strings.NewReader("completed_on="+date),
			)
			request.Header.Set("Content-Type", echo.MIMEApplicationForm)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		},
		Entry("empty", ""),
		Entry("invalid calendar date", "2020-02-30"),
		Entry("text", "yesterday"),
	)

	DescribeTable(
		"rejects unsupported representations and invalid IDs without correction",
		func(ctx SpecContext, path, accept string, status int) {
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: NewMockService(GinkgoT())},
			)
			app := echo.New()
			app.PATCH("/chores/:id/completion", handler.CorrectCompletion)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				path,
				strings.NewReader(`{"completed_on":"2020-02-04"}`),
			)
			request.Header.Set("Accept", accept)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
		},
		Entry("unsupported", "/chores/42/completion", "image/png", 406),
		Entry("invalid ID", "/chores/invalid/completion", hypermedia.MediaType, 422),
	)
	DescribeTable(
		"corrects a one-off completion as JSON independently of input encoding",
		func(ctx SpecContext, contentType, body string) {
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
				strings.NewReader(body),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", contentType)
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
		Entry("JSON", "application/json", `{"completed_on":"2020-02-04"}`),
		Entry("form", echo.MIMEApplicationForm, "completed_on=2020-02-04"),
	)
})
