package chores

import (
	"strconv"

	api "github.com/bpsoos/shiftbell/internal/models/api"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
)

const choreCollectionHref = "/chores"

func collectionViewModel(
	collection choreCollectionResponse,
) choreviewmodels.Collection {
	return choreviewmodels.Collection{
		Collection: collection,
	}
}

func detailViewModel(
	chore choreRepresentation,
) choreviewmodels.Detail {
	return choreviewmodels.Detail{
		Chore:    chore,
		BackHref: linkHref(chore.Links, "collection"),
	}
}

func creationViewModel(
	creation choreCreationResponse,
) choreviewmodels.Creation {
	model := choreviewmodels.Creation{
		Creation: creation,
		BackHref: choreCollectionHref,
	}
	for _, choice := range creation.Choices {
		switch choice.Label {
		case "Specify new":
			model.SpecifyNewHref = choice.Href
		case "Select template":
			model.SelectTemplateHref = choice.Href
		case "One-off":
			model.OneOffHref = choice.Href
		case "Scheduled":
			model.ScheduledHref = choice.Href
		}
	}
	return model
}

func manualOneOffFormViewModel(
	creation choreCreationResponse,
) choreviewmodels.ManualOneOffForm {
	return choreviewmodels.ManualOneOffForm{
		ActionHref: creationActionHref(creation),
		CancelHref: choreCollectionHref,
	}
}

func templateOneOffFormViewModel(
	creation choreCreationResponse,
) choreviewmodels.TemplateOneOffForm {
	model := choreviewmodels.TemplateOneOffForm{
		ActionHref: creationActionHref(creation),
		CancelHref: choreCollectionHref,
	}
	if creation.Template != nil {
		model.ChoreTemplateId = creation.Template.Id
		model.ChoreTemplateName = creation.Template.Name
		model.ChoreTemplateDescription = creation.Template.Description
	}
	return model
}

func manualOneOffFormErrorViewModel(
	feedback formFeedback,
) choreviewmodels.ManualOneOffForm {
	return choreviewmodels.ManualOneOffForm{
		ActionHref:   feedback.Action.Href,
		CancelHref:   feedback.CancelHref,
		SummaryError: formFeedbackMessage(feedback),
		Submitted:    true,
		Name:         fieldViewModel(feedback, "name"),
		Description:  fieldViewModel(feedback, "description"),
		Deadline:     fieldViewModel(feedback, "deadline"),
	}
}

func templateOneOffFormErrorViewModel(
	feedback formFeedback,
) choreviewmodels.TemplateOneOffForm {
	templateId, err := strconv.Atoi(feedback.Values["chore_template_id"])
	if err != nil {
		templateId = 0
	}
	return choreviewmodels.TemplateOneOffForm{
		ActionHref:      feedback.Action.Href,
		CancelHref:      feedback.CancelHref,
		SummaryError:    formFeedbackMessage(feedback),
		Submitted:       true,
		ChoreTemplateId: templateId,
		Deadline:        fieldViewModel(feedback, "deadline"),
	}
}

func errorViewModel(response api.ErrorResponse) choreviewmodels.Error {
	model := choreviewmodels.Error{Message: response.Error}
	for _, relation := range []struct {
		name  string
		label string
	}{
		{name: "self", label: "Back to chore"},
		{name: "collection", label: "Back to chores"},
	} {
		if href := linkHref(response.Links, relation.name); href != "" {
			model.Links = append(model.Links, choreviewmodels.Link{
				Label: relation.label,
				Href:  href,
			})
		}
	}
	return model
}

func fieldViewModel(feedback formFeedback, name string) choreviewmodels.Field {
	return choreviewmodels.Field{
		Value: feedback.Values[name],
		Error: feedback.FieldErrors[name],
	}
}

func creationActionHref(creation choreCreationResponse) string {
	return creation.Actions.Href("create")
}

func linkHref(links api.Relations, name string) string {
	return links.Href(name)
}

func formFeedbackMessage(feedback formFeedback) string {
	if feedback.Error.Error != "" {
		return feedback.Error.Error
	}
	return "Please correct the highlighted fields and try again."
}
