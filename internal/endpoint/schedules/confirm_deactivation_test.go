package schedules_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/schedules"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	confirmationmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Schedule deactivation confirmation", func() {
	var service *MockService
	var view *MockView
	var app *echo.Echo
	var request *http.Request
	BeforeEach(func(ctx SpecContext) {
		service = NewMockService(GinkgoT())
		view = NewMockView(GinkgoT())
		handler := schedules.NewHandler(
			&schedules.HandlerDeps{Service: service, View: view},
		)
		app = echo.New()
		app.GET("/schedules/:id/deactivation", handler.ConfirmDeactivation)
		request = httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/schedules/42/deactivation",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
	})

	It(
		"requests confirmation without mutating",
		func(ctx SpecContext) {
			service.EXPECT().
				Get(request.Context(), 42).
				Return(&models.Schedule{Id: 42, Name: "Kitchen"}, nil).
				Once()
			view.EXPECT().ConfirmationDialog(confirmationmodels.Dialog{
				Heading:      "Deactivate schedule?",
				Prompt:       "Deactivate",
				Name:         "Kitchen",
				ActionHref:   "/schedules/42/deactivation",
				ActionMethod: "put",
				ActionLabel:  "Deactivate permanently",
				Icon:         "archive",
			}).Return(templ.Raw("confirmation")).Once()
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("confirmation"))
		},
	)

	It(
		"does not offer deactivation again for an inactive schedule",
		func(ctx SpecContext) {
			now := time.Now()
			service.EXPECT().
				Get(request.Context(), 42).
				Return(&models.Schedule{Id: 42, DeactivatedAt: &now}, nil).
				Once()
			view.EXPECT().ConfirmationDialog(confirmationmodels.Dialog{
				Heading:    "Schedule already deactivated",
				Supporting: []string{"This schedule is permanently deactivated."},
			}).Return(templ.Raw("inactive")).Once()
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		},
	)

	DescribeTable(
		"keeps load failures inside the dialog",
		func(ctx SpecContext, err error, status int) {
			service.EXPECT().Get(request.Context(), 42).Return(nil, err).Once()
			view.EXPECT().ConfirmationDialog(confirmationmodels.Dialog{
				Heading: "Unable to deactivate schedule",
				Error:   "The schedule could not be loaded. Close this dialog and try again.",
			}).Return(templ.Raw("load error")).Once()
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
		},
		Entry("missing", models.ErrNotFound, 404),
		Entry("storage", errors.New("storage"), 500),
	)

	DescribeTable("requires HTMX HTML", func(ctx SpecContext, accept, hx string) {
		request.Header.Set("Accept", accept)
		request.Header.Set("HX-Request", hx)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
	}, Entry("full page", "text/html", ""), Entry("JSON", hypermedia.MediaType, "true"), Entry("unsupported", "image/png", "true"))
})
