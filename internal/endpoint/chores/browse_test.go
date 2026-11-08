package chores_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/a-h/templ"
	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	profilemodels "github.com/bpsoos/shiftbell/internal/models/profiles"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Browse chores", func() {
	It("reports a profile-state lookup failure", func(ctx SpecContext) {
		profiles := NewMockProfileService(GinkgoT())
		profiles.EXPECT().
			HasMultipleProfiles(ctx).
			Return(false, errors.New("storage")).
			Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().
			Error(mock.Anything, true).
			Run(func(model choreviewmodels.Error, fullPage bool) {
				Expect(
					model.Message,
				).To(Equal("Could not load profile settings. Try again."))
			}).
			Return(templ.Raw("profile state failed")).
			Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: NewMockService(
				GinkgoT(),
			),
			ProfileService:    profiles,
			View:              view,
			ProfilesUIEnabled: true,
		})
		app := echo.New()
		app.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores", nil)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusInternalServerError))
		Expect(response.Body.String()).To(Equal("profile state failed"))
	})

	DescribeTable(
		"provides collection visibility using the lightweight profile check",
		func(ctx SpecContext, multiple bool) {
			service := NewMockService(GinkgoT())
			view := NewMockView(GinkgoT())
			profiles := NewMockProfileService(GinkgoT())
			profiles.EXPECT().HasMultipleProfiles(ctx).Return(multiple, nil).Once()
			service.EXPECT().
				Browse(ctx, &choremodels.BrowseChoresParams{Search: "Kitchen", Offset: 2, Limit: 20}).
				Return(&choremodels.ChorePage{Chores: []choremodels.Chore{{Id: 42, Status: choremodels.ChoreStatusActive}}}, nil).
				Once()
			view.EXPECT().
				Collection(mock.Anything, false).
				Run(func(model choreviewmodels.Collection, fullPage bool) {
					Expect(model.Mine).To(BeFalse())
					Expect(model.HasMultipleProfiles).To(Equal(multiple))
					Expect(model.AllHref).To(Equal("/chores?search=Kitchen"))
					Expect(model.MineHref).To(Equal("/chores?scope=mine&search=Kitchen"))
					Expect(
						model.Items[0].DetailHref,
					).To(Equal("/chores/42?return_to=" + url.QueryEscape("/chores?search=Kitchen&offset=2")))
				}).
				Return(templ.Raw("all chores")).
				Once()
			handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
				Service:           service,
				View:              view,
				ProfilesUIEnabled: true,
				ProfileService:    profiles,
			})
			app := echo.New()
			app.GET("/chores", handler.GetBatch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores?search=Kitchen&offset=2",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
		},
		Entry("single profile", false),
		Entry("multiple profiles", true),
	)

	DescribeTable(
		"resolves Mine to the active profile and preserves empty results",
		func(ctx SpecContext, available []profilemodels.Profile, cookie string, owner int, status choremodels.ChoreStatus) {
			service := NewMockService(GinkgoT())
			profiles := NewMockProfileService(GinkgoT())
			view := NewMockView(GinkgoT())
			profiles.EXPECT().List(ctx).Return(available, nil).Once()
			service.EXPECT().Browse(ctx, &choremodels.BrowseChoresParams{
				Status: status,
				OwnerFilter: choremodels.OwnerFilter{
					Mode: choremodels.OwnerFilterProfile,
					Id:   &owner,
				},
				Search: "Kitchen",
				Offset: 2,
				Limit:  2,
			}).Return(&choremodels.ChorePage{}, nil).Once()
			view.EXPECT().
				Collection(mock.Anything, false).
				Run(func(model choreviewmodels.Collection, fullPage bool) {
					Expect(model.Mine).To(BeTrue())
					Expect(model.HasMultipleProfiles).To(Equal(len(available) > 1))
					Expect(model.Items).To(BeEmpty())
					Expect(model.OwnerFilter.Id).To(Equal(&owner))
					Expect(
						model.AllHref,
					).To(Equal("/chores?limit=2&search=Kitchen&status=" + string(status)))
					Expect(
						model.MineHref,
					).To(Equal("/chores?limit=2&scope=mine&search=Kitchen&status=" + string(status)))
					Expect(
						model.PreviousHref,
					).To(Equal("/chores?limit=2&offset=0&scope=mine&search=Kitchen&status=" + string(status)))
				}).
				Return(templ.Raw("empty mine")).
				Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{
					Service:           service,
					View:              view,
					ProfileService:    profiles,
					ProfilesUIEnabled: true,
				},
			)
			app := echo.New()
			app.GET("/chores", handler.GetBatch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores?scope=mine&search=Kitchen&offset=2&limit=2&status="+string(
					status,
				),
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			request.AddCookie(&http.Cookie{Name: "shiftbell_profile", Value: cookie})
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("empty mine"))
			Expect(response.Header().Get("Cache-Control")).To(Equal("private, no-store"))
		},
		Entry(
			"selected profile",
			[]profilemodels.Profile{{Id: 7}, {Id: 8}},
			"7",
			7,
			choremodels.ChoreStatusActive,
		),
		Entry(
			"switched profile",
			[]profilemodels.Profile{{Id: 7}, {Id: 8}},
			"8",
			8,
			choremodels.ChoreStatusActive,
		),
		Entry(
			"implicit single profile",
			[]profilemodels.Profile{{Id: 7}},
			"",
			7,
			choremodels.ChoreStatusActive,
		),
		Entry(
			"single profile ignores stale selection",
			[]profilemodels.Profile{{Id: 7}},
			"8",
			7,
			choremodels.ChoreStatusActive,
		),
		Entry(
			"completion history",
			[]profilemodels.Profile{{Id: 7}, {Id: 8}},
			"7",
			7,
			choremodels.ChoreStatusCompleted,
		),
	)

	DescribeTable(
		"requires an active profile for Mine",
		func(ctx SpecContext, available []profilemodels.Profile, cookie, redirect string) {
			profiles := NewMockProfileService(GinkgoT())
			profiles.EXPECT().List(ctx).Return(available, nil).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{
					Service:           NewMockService(GinkgoT()),
					ProfileService:    profiles,
					ProfilesUIEnabled: true,
				},
			)
			app := echo.New()
			app.GET("/chores", handler.GetBatch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores?scope=mine",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			request.AddCookie(&http.Cookie{Name: "shiftbell_profile", Value: cookie})
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Header().Get("HX-Redirect")).To(Equal(redirect))
		},
		Entry("setup", []profilemodels.Profile{}, "", "/setup"),
		Entry(
			"missing selection",
			[]profilemodels.Profile{{Id: 7}, {Id: 8}},
			"",
			"/profiles/selection?return_to=%2Fchores%3Fscope%3Dmine",
		),
		Entry(
			"stale selection",
			[]profilemodels.Profile{{Id: 7}, {Id: 8}},
			"99",
			"/profiles/selection?return_to=%2Fchores%3Fscope%3Dmine",
		),
	)

	It(
		"shows profile lookup failures instead of falling back to all chores",
		func(ctx SpecContext) {
			profiles := NewMockProfileService(GinkgoT())
			profiles.EXPECT().List(ctx).Return(nil, errors.New("storage")).Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().
				Error(mock.Anything, false).
				Return(templ.Raw("profile lookup failed")).
				Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{
					Service:           NewMockService(GinkgoT()),
					ProfileService:    profiles,
					View:              view,
					ProfilesUIEnabled: true,
				},
			)
			app := echo.New()
			app.GET("/chores", handler.GetBatch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores?scope=mine",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusInternalServerError))
			Expect(response.Body.String()).To(Equal("profile lookup failed"))
		},
	)

	DescribeTable(
		"gates the Mine UI without reading profiles or chores",
		func(ctx SpecContext, enabled bool, accept string, status int) {
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{
					Service:           NewMockService(GinkgoT()),
					ProfileService:    NewMockProfileService(GinkgoT()),
					ProfilesUIEnabled: enabled,
				},
			)
			app := echo.New()
			app.GET("/chores", handler.GetBatch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores?scope=mine",
				nil,
			)
			request.Header.Set("Accept", accept)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
		},
		Entry("flag disabled", false, "text/html", http.StatusNotFound),
		Entry(
			"JSON uses explicit owner parameters",
			true,
			hypermedia.MediaType,
			http.StatusNotAcceptable,
		),
	)

	DescribeTable(
		"preserves ownership, search and status in pagination links",
		func(ctx SpecContext, ownerQuery string, filter choremodels.OwnerFilter) {
			service := NewMockService(GinkgoT())
			service.EXPECT().Browse(ctx, &choremodels.BrowseChoresParams{
				Status:      choremodels.ChoreStatusCompleted,
				OwnerFilter: filter,
				Search:      "Kitchen",
				Offset:      2,
				Limit:       2,
			}).Return(&choremodels.ChorePage{More: true}, nil).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			app := echo.New()
			app.GET("/chores", handler.GetBatch)
			href := "/chores?status=completed&" + ownerQuery + "&search=Kitchen&offset=2&limit=2"
			request := httptest.NewRequestWithContext(ctx, http.MethodGet, href, nil)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			var body struct {
				Links []struct{ Rel, Href string } `json:"_links"`
			}
			Expect(json.Unmarshal(response.Body.Bytes(), &body)).To(Succeed())
			Expect(body.Links).To(ConsistOf(
				And(HaveField("Rel", "self"), HaveField("Href", href)),
				And(
					HaveField("Rel", "next"),
					HaveField(
						"Href",
						"/chores?limit=2&offset=4&"+ownerQuery+"&search=Kitchen&status=completed",
					),
				),
				And(
					HaveField("Rel", "previous"),
					HaveField(
						"Href",
						"/chores?limit=2&offset=0&"+ownerQuery+"&search=Kitchen&status=completed",
					),
				),
			))
		},
		Entry(
			"all",
			"owner_mode=all",
			choremodels.OwnerFilter{Mode: choremodels.OwnerFilterAll},
		),
		Entry(
			"unassigned",
			"owner_mode=unassigned",
			choremodels.OwnerFilter{Mode: choremodels.OwnerFilterUnassigned},
		),
		Entry(
			"profile",
			"owner_id=7&owner_mode=profile",
			choremodels.OwnerFilter{Mode: choremodels.OwnerFilterProfile, Id: new(7)},
		),
	)

	It("returns invalid owner filters as JSON validation errors", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().
			Browse(ctx, &choremodels.BrowseChoresParams{OwnerFilter: choremodels.OwnerFilter{Mode: choremodels.OwnerFilterProfile, Id: new(7)}, Limit: 20}).
			Return(nil, validationerrors.ErrInvalidOwnerFilter).
			Once()
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		app := echo.New()
		app.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores?owner_mode=profile&owner_id=7",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(
			response.Body.Bytes(),
		).To(MatchJSON(`{"error":"invalid owner filter","_links":[],"_actions":[]}`))
	})

	DescribeTable("rejects malformed owner selectors before calling the service",
		func(ctx SpecContext, query string) {
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: NewMockService(GinkgoT())},
			)
			app := echo.New()
			app.GET("/chores", handler.GetBatch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores?"+query,
				nil,
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
			Expect(
				response.Body.Bytes(),
			).To(MatchJSON(`{"error":"invalid owner filter","_links":[],"_actions":[]}`))
		},
		Entry("non-numeric ID", "owner_mode=profile&owner_id=invalid"),
		Entry("empty ID", "owner_mode=profile&owner_id="),
		Entry("fraction", "owner_mode=profile&owner_id=1.5"),
		Entry("overflow", "owner_mode=profile&owner_id=999999999999999999999999"),
		Entry("whitespace", "owner_mode=profile&owner_id=%207"),
		Entry("repeated ID", "owner_mode=profile&owner_id=7&owner_id=8"),
	)

	It("rejects ambiguous repeated owner filters", func(ctx SpecContext) {
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: NewMockService(GinkgoT())},
		)
		app := echo.New()
		app.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores?owner_mode=profile&owner_mode=unassigned&owner_id=7",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(
			response.Body.Bytes(),
		).To(MatchJSON(`{"error":"invalid owner filter","_links":[],"_actions":[]}`))
	})

	It("provides ownership state to the view placeholder", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().
			Browse(ctx, &choremodels.BrowseChoresParams{OwnerFilter: choremodels.OwnerFilter{Mode: choremodels.OwnerFilterProfile, Id: new(7)}, Limit: 20}).
			Return(&choremodels.ChorePage{}, nil).
			Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().
			Collection(mock.Anything, false).
			Run(func(model choreviewmodels.Collection, fullPage bool) {
				Expect(
					model.OwnerFilter,
				).To(Equal(choremodels.OwnerFilter{Mode: choremodels.OwnerFilterProfile, Id: new(7)}))
				Expect(model.SelfHref).To(Equal("/chores?owner_mode=profile&owner_id=7"))
			}).
			Return(templ.Raw("placeholder")).
			Once()
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service, View: view},
		)
		app := echo.New()
		app.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores?owner_mode=profile&owner_id=7",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("placeholder"))
	})

	DescribeTable(
		"includes owners and unassigned chores without listing profiles",
		func(ctx SpecContext, status choremodels.ChoreStatus) {
			service := NewMockService(GinkgoT())
			view := NewMockView(GinkgoT())
			profiles := NewMockProfileService(GinkgoT())
			profiles.EXPECT().HasMultipleProfiles(ctx).Return(true, nil).Once()
			service.EXPECT().
				Browse(ctx, &choremodels.BrowseChoresParams{Status: status, Limit: 20}).
				Return(&choremodels.ChorePage{Chores: []choremodels.Chore{
					{
						Id:               42,
						Status:           status,
						OwnerProfileId:   new(7),
						OwnerProfileName: "Alex",
					},
					{Id: 43, Status: status},
				}}, nil).
				Once()
			view.EXPECT().
				Collection(mock.Anything, false).
				Run(func(model choreviewmodels.Collection, fullPage bool) {
					Expect(model.Items).To(HaveLen(2))
					Expect(model.Items[0].OwnerProfileId).To(Equal(new(7)))
					Expect(model.Items[0].OwnerProfileName).To(Equal("Alex"))
					Expect(model.Items[1].OwnerProfileId).To(BeNil())
					Expect(model.Items[1].OwnerProfileName).To(BeEmpty())
				}).
				Return(templ.Raw("owned collection")).
				Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{
					Service:           service,
					View:              view,
					ProfileService:    profiles,
					ProfilesUIEnabled: true,
				},
			)
			app := echo.New()
			app.GET("/chores", handler.GetBatch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores?status="+string(status),
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("owned collection"))
		},
		Entry("active", choremodels.ChoreStatusActive),
		Entry("completed", choremodels.ChoreStatusCompleted),
	)

	It("rejects a non-numeric offset", func(ctx SpecContext) {
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: NewMockService(GinkgoT())},
		)
		e := echo.New()
		e.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores?offset=invalid",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"error": "invalid offset",
			"_links": [],
			"_actions": []
		}`))
	})

	It("rejects a non-numeric limit", func(ctx SpecContext) {
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: NewMockService(GinkgoT())},
		)
		e := echo.New()
		e.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores?limit=invalid",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"error": "invalid limit",
			"_links": [],
			"_actions": []
		}`))
	})

	It("passes filters and pagination to the service", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Browse(ctx, &choremodels.BrowseChoresParams{
			Status: choremodels.ChoreStatusCompleted,
			Search: "Kitchen",
			Offset: 4,
			Limit:  7,
		}).Return(&choremodels.ChorePage{}, nil).Once()
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores?status=completed&search=Kitchen&offset=4&limit=7",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
	})

	It(
		"returns the active chore collection with its creation affordance",
		func(ctx SpecContext) {
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Browse(ctx, &choremodels.BrowseChoresParams{
				Offset: 0,
				Limit:  20,
			}).Return(&choremodels.ChorePage{
				Chores: []choremodels.Chore{
					{
						Id:          42,
						Status:      choremodels.ChoreStatusActive,
						Name:        "Kitchen",
						Description: "Clean counters",
						Deadline:    deadline,
					},
				},
			}, nil).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			e := echo.New()
			e.GET("/chores", handler.GetBatch)
			request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores", nil)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"items": [{
				"id": 42,
				"schedule_id": null,
				"owner_profile_id": null,
				"status": "active",
				"name": "Kitchen",
				"description": "Clean counters",
				"deadline": "2020-02-03",
				"completed_on": null,
				"_links": [
					{"rel": "self", "href": "/chores/42"},
					{"rel": "collection", "href": "/chores"}
				]
			}],
			"more": false,
			"_links": [{"rel": "self", "href": "/chores"}],
			"_actions": [{"rel": "create", "href": "/chores/new"}]
		}`))
		},
	)

	It(
		"passes search and status filters to the HTML collection",
		func(ctx SpecContext) {
			deadline := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Browse(ctx, &choremodels.BrowseChoresParams{
				Status: choremodels.ChoreStatusCompleted,
				Search: "Kitchen",
				Offset: 4,
				Limit:  7,
			}).Return(&choremodels.ChorePage{
				Chores: []choremodels.Chore{{
					Id:          42,
					Status:      choremodels.ChoreStatusCompleted,
					Name:        "Kitchen",
					Description: "Wash and fold",
					Deadline:    deadline,
				}},
			}, nil).Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().Collection(choreviewmodels.Collection{
				Items: []choreviewmodels.CollectionItem{{
					Status:      choremodels.ChoreStatusCompleted,
					Name:        "Kitchen",
					Description: "Wash and fold",
					Deadline:    "2026-08-15",
					DetailHref:  "/chores/42",
				}},
				SelfHref:        "/chores?status=completed&search=Kitchen&offset=4&limit=7",
				PreviousHref:    "/chores?limit=7&offset=0&search=Kitchen&status=completed",
				CreateHref:      "/chores/new",
				Status:          choremodels.ChoreStatusCompleted,
				Search:          "Kitchen",
				SearchOpen:      true,
				AutofocusSearch: true,
			}, false).Return(templ.Raw("collection sentinel")).Once()
			handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
				Service: service,
				View:    view,
			})
			e := echo.New()
			e.GET("/chores", handler.GetBatch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores?status=completed&search=Kitchen&offset=4&limit=7",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			request.Header.Set(
				"HX-Current-URL",
				"http://example.com/chores?status=completed&search=Kit",
			)
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("collection sentinel"))
		},
	)

	It("consumes the chore-and-template success flash", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Browse(ctx, &choremodels.BrowseChoresParams{
			Offset: 0,
			Limit:  20,
		}).Return(&choremodels.ChorePage{}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().Collection(choreviewmodels.Collection{
			Items:      []choreviewmodels.CollectionItem{},
			SelfHref:   "/chores",
			CreateHref: "/chores/new",
			Status:     choremodels.ChoreStatusActive,
			Notice:     "Chore added and template saved.",
		}, true).Return(templ.Raw("collection sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores", nil)
		request.Header.Set("Accept", "text/html")
		request.AddCookie(&http.Cookie{
			Name:  "shiftbell_flash",
			Value: "chore-and-template-created",
		})
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("collection sentinel"))
		cookies := response.Result().Cookies()
		Expect(cookies).To(HaveLen(1))
		Expect(cookies[0].Name).To(Equal("shiftbell_flash"))
		Expect(cookies[0].MaxAge).To(Equal(-1))
	})

	It("consumes the chore completion success flash", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Browse(ctx, &choremodels.BrowseChoresParams{
			Offset: 0,
			Limit:  20,
		}).Return(&choremodels.ChorePage{}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().Collection(choreviewmodels.Collection{
			Items:      []choreviewmodels.CollectionItem{},
			SelfHref:   "/chores",
			CreateHref: "/chores/new",
			Status:     choremodels.ChoreStatusActive,
			Notice:     "Chore completed.",
		}, true).Return(templ.Raw("collection sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores", nil)
		request.Header.Set("Accept", "text/html")
		request.AddCookie(&http.Cookie{
			Name:  "shiftbell_flash",
			Value: "chore-completed",
		})
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("collection sentinel"))
		cookies := response.Result().Cookies()
		Expect(cookies).To(HaveLen(1))
		Expect(cookies[0].Name).To(Equal("shiftbell_flash"))
		Expect(cookies[0].MaxAge).To(Equal(-1))
	})

	It("consumes the chore deletion success flash", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Browse(ctx, &choremodels.BrowseChoresParams{
			Offset: 0,
			Limit:  20,
		}).Return(&choremodels.ChorePage{}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().Collection(choreviewmodels.Collection{
			Items:      []choreviewmodels.CollectionItem{},
			SelfHref:   "/chores",
			CreateHref: "/chores/new",
			Status:     choremodels.ChoreStatusActive,
			Notice:     "Chore deleted.",
		}, true).Return(templ.Raw("collection sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores", nil)
		request.Header.Set("Accept", "text/html")
		request.AddCookie(&http.Cookie{
			Name:  "shiftbell_flash",
			Value: "chore-deleted",
		})
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("collection sentinel"))
	})
})
