package choretemplates_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
	choretemplatesendpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choretemplateapimodels "github.com/bpsoos/shiftbell/internal/models/api/choretemplates"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBrowse(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chore Templates Endpoint Suite")
}

var _ = Describe("Browse chore templates", func() {
	It("ignores API-only filters for the HTML collection", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Browse(ctx, &choretemplatemodels.BrowseChoreTemplatesParams{
			Filter: choretemplatemodels.ChoreTemplateFilterActive,
			Offset: 4,
			Limit:  7,
		}).Return(&choretemplatemodels.ChoreTemplatePage{}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().Collection(viewmodels.Collection{
			Collection: choretemplateapimodels.CollectionResponse{
				Items: []choretemplateapimodels.Response{},
				Links: api.Relations{
					{Rel: "self", Href: "/chore-templates?limit=7&offset=4"},
					{Rel: "previous", Href: "/chore-templates?limit=7&offset=0"},
				},
				Actions: api.Relations{
					{Rel: "create", Href: "/chore-templates"},
				},
			},
		}, true).Return(templ.Raw("collection sentinel")).Once()
		handler := choretemplatesendpoint.NewHandler(&choretemplatesendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.GET("/chore-templates", handler.Browse)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chore-templates?state=deactivated&search=Kitchen&offset=4&limit=7",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("collection sentinel"))
	})

	It("ignores API-only search for the HTML picker", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Browse(ctx, &choretemplatemodels.BrowseChoreTemplatesParams{
			Filter: choretemplatemodels.ChoreTemplateFilterActive,
			Offset: 4,
			Limit:  7,
		}).Return(&choretemplatemodels.ChoreTemplatePage{}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().Picker(viewmodels.Picker{
			Collection: choretemplateapimodels.PickerCollectionResponse{
				Items: []choretemplateapimodels.PickerItemResponse{},
				Links: api.Relations{
					{
						Rel:  "self",
						Href: "/chore-templates?limit=7&offset=4&picker=1",
					},
					{
						Rel:  "previous",
						Href: "/chore-templates?limit=7&offset=0&picker=1",
					},
				},
			},
			BackHref:   "/chores/new",
			ManualHref: "/chores/new?source=manual",
		}, true).Return(templ.Raw("picker sentinel")).Once()
		handler := choretemplatesendpoint.NewHandler(&choretemplatesendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.GET("/chore-templates", handler.Browse)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chore-templates?picker=1&search=Kitchen&offset=4&limit=7",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("picker sentinel"))
	})

	It("preserves search and state behavior for vendor JSON", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Browse(ctx, &choretemplatemodels.BrowseChoreTemplatesParams{
			Filter: choretemplatemodels.ChoreTemplateFilterDeactivated,
			Search: "Kitchen",
			Offset: 4,
			Limit:  7,
		}).Return(&choretemplatemodels.ChoreTemplatePage{}, nil).Once()
		handler := choretemplatesendpoint.NewHandler(&choretemplatesendpoint.HandlerDeps{
			Service: service,
		})
		e := echo.New()
		e.GET("/chore-templates", handler.Browse)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chore-templates?state=deactivated&search=Kitchen&offset=4&limit=7",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"items": [],
			"more": false,
			"_links": [
				{
					"rel": "self",
					"href": "/chore-templates?state=deactivated&search=Kitchen&offset=4&limit=7"
				},
				{
					"rel": "previous",
					"href": "/chore-templates?limit=7&offset=0&search=Kitchen&state=deactivated"
				}
			],
			"_actions": [{"rel": "create", "href": "/chore-templates"}]
		}`))
	})

	It("consumes the template deactivation success flash", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Browse(ctx, &choretemplatemodels.BrowseChoreTemplatesParams{
			Filter: choretemplatemodels.ChoreTemplateFilterActive,
			Offset: 0,
			Limit:  20,
		}).Return(&choretemplatemodels.ChoreTemplatePage{}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().Collection(viewmodels.Collection{
			Collection: choretemplateapimodels.CollectionResponse{
				Items: []choretemplateapimodels.Response{},
				Links: api.Relations{{Rel: "self", Href: "/chore-templates"}},
				Actions: api.Relations{
					{Rel: "create", Href: "/chore-templates"},
				},
			},
			Notice: "Template deactivated.",
		}, true).Return(templ.Raw("collection sentinel")).Once()
		handler := choretemplatesendpoint.NewHandler(&choretemplatesendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.GET("/chore-templates", handler.Browse)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chore-templates",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.AddCookie(&http.Cookie{
			Name:  "shiftbell_template_flash",
			Value: "template-deactivated",
		})
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("collection sentinel"))
	})
})
