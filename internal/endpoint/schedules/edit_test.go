package schedules_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/schedules"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Edit", func() {
	var (
		service *MockService
		app     *echo.Echo
	)
	BeforeEach(func() {
		service = NewMockService(GinkgoT())
		handler := schedules.NewHandler(&schedules.HandlerDeps{Service: service})
		app = echo.New()
		app.PATCH("/schedules/:id", handler.Edit)
	})

	It(
		"edits the name and interval without accepting a replacement template",
		func(ctx SpecContext) {
			choreID := 11
			deadline := time.Date(2020, 4, 2, 0, 0, 0, 0, time.UTC)
			service.EXPECT().
				Edit(ctx, &models.EditScheduleParams{Id: 42, Name: " Biweekly laundry ", IntervalDays: 14}).
				Return(&models.Schedule{
					Id:                  42,
					Name:                "Biweekly laundry",
					ChoreTemplateId:     7,
					ChoreTemplateName:   "Wash clothes",
					IntervalDays:        14,
					ActiveChoreId:       &choreID,
					ActiveChoreDeadline: &deadline,
				}, nil).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/schedules/42",
				strings.NewReader(
					`{"name":" Biweekly laundry ","interval_days":14,"chore_template_id":99}`,
				),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Header().Get("Vary")).To(Equal("Accept"))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"id":42,"name":"Biweekly laundry","chore_template_id":7,"chore_template_name":"Wash clothes",
			"interval_days":14,"deactivated_at":null,"active_chore_id":11,"active_chore_deadline":"2020-04-02T00:00:00Z",
			"_links":[{"rel":"self","href":"/schedules/42"},{"rel":"collection","href":"/schedules"},
				{"rel":"chore_template","href":"/chore-templates/7"},{"rel":"active_chore","href":"/chores/11"}],
			"_actions":[{"rel":"edit","href":"/schedules/42"}]
		}`))
		},
	)

	DescribeTable(
		"maps retryable edit errors",
		func(ctx SpecContext, serviceErr error, status int, message string) {
			service.EXPECT().
				Edit(ctx, &models.EditScheduleParams{Id: 42, Name: "Laundry", IntervalDays: 14}).
				Return(nil, fmt.Errorf("edit schedule: %w", serviceErr)).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/schedules/42",
				strings.NewReader(`{"name":"Laundry","interval_days":14}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(status))
			Expect(response.Body.Bytes()).To(MatchJSON(fmt.Sprintf(`{
			"error":%q,"_links":[{"rel":"collection","href":"/schedules"}],
			"_actions":[{"rel":"edit","href":"/schedules/42"}]
		}`, message)))
		},
		Entry(
			"name conflict",
			models.ErrNameConflict,
			http.StatusConflict,
			models.ErrNameConflict.Error(),
		),
		Entry(
			"invalid name",
			validationerrors.ErrInvalidName,
			http.StatusUnprocessableEntity,
			validationerrors.ErrInvalidName.Error(),
		),
		Entry(
			"invalid interval",
			validationerrors.ErrInvalidInterval,
			http.StatusUnprocessableEntity,
			validationerrors.ErrInvalidInterval.Error(),
		),
	)

	DescribeTable(
		"maps errors without advertising an edit retry",
		func(ctx SpecContext, serviceErr error, status int, message string) {
			service.EXPECT().
				Edit(ctx, &models.EditScheduleParams{Id: 42, Name: "Laundry", IntervalDays: 14}).
				Return(nil, fmt.Errorf("edit schedule: %w", serviceErr)).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/schedules/42",
				strings.NewReader(`{"name":"Laundry","interval_days":14}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(status))
			Expect(response.Body.Bytes()).To(MatchJSON(fmt.Sprintf(`{
			"error":%q,"_links":[{"rel":"collection","href":"/schedules"}]
		}`, message)))
		},
		Entry(
			"inactive",
			models.ErrInactive,
			http.StatusUnprocessableEntity,
			models.ErrInactive.Error(),
		),
		Entry(
			"not found",
			models.ErrNotFound,
			http.StatusNotFound,
			models.ErrNotFound.Error(),
		),
		Entry(
			"internal failure",
			errors.New("database unavailable"),
			http.StatusInternalServerError,
			"something went wrong",
		),
	)

	DescribeTable("rejects invalid request bodies", func(ctx SpecContext, body string) {
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPatch,
			"/schedules/42",
			strings.NewReader(body),
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(
			response.Body.Bytes(),
		).To(MatchJSON(`{"error":"invalid request body","_links":[{"rel":"collection","href":"/schedules"}],"_actions":[{"rel":"edit","href":"/schedules/42"}]}`))
	},
		Entry("malformed JSON", `{`),
		Entry("empty", ""),
		Entry("fractional interval", `{"name":"Laundry","interval_days":1.5}`),
		Entry("string interval", `{"name":"Laundry","interval_days":"14"}`),
	)

	DescribeTable("rejects invalid IDs", func(ctx SpecContext, id string) {
		request := httptest.NewRequestWithContext(ctx, http.MethodPatch, "/schedules/"+id,
			strings.NewReader(`{"name":"Laundry","interval_days":14}`))
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(
			response.Body.Bytes(),
		).To(MatchJSON(`{"error":"invalid schedule id","_links":[{"rel":"collection","href":"/schedules"}]}`))
	}, Entry("zero", "0"), Entry("negative", "-1"), Entry("non-numeric", "invalid"))

	It("rejects unsupported representations", func(ctx SpecContext) {
		request := httptest.NewRequestWithContext(ctx, http.MethodPatch, "/schedules/42",
			strings.NewReader(`{"name":"Laundry","interval_days":14}`))
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
	})
})
