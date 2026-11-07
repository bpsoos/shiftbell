package chores_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

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

var _ = Describe("Set chore assignment", func() {
	var service *MockService
	var view *MockView
	var app *echo.Echo
	BeforeEach(func() {
		service = NewMockService(GinkgoT())
		view = NewMockView(GinkgoT())
		handler := endpoint.NewHandler(
			&endpoint.HandlerDeps{Service: service, View: view, ProfilesUIEnabled: true},
		)
		app = echo.New()
		app.PUT("/chores/:id/assignment", handler.SetAssignment)
	})

	DescribeTable(
		"replaces assignments through JSON",
		func(ctx SpecContext, body string, owner int) {
			service.EXPECT().
				SetAssignment(ctx, &models.SetAssignmentParams{Id: 42, OwnerProfileId: owner}).
				Return(&models.Chore{Id: 42, Status: models.ChoreStatusCompleted, OwnerProfileId: &owner}, nil).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPut,
				"/chores/42/assignment",
				strings.NewReader(body),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			var representation api.Representation
			Expect(json.Unmarshal(response.Body.Bytes(), &representation)).To(Succeed())
			Expect(representation.OwnerProfileId).To(Equal(&owner))
			Expect(
				representation.Actions.Href("set_assignment"),
			).To(Equal("/chores/42/assignment"))
			Expect(
				representation.Actions.Href("clear_assignment"),
			).To(Equal("/chores/42/assignment"))
		},
		Entry("assign", `{"owner_profile_id":7}`, 7),
	)

	DescribeTable(
		"renders the updated chore after assigning through HTMX",
		func(ctx SpecContext, status models.ChoreStatus) {
			owner := 7
			service.EXPECT().
				SetAssignment(ctx, &models.SetAssignmentParams{Id: 42, OwnerProfileId: owner}).
				Return(&models.Chore{Id: 42, Status: status, OwnerProfileId: &owner, OwnerProfileName: "Alex"}, nil).
				Once()
			view.EXPECT().
				Detail(mock.Anything, false).
				Run(func(model viewmodels.Detail, fullPage bool) {
					Expect(model.Chore.Id).To(Equal(42))
					Expect(model.Chore.OwnerProfileId).To(Equal(&owner))
					Expect(model.Chore.OwnerProfileName).To(Equal("Alex"))
					Expect(model.Chore.Status).To(Equal(status))
					Expect(
						model.Chore.Actions.Href("clear_assignment"),
					).To(Equal("/chores/42/assignment"))
				}).
				Return(templ.Raw("assignment sentinel")).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPut,
				"/chores/42/assignment",
				strings.NewReader("owner_profile_id=7"),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Header().Get("HX-Push-Url")).To(Equal("/chores/42"))
			Expect(response.Body.String()).To(Equal("assignment sentinel"))
		},
		Entry("active chore", models.ChoreStatusActive),
		Entry("completed chore", models.ChoreStatusCompleted),
	)

	It("hides HTML assignment when profiles UI is disabled", func(ctx SpecContext) {
		handler := endpoint.NewHandler(
			&endpoint.HandlerDeps{Service: service, View: view},
		)
		app := echo.New()
		app.PUT("/chores/:id/assignment", handler.SetAssignment)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPut,
			"/chores/42/assignment",
			strings.NewReader("owner_profile_id=7"),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusNotFound))
	})

	DescribeTable(
		"rejects invalid requests without mutating",
		func(ctx SpecContext, path, contentType, body string, status int) {
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPut,
				path,
				strings.NewReader(body),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", contentType)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
		},
		Entry("missing owner", "/chores/42/assignment", "application/json", `{}`, 422),
		Entry(
			"negative owner",
			"/chores/42/assignment",
			"application/json",
			`{"owner_profile_id":-1}`,
			422,
		),
		Entry("malformed JSON", "/chores/42/assignment", "application/json", `{`, 400),
		Entry(
			"invalid chore ID",
			"/chores/no/assignment",
			"application/json",
			`{"owner_profile_id":7}`,
			422,
		),
		Entry(
			"unsupported body",
			"/chores/42/assignment",
			"text/plain",
			"owner_profile_id=7",
			415,
		),
	)

	DescribeTable(
		"maps service errors",
		func(ctx SpecContext, failure error, status int) {
			service.EXPECT().
				SetAssignment(ctx, &models.SetAssignmentParams{Id: 42, OwnerProfileId: 7}).
				Return(nil, failure).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPut,
				"/chores/42/assignment",
				strings.NewReader(`{"owner_profile_id":7}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
		},
		Entry("missing profile", models.ErrInvalidOwnerProfile, 422),
		Entry("missing chore", models.ErrNotFound, 404),
		Entry("unexpected failure", errors.New("private details"), 500),
	)

	DescribeTable(
		"renders HTML errors through the existing error view",
		func(ctx SpecContext, htmx string, fullPage bool) {
			service.EXPECT().
				SetAssignment(ctx, &models.SetAssignmentParams{Id: 42, OwnerProfileId: 7}).
				Return(nil, models.ErrInvalidOwnerProfile).
				Once()
			view.EXPECT().
				Error(viewmodels.Error{Message: models.ErrInvalidOwnerProfile.Error()}, fullPage).
				Return(templ.Raw("error sentinel")).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPut,
				"/chores/42/assignment",
				strings.NewReader("owner_profile_id=7"),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", htmx)
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
			Expect(response.Body.String()).To(Equal("error sentinel"))
		},
		Entry("ordinary HTML", "", true),
		Entry("HTMX", "true", false),
	)

	DescribeTable(
		"rejects unsupported representations",
		func(ctx SpecContext, accept string) {
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPut,
				"/chores/42/assignment",
				strings.NewReader(`{"owner_profile_id":7}`),
			)
			request.Header.Set("Accept", accept)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		},
		Entry("unsupported accept", "application/xml"),
	)
})
