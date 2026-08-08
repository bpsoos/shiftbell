package choretemplates

import (
	api "github.com/bpsoos/shiftbell/internal/models/api"
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

func detailViewModel(representation representation) viewmodels.Detail {
	return viewmodels.Detail{
		ChoreTemplate:  representation,
		CollectionHref: knownLink(representation.Links, "collection"),
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

func knownLink(links map[string]api.Link, relation string) string {
	if link, ok := links[relation]; ok {
		return link.Href
	}
	return ""
}
