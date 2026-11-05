package profiles_test

import (
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	validation "github.com/bpsoos/shiftbell/internal/models/validation"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupStatus", func() {
	It("advertises initial creation when required", func(ctx SpecContext) {
		handler, service, _ := newHandler()
		service.On("SetupRequired", ctx).Return(true, nil).Once()
		response := perform(
			ctx,
			handler.SetupStatus,
			http.MethodGet,
			"/setup",
			hypermedia.MediaType,
			"",
			"",
			"",
		)
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"required":true,
			"_links":[{"rel":"self","href":"/setup"},{"rel":"profiles","href":"/profiles"},{"rel":"chores","href":"/chores"}],
			"_actions":[{"rel":"create","href":"/setup"}]
		}`))
	})

	It("removes the setup action after any profile exists", func(ctx SpecContext) {
		handler, service, _ := newHandler()
		service.On("SetupRequired", ctx).Return(false, nil).Once()
		response := perform(
			ctx,
			handler.SetupStatus,
			http.MethodGet,
			"/setup",
			hypermedia.MediaType,
			"",
			"",
			"",
		)
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(ContainSubstring(`"required":false`))
		Expect(response.Body.String()).To(ContainSubstring(`"_actions":[]`))
	})

	DescribeTable(
		"renders the first-profile form",
		func(ctx SpecContext, hx string, fullPage bool) {
			handler, service, view := newHandler()
			service.On("SetupRequired", ctx).Return(true, nil).Once()
			view.On("Form", viewmodels.Form{ActionHref: "/setup", Setup: true}, fullPage).
				Return(templ.Raw("What is your name?")).
				Once()
			response := perform(
				ctx,
				handler.SetupStatus,
				http.MethodGet,
				"/setup",
				"text/html",
				"",
				"",
				hx,
			)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("What is your name?"))
		},
		Entry("page", "", true),
		Entry("fragment", "true", false),
	)

	DescribeTable(
		"redirects browsers once setup is complete",
		func(ctx SpecContext, hx string, status int, header string) {
			handler, service, _ := newHandler()
			service.On("SetupRequired", ctx).Return(false, nil).Once()
			response := perform(
				ctx,
				handler.SetupStatus,
				http.MethodGet,
				"/setup",
				"text/html",
				"",
				"",
				hx,
			)
			Expect(response.Code).To(Equal(status))
			Expect(response.Header().Get(header)).To(Equal("/chores"))
		},
		Entry("browser", "", http.StatusSeeOther, "Location"),
		Entry("HTMX", "true", http.StatusOK, "HX-Redirect"),
	)

	It("reports storage failures", func(ctx SpecContext) {
		handler, service, _ := newHandler()
		service.On("SetupRequired", ctx).Return(false, errors.New("secret")).Once()
		response := perform(
			ctx,
			handler.SetupStatus,
			http.MethodGet,
			"/setup",
			hypermedia.MediaType,
			"",
			"",
			"",
		)
		Expect(response.Code).To(Equal(http.StatusInternalServerError))
		Expect(response.Body.String()).NotTo(ContainSubstring("secret"))
	})

	It("rejects unsupported representations", func(ctx SpecContext) {
		handler, _, _ := newHandler()
		response := perform(
			ctx,
			handler.SetupStatus,
			http.MethodGet,
			"/setup",
			"application/xml",
			"",
			"",
			"",
		)
		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
	})
})

var _ = Describe("Setup", func() {
	It("creates the first profile through the setup service", func(ctx SpecContext) {
		handler, service, _ := newHandler()
		service.On("Setup", ctx, &models.CreateParams{Name: " Alex "}).
			Return(&models.Profile{Id: 1, Name: "Alex"}, nil).
			Once()
		response := perform(
			ctx,
			handler.Setup,
			http.MethodPost,
			"/setup",
			hypermedia.MediaType,
			"application/json",
			`{"name":" Alex "}`,
			"",
		)
		Expect(response.Code).To(Equal(http.StatusCreated))
		Expect(
			response.Body.Bytes(),
		).To(MatchJSON(`{"id":1,"name":"Alex","_links":[{"rel":"collection","href":"/profiles"}]}`))
	})

	DescribeTable(
		"redirects HTML clients to chores after setup",
		func(ctx SpecContext, hx string, status int, header string) {
			handler, service, _ := newHandler()
			service.On("Setup", ctx, &models.CreateParams{Name: "Alex"}).
				Return(&models.Profile{Id: 1, Name: "Alex"}, nil).
				Once()
			response := perform(
				ctx,
				handler.Setup,
				http.MethodPost,
				"/setup",
				"text/html",
				"application/x-www-form-urlencoded",
				"name=Alex",
				hx,
			)
			Expect(response.Code).To(Equal(status))
			Expect(response.Header().Get(header)).To(Equal("/chores"))
		},
		Entry("browser", "", http.StatusSeeOther, "Location"),
		Entry("HTMX", "true", http.StatusOK, "HX-Redirect"),
	)

	It("reports a competing setup as a conflict", func(ctx SpecContext) {
		handler, service, _ := newHandler()
		service.On("Setup", ctx, &models.CreateParams{Name: "Alex"}).
			Return(nil, models.ErrAlreadyInitialized).
			Once()
		response := perform(
			ctx,
			handler.Setup,
			http.MethodPost,
			"/setup",
			hypermedia.MediaType,
			"application/json",
			`{"name":"Alex"}`,
			"",
		)
		Expect(response.Code).To(Equal(http.StatusConflict))
		Expect(
			response.Body.String(),
		).To(ContainSubstring(models.ErrAlreadyInitialized.Error()))
	})

	It(
		"offers a continuation when another browser has completed setup",
		func(ctx SpecContext) {
			handler, service, view := newHandler()
			service.On("Setup", ctx, &models.CreateParams{Name: "Alex"}).
				Return(nil, models.ErrAlreadyInitialized).
				Once()
			view.On("Form", viewmodels.Form{
				ActionHref: "/setup", CancelHref: "/chores", Setup: true,
				Name:         viewmodels.Field{Value: "Alex"},
				SummaryError: "Setup is already complete. Continue to chores.",
			}, true).Return(templ.Raw("continue")).Once()
			response := perform(
				ctx,
				handler.Setup,
				http.MethodPost,
				"/setup",
				"text/html",
				"application/x-www-form-urlencoded",
				"name=Alex",
				"",
			)
			Expect(response.Code).To(Equal(http.StatusConflict))
			Expect(response.Body.String()).To(Equal("continue"))
		},
	)

	It(
		"keeps setup context when rendering name validation errors",
		func(ctx SpecContext) {
			handler, service, view := newHandler()
			service.On("Setup", ctx, &models.CreateParams{Name: "Alex Smith"}).
				Return(nil, validation.ErrInvalidName).
				Once()
			view.On("Form", viewmodels.Form{
				ActionHref: "/setup",
				Setup:      true,
				Name: viewmodels.Field{
					Value: "Alex Smith",
					Error: "Enter 1–200 printable characters with no whitespace.",
				},
			}, true).Return(templ.Raw("invalid name")).Once()
			response := perform(
				ctx,
				handler.Setup,
				http.MethodPost,
				"/setup",
				"text/html",
				"application/x-www-form-urlencoded",
				"name=Alex+Smith",
				"",
			)
			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		},
	)
})
