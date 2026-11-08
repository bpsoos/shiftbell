package profiles_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/profiles"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Selection status", func() {
	DescribeTable(
		"rejects unavailable requests without reading profiles",
		func(ctx SpecContext, enabled bool, accept string, status int) {
			handler := profiles.NewHandler(
				&profiles.HandlerDeps{Service: NewMockService(GinkgoT())},
				&profiles.HandlerConfig{UIEnabled: enabled},
			)
			app := echo.New()
			app.GET("/profiles/selection/status", handler.SelectionStatus)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/profiles/selection/status",
				nil,
			)
			request.Header.Set("Accept", accept)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
		},
		Entry("flag disabled", false, "text/html", 404),
		Entry("JSON request", true, hypermedia.MediaType, 406),
	)

	It("redirects the browser to setup when no profiles exist", func(ctx SpecContext) {
		handler, service, _ := newHandler()
		service.EXPECT().List(ctx).Return(nil, nil).Once()
		app := echo.New()
		app.GET("/profiles/selection/status", handler.SelectionStatus)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/profiles/selection/status",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Header().Get("HX-Redirect")).To(Equal("/setup"))
		Expect(response.Body.String()).To(BeEmpty())
	})

	It("leaves the slot empty for a single implicit profile", func(ctx SpecContext) {
		handler, service, _ := newHandler()
		service.EXPECT().
			List(ctx).
			Return([]models.Profile{{Id: 7, Name: "Alex"}}, nil).
			Once()
		app := echo.New()
		app.GET("/profiles/selection/status", handler.SelectionStatus)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/profiles/selection/status",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(BeEmpty())
		Expect(response.Header().Get("HX-Redirect")).To(BeEmpty())
		Expect(response.Header().Get("Cache-Control")).To(Equal("private, no-store"))
		Expect(response.Result().Cookies()).To(BeEmpty())
	})

	DescribeTable(
		"redirects to the picker without a valid selection",
		func(ctx SpecContext, cookie string) {
			handler, service, _ := newHandler()
			service.EXPECT().
				List(ctx).
				Return([]models.Profile{{Id: 7, Name: "Alex"}, {Id: 8, Name: "Sam"}}, nil).
				Once()
			app := echo.New()
			app.GET("/profiles/selection/status", handler.SelectionStatus)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"http://example.com/profiles/selection/status",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			request.Header.Set(
				"HX-Current-URL",
				"http://example.com/chores?status=completed",
			)
			request.AddCookie(&http.Cookie{Name: "shiftbell_profile", Value: cookie})
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(
				response.Header().Get("HX-Redirect"),
			).To(Equal("/profiles/selection?return_to=%2Fchores%3Fstatus%3Dcompleted"))
		},
		Entry("missing selection", ""),
		Entry("stale selection", "99"),
	)

	It("renders only the selected-profile snippet", func(ctx SpecContext) {
		handler, service, view := newHandler()
		service.EXPECT().
			List(ctx).
			Return([]models.Profile{{Id: 7, Name: "Alex"}, {Id: 8, Name: "Sam"}}, nil).
			Once()
		view.EXPECT().SelectedProfile(viewmodels.Navigation{
			Profile: models.Profile{
				Id:   8,
				Name: "Sam",
			},
			Profiles: []models.Profile{{Id: 7, Name: "Alex"}, {Id: 8, Name: "Sam"}},
		}).Return(templ.Raw("selected profile sentinel")).Once()
		app := echo.New()
		app.GET("/profiles/selection/status", handler.SelectionStatus)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://example.com/profiles/selection/status",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		request.Header.Set("HX-Current-URL", "http://example.com/schedules")
		request.AddCookie(&http.Cookie{Name: "shiftbell_profile", Value: "8"})
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("selected profile sentinel"))
		Expect(response.Header().Get("Cache-Control")).To(Equal("private, no-store"))
		Expect(response.Header().Get("HX-Redirect")).To(BeEmpty())
	})

	It("ignores a current-page URL from another host", func(ctx SpecContext) {
		handler, service, _ := newHandler()
		service.EXPECT().List(ctx).Return([]models.Profile{{Id: 7}, {Id: 8}}, nil).Once()
		app := echo.New()
		app.GET("/profiles/selection/status", handler.SelectionStatus)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://example.com/profiles/selection/status",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		request.Header.Set("HX-Current-URL", "https://other.example/schedules")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(
			response.Header().Get("HX-Redirect"),
		).To(Equal("/profiles/selection?return_to=%2Fchores"))
	})

	It(
		"returns a lookup failure without inserting a page into the slot",
		func(ctx SpecContext) {
			handler, service, _ := newHandler()
			service.EXPECT().List(ctx).Return(nil, errors.New("storage failure")).Once()
			app := echo.New()
			app.GET("/profiles/selection/status", handler.SelectionStatus)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/profiles/selection/status",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusInternalServerError))
			Expect(response.Body.String()).To(BeEmpty())
		},
	)
})
