package schedules_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/schedules"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	apimodels "github.com/bpsoos/shiftbell/internal/models/api/schedules"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/schedules"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Edit", func() {
	var (
		service *MockService
		view    *MockView
		app     *echo.Echo
	)
	BeforeEach(func() {
		service = NewMockService(GinkgoT())
		view = NewMockView(GinkgoT())
		handler := schedules.NewHandler(
			&schedules.HandlerDeps{Service: service, View: view},
		)
		app = echo.New()
		app.PATCH("/schedules/:id", handler.Edit)
	})

	It("does not edit for an unsupported representation", func(ctx SpecContext) {
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPatch,
			"/schedules/42",
			strings.NewReader(`{"name":"Kitchen","interval_days":7}`),
		)
		request.Header.Set("Accept", "image/png")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
	})

	DescribeTable(
		"renders edited schedule details as HTML",
		func(ctx SpecContext, hxRequest string, fullPage bool) {
			service.EXPECT().
				Edit(ctx, &models.EditScheduleParams{Id: 42, Name: "Fortnightly kitchen", IntervalDays: 14}).
				Return(&models.Schedule{Id: 42, Name: "Fortnightly kitchen", ChoreTemplateId: 7, ChoreTemplateName: "Kitchen", IntervalDays: 14}, nil).
				Once()
			view.EXPECT().Detail(viewmodels.Detail{Schedule: apimodels.Representation{
				Response: apimodels.Response{
					Id:                42,
					Name:              "Fortnightly kitchen",
					ChoreTemplateId:   7,
					ChoreTemplateName: "Kitchen",
					IntervalDays:      14,
					Links: api.Relations{
						{Rel: "self", Href: "/schedules/42"},
						{Rel: "collection", Href: "/schedules"},
						{Rel: "chore_template", Href: "/chore-templates/7"},
					},
				},
				Actions: api.Relations{
					{Rel: "edit", Href: "/schedules/42"},
					{Rel: "deactivate", Href: "/schedules/42/deactivation"},
				},
			}, Notice: "Schedule updated."}, fullPage).Return(templ.Raw("edited schedule")).Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/schedules/42",
				strings.NewReader(`{"name":"Fortnightly kitchen","interval_days":14}`),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", hxRequest)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(
				response.Header().Get("Content-Type"),
			).To(Equal(echo.MIMETextHTMLCharsetUTF8))
			Expect(response.Body.String()).To(Equal("edited schedule"))
		},
		Entry("full page", "", true),
		Entry("HTMX fragment", "true", false),
	)

	It(
		"renders a name conflict as an HTML form even for JSON input",
		func(ctx SpecContext) {
			service.EXPECT().
				Edit(ctx, &models.EditScheduleParams{Id: 42, Name: "Taken", IntervalDays: 7}).
				Return(nil, models.ErrNameConflict).
				Once()
			view.EXPECT().EditForm(viewmodels.EditForm{
				ActionHref:   "/schedules/42",
				CancelHref:   "/schedules/42",
				Name:         "Taken",
				IntervalDays: "7",
				NameError:    "An active schedule already uses this name.",
			}, false).Return(templ.Raw("edit conflict")).Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/schedules/42",
				strings.NewReader(`{"name":"Taken","interval_days":7}`),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusConflict))
			Expect(response.Body.String()).To(Equal("edit conflict"))
		},
	)

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
			"_actions":[{"rel":"edit","href":"/schedules/42"},{"rel":"deactivate","href":"/schedules/42/deactivation"}]
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

	Context("request decoding", func() {
		DescribeTable(
			"decodes input independently of JSON response negotiation",
			func(ctx SpecContext, contentType, body string) {
				service.EXPECT().
					Edit(ctx, &models.EditScheduleParams{Id: 42, Name: "Kitchen", IntervalDays: 14}).
					Return(&models.Schedule{Id: 42, Name: "Kitchen", IntervalDays: 14}, nil).
					Once()
				request := httptest.NewRequestWithContext(
					ctx,
					http.MethodPatch,
					"/schedules/42",
					strings.NewReader(body),
				)
				request.Header.Set("Content-Type", contentType)
				request.Header.Set("Accept", hypermedia.MediaType)
				response := httptest.NewRecorder()
				app.ServeHTTP(response, request)
				Expect(response.Code).To(Equal(http.StatusOK))
				Expect(
					response.Header().Get("Content-Type"),
				).To(Equal(hypermedia.MediaType))
				Expect(response.Header().Get("HX-Replace-Url")).To(BeEmpty())
				Expect(response.Body.String()).To(ContainSubstring(`"name":"Kitchen"`))
			},
			Entry(
				"JSON",
				echo.MIMEApplicationJSON,
				`{"name":"Kitchen","interval_days":14}`,
			),
			Entry(
				"vendor JSON",
				hypermedia.MediaType,
				`{"name":"Kitchen","interval_days":14}`,
			),
			Entry(
				"legacy missing content type",
				"",
				`{"name":"Kitchen","interval_days":14}`,
			),
			Entry("form", echo.MIMEApplicationForm, "name=Kitchen&interval_days=14"),
			Entry(
				"form with charset",
				echo.MIMEApplicationForm+"; charset=UTF-8",
				"name=Kitchen&interval_days=14",
			),
		)

		DescribeTable(
			"renders HTML feedback independently of input encoding",
			func(ctx SpecContext, contentType, body string) {
				service.EXPECT().
					Edit(ctx, &models.EditScheduleParams{Id: 42, Name: "Kitchen", IntervalDays: 14}).
					Return(nil, models.ErrNameConflict).
					Once()
				view.EXPECT().EditForm(viewmodels.EditForm{
					ActionHref:   "/schedules/42",
					CancelHref:   "/schedules/42",
					Name:         "Kitchen",
					IntervalDays: "14",
					NameError:    "An active schedule already uses this name.",
				}, false).Return(templ.Raw("inline feedback")).Once()
				request := httptest.NewRequestWithContext(
					ctx,
					http.MethodPatch,
					"/schedules/42",
					strings.NewReader(body),
				)
				request.Header.Set("Content-Type", contentType)
				request.Header.Set("Accept", "text/html")
				request.Header.Set("HX-Request", "true")
				response := httptest.NewRecorder()
				app.ServeHTTP(response, request)
				Expect(response.Code).To(Equal(http.StatusConflict))
				Expect(response.Body.String()).To(Equal("inline feedback"))
			},
			Entry(
				"JSON",
				echo.MIMEApplicationJSON,
				`{"name":"Kitchen","interval_days":14}`,
			),
			Entry("form", echo.MIMEApplicationForm, "name=Kitchen&interval_days=14"),
		)

		It(
			"renders invalid form intervals as JSON when requested",
			func(ctx SpecContext) {
				request := httptest.NewRequestWithContext(
					ctx,
					http.MethodPatch,
					"/schedules/42",
					strings.NewReader("name=Kitchen&interval_days=weekly"),
				)
				request.Header.Set("Content-Type", echo.MIMEApplicationForm)
				request.Header.Set("Accept", hypermedia.MediaType)
				response := httptest.NewRecorder()
				app.ServeHTTP(response, request)
				Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
				Expect(
					response.Body.String(),
				).To(ContainSubstring(`"error":"invalid interval"`))
			},
		)

		DescribeTable(
			"rejects unsupported input encodings without editing",
			func(ctx SpecContext, contentType string) {
				request := httptest.NewRequestWithContext(
					ctx,
					http.MethodPatch,
					"/schedules/42",
					strings.NewReader(`{"name":"Kitchen","interval_days":14}`),
				)
				request.Header.Set("Content-Type", contentType)
				request.Header.Set("Accept", hypermedia.MediaType)
				response := httptest.NewRecorder()
				app.ServeHTTP(response, request)
				Expect(response.Code).To(Equal(http.StatusUnsupportedMediaType))
				Expect(
					response.Body.String(),
				).To(ContainSubstring(`"error":"unsupported media type"`))
			},
			Entry("text", "text/plain"),
			Entry("malformed", "application/json; charset"),
			Entry("form-like", echo.MIMEApplicationForm+"-invalid"),
		)
	})

	Context("form submissions", func() {
		var request *http.Request
		BeforeEach(func(ctx SpecContext) {
			request = httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/schedules/42",
				strings.NewReader(url.Values{
					"name": {
						" Kitchen ",
					},
					"interval_days":     {"14"},
					"chore_template_id": {"99"},
				}.Encode()),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("Content-Type", echo.MIMEApplicationForm)
			request.Header.Set("HX-Request", "true")
		})

		It(
			"saves editable fields and replaces the edit URL with the detail URL",
			func(ctx SpecContext) {
				service.EXPECT().
					Edit(request.Context(), &models.EditScheduleParams{Id: 42, Name: " Kitchen ", IntervalDays: 14}).
					Return(&models.Schedule{Id: 42, Name: "Kitchen", IntervalDays: 14, ChoreTemplateId: 7}, nil).
					Once()
				view.EXPECT().Detail(viewmodels.Detail{Schedule: apimodels.Representation{
					Response: apimodels.Response{
						Id:              42,
						Name:            "Kitchen",
						IntervalDays:    14,
						ChoreTemplateId: 7,
						Links: api.Relations{
							{
								Rel:  "self",
								Href: "/schedules/42",
							},
							{Rel: "collection", Href: "/schedules"},
							{Rel: "chore_template", Href: "/chore-templates/7"},
						},
					},
					Actions: api.Relations{
						{Rel: "edit", Href: "/schedules/42"},
						{Rel: "deactivate", Href: "/schedules/42/deactivation"},
					},
				}, Notice: "Schedule updated."}, false).Return(templ.Raw("saved")).Once()
				response := httptest.NewRecorder()
				app.ServeHTTP(response, request)
				Expect(response.Code).To(Equal(http.StatusOK))
				Expect(response.Header().Get("HX-Replace-Url")).To(Equal("/schedules/42"))
				Expect(response.Body.String()).To(Equal("saved"))
			},
		)

		DescribeTable(
			"preserves submitted values on validation and storage failures",
			func(ctx SpecContext, err error, status int, nameError, intervalError, summary string) {
				service.EXPECT().
					Edit(request.Context(), &models.EditScheduleParams{Id: 42, Name: " Kitchen ", IntervalDays: 14}).
					Return(nil, err).
					Once()
				view.EXPECT().EditForm(viewmodels.EditForm{
					ActionHref:        "/schedules/42",
					CancelHref:        "/schedules/42",
					Name:              " Kitchen ",
					IntervalDays:      "14",
					NameError:         nameError,
					IntervalDaysError: intervalError,
					SummaryError:      summary,
				}, false).Return(templ.Raw("feedback")).Once()
				response := httptest.NewRecorder()
				app.ServeHTTP(response, request)
				Expect(response.Code).To(Equal(status))
				Expect(response.Header().Get("HX-Replace-Url")).To(BeEmpty())
				Expect(response.Body.String()).To(Equal("feedback"))
			},
			Entry(
				"name conflict",
				models.ErrNameConflict,
				409,
				"An active schedule already uses this name.",
				"",
				"",
			),
			Entry(
				"invalid name",
				validationerrors.ErrInvalidName,
				422,
				"Enter a name between 1 and 200 characters.",
				"",
				"",
			),
			Entry(
				"invalid interval",
				validationerrors.ErrInvalidInterval,
				422,
				"",
				"Enter a whole number between 1 and 3650.",
				"",
			),
			Entry(
				"storage failure",
				errors.New("storage"),
				500,
				"",
				"",
				"The schedule could not be updated. Try again.",
			),
		)

		DescribeTable(
			"rejects non-integer intervals without mutating",
			func(ctx SpecContext, interval string) {
				request = httptest.NewRequestWithContext(
					ctx,
					http.MethodPatch,
					"/schedules/42",
					strings.NewReader(url.Values{
						"name": {"Kitchen"}, "interval_days": {interval},
					}.Encode()),
				)
				request.Header.Set("Accept", "text/html")
				request.Header.Set("Content-Type", echo.MIMEApplicationForm)
				request.Header.Set("HX-Request", "true")
				view.EXPECT().EditForm(viewmodels.EditForm{
					ActionHref:        "/schedules/42",
					CancelHref:        "/schedules/42",
					Name:              "Kitchen",
					IntervalDays:      interval,
					IntervalDaysError: "Enter a whole number between 1 and 3650.",
				}, false).Return(templ.Raw("interval error")).Once()
				response := httptest.NewRecorder()
				app.ServeHTTP(response, request)
				Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
			},
			Entry("empty", ""),
			Entry("fraction", "1.5"),
			Entry("text", "weekly"),
			Entry("overflow", "999999999999999999999999"),
		)

		DescribeTable(
			"rejects schedules that are no longer editable",
			func(ctx SpecContext, err error, status int) {
				service.EXPECT().
					Edit(request.Context(), &models.EditScheduleParams{Id: 42, Name: " Kitchen ", IntervalDays: 14}).
					Return(nil, err).
					Once()
				view.EXPECT().
					Error(viewmodels.Error{Message: err.Error(), Links: api.Relations{{Rel: "collection", Href: "/schedules"}}}, false).
					Return(templ.Raw("unavailable")).
					Once()
				response := httptest.NewRecorder()
				app.ServeHTTP(response, request)
				Expect(response.Code).To(Equal(status))
			},
			Entry("deactivated", models.ErrInactive, 422),
			Entry("missing", models.ErrNotFound, 404),
		)
	})
})
