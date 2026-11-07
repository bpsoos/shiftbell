package chores_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/a-h/templ"
	endpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	api "github.com/bpsoos/shiftbell/internal/models/api/chores"
	models "github.com/bpsoos/shiftbell/internal/models/chores"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Clear chore assignment", func() {
	var service *MockService
	var view *MockView
	var app *echo.Echo

	BeforeEach(func() {
		service = NewMockService(GinkgoT())
		view = NewMockView(GinkgoT())
		handler := endpoint.NewHandler(
			&endpoint.HandlerDeps{Service: service, View: view},
		)
		app = echo.New()
		app.DELETE("/chores/:id/assignment", handler.ClearAssignment)
	})

	DescribeTable(
		"clears ownership without a request body",
		func(ctx SpecContext, status models.ChoreStatus) {
			service.EXPECT().
				ClearAssignment(ctx, 42).
				Return(&models.Chore{Id: 42, Status: status}, nil).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodDelete,
				"/chores/42/assignment",
				nil,
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			var representation api.Representation
			Expect(json.Unmarshal(response.Body.Bytes(), &representation)).To(Succeed())
			Expect(representation.Id).To(Equal(42))
			Expect(representation.Status).To(Equal(status))
			Expect(representation.OwnerProfileId).To(BeNil())
			Expect(representation.Actions.Href("clear_assignment")).To(BeEmpty())
			Expect(
				representation.Actions.Href("set_assignment"),
			).To(Equal("/chores/42/assignment"))
		},
		Entry("active chore", models.ChoreStatusActive),
		Entry("completed chore", models.ChoreStatusCompleted),
	)

	DescribeTable("renders the HTML placeholder", func(ctx SpecContext, htmx string) {
		service.EXPECT().
			ClearAssignment(ctx, 42).
			Return(&models.Chore{Id: 42}, nil).
			Once()
		view.EXPECT().Assignment(mock.Anything).Run(func(model viewmodels.Assignment) {
			Expect(model.Chore.Id).To(Equal(42))
			Expect(model.Chore.OwnerProfileId).To(BeNil())
		}).Return(templ.Raw("assignment sentinel")).Once()
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodDelete,
			"/chores/42/assignment",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", htmx)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("assignment sentinel"))
	}, Entry("ordinary HTML", ""), Entry("HTMX", "true"))

	DescribeTable(
		"maps service errors",
		func(ctx SpecContext, failure error, status int) {
			service.EXPECT().
				ClearAssignment(ctx, 42).
				Return(nil, failure).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodDelete,
				"/chores/42/assignment",
				nil,
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
		},
		Entry("missing chore", models.ErrNotFound, http.StatusNotFound),
		Entry("storage failure", errors.New("storage"), http.StatusInternalServerError),
	)

	DescribeTable(
		"rejects invalid requests before calling the service",
		func(ctx SpecContext, path, accept string, status int) {
			request := httptest.NewRequestWithContext(ctx, http.MethodDelete, path, nil)
			request.Header.Set("Accept", accept)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
		},
		Entry(
			"invalid ID",
			"/chores/no/assignment",
			hypermedia.MediaType,
			http.StatusUnprocessableEntity,
		),
		Entry(
			"unsupported accept",
			"/chores/42/assignment",
			"application/xml",
			http.StatusNotAcceptable,
		),
	)
})
