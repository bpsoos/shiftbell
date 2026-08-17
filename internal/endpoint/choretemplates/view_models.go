package choretemplates

import (
	"github.com/bpsoos/shiftbell/internal/endpoint/routes"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
)

func collectionViewModel(
	representation collectionResponse,
	filter models.ChoreTemplateFilter,
	search string,
	searchOpen bool,
	autofocusSearch bool,
) viewmodels.Collection {
	if filter == "" {
		filter = models.ChoreTemplateFilterActive
	}
	items := make([]viewmodels.CollectionItem, len(representation.Items))
	for i, item := range representation.Items {
		items[i] = viewmodels.CollectionItem{
			Name:        item.Name,
			Description: item.Description,
			DetailHref:  item.Links.Href("self"),
		}
	}
	return viewmodels.Collection{
		Items:           items,
		PreviousHref:    representation.Links.Href("previous"),
		NextHref:        representation.Links.Href("next"),
		Filter:          filter,
		Search:          search,
		SearchOpen:      searchOpen,
		AutofocusSearch: autofocusSearch,
	}
}

func pickerViewModel(
	representation pickerCollectionResponse,
	search string,
	autofocusSearch bool,
) viewmodels.Picker {
	items := make([]viewmodels.PickerItem, len(representation.Items))
	for i, item := range representation.Items {
		items[i] = viewmodels.PickerItem{
			Name: item.Name,
			SelectHref: (routes.ChoreCreation{
				ChoreTemplateId: item.Id,
				Recurrence:      routes.ChoreCreationRecurrenceOneOff,
			}).Href(),
		}
	}
	return viewmodels.Picker{
		Items:        items,
		PreviousHref: representation.Links.Href("previous"),
		NextHref:     representation.Links.Href("next"),
		BackHref:     (routes.ChoreCreation{}).Href(),
		ManualHref: (routes.ChoreCreation{
			Source:     routes.ChoreCreationSourceManual,
			Recurrence: routes.ChoreCreationRecurrenceOneOff,
		}).Href(),
		Search:          search,
		AutofocusSearch: autofocusSearch,
	}
}

func detailViewModel(representation representation, notice string) viewmodels.Detail {
	return viewmodels.Detail{
		ChoreTemplate:  representation,
		CollectionHref: knownLink(representation.Links, "collection"),
		EditHref:       editNavigationHref(representation),
		DeactivateHref: deactivationConfirmationHref(representation),
		Notice:         notice,
	}
}

func editFormViewModel(choreTemplate *models.ChoreTemplate) viewmodels.EditForm {
	selfHref := newResponse(choreTemplate).Links.Href("self")
	return viewmodels.EditForm{
		ActionHref: selfHref,
		CancelHref: selfHref,
		Name: viewmodels.Field{
			Value: choreTemplate.Name,
		},
		Description: viewmodels.Field{
			Value: choreTemplate.Description,
		},
	}
}

func editFormFeedbackViewModel(
	id int,
	request editRequest,
	fieldErrors map[string]string,
	summaryError string,
) viewmodels.EditForm {
	selfHref := resourceHref(id)
	return viewmodels.EditForm{
		ActionHref: selfHref,
		CancelHref: selfHref,
		Name: viewmodels.Field{
			Value: request.Name,
			Error: fieldErrors["name"],
		},
		Description: viewmodels.Field{
			Value: request.Description,
			Error: fieldErrors["description"],
		},
		SummaryError: summaryError,
	}
}

func errorViewModel(response errorResponse) viewmodels.Error {
	model := viewmodels.Error{Message: response.Error}
	if href := knownLink(response.Links, "self"); href != "" {
		model.Links = append(model.Links, viewmodels.Link{
			Label: "Back to chore template",
			Href:  href,
		})
	}
	if href := knownLink(response.Links, "collection"); href != "" {
		model.Links = append(model.Links, viewmodels.Link{
			Label: "Back to chore templates",
			Href:  href,
		})
	}
	return model
}

func knownLink(links api.Relations, relation string) string {
	return links.Href(relation)
}

func editNavigationHref(representation representation) string {
	if href := representation.Actions.Href("edit"); href != "" {
		return href + "/edit"
	}
	return ""
}

func deactivationConfirmationHref(representation representation) string {
	if href := representation.Actions.Href("deactivate"); href != "" {
		return href
	}
	return ""
}
