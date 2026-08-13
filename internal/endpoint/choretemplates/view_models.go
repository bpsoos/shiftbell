package choretemplates

import (
	api "github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
)

func collectionViewModel(representation collectionResponse) viewmodels.Collection {
	return viewmodels.Collection{Collection: representation}
}

func pickerViewModel(representation pickerCollectionResponse) viewmodels.Picker {
	return viewmodels.Picker{
		Collection: representation,
		BackHref:   "/chores/new",
		ManualHref: "/chores/new?source=manual",
	}
}

func detailViewModel(representation representation, notice string) viewmodels.Detail {
	return viewmodels.Detail{
		ChoreTemplate:  representation,
		CollectionHref: knownLink(representation.Links, "collection"),
		EditHref:       editNavigationHref(representation),
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
