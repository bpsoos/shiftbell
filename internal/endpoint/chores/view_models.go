package chores

import (
	"strconv"
	"time"

	api "github.com/bpsoos/shiftbell/internal/models/api"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
)

func collectionViewModel(
	collection choreCollectionResponse,
	status choremodels.ChoreStatus,
	search string,
	searchOpen bool,
	autofocusSearch bool,
) choreviewmodels.Collection {
	if status == "" {
		status = choremodels.ChoreStatusActive
	}
	items := make([]choreviewmodels.CollectionItem, len(collection.Items))
	for i, chore := range collection.Items {
		items[i] = choreviewmodels.CollectionItem{
			Chore:        chore,
			CompleteHref: completionHref(chore),
		}
	}
	return choreviewmodels.Collection{
		Items:           items,
		Links:           collection.Links,
		Actions:         collection.Actions,
		Status:          status,
		Search:          search,
		SearchOpen:      searchOpen,
		AutofocusSearch: autofocusSearch,
	}
}

func completionDialogViewModel(
	chore choreResponse,
) choreviewmodels.CompletionDialog {
	return choreviewmodels.CompletionDialog{
		Name:       chore.Name,
		ActionHref: completionHref(chore),
	}
}

func detailViewModel(
	chore choreRepresentation,
	notice string,
) choreviewmodels.Detail {
	return choreviewmodels.Detail{
		Chore:      chore,
		BackHref:   linkHref(chore.Links, "collection"),
		EditHref:   editFormHref(chore),
		DeleteHref: deletionConfirmationHref(chore),
		Notice:     notice,
	}
}

func editFormViewModel(chore *choremodels.ChoreDetails) choreviewmodels.EditForm {
	response := newChoreResponse(chore)
	return choreviewmodels.EditForm{
		ActionHref: response.Links.Href("self"),
		CancelHref: response.Links.Href("self"),
		Scheduled:  chore.ScheduleId != nil,
		Name: choreviewmodels.Field{
			Value: chore.Name,
		},
		Description: choreviewmodels.Field{
			Value: chore.Description,
		},
		Deadline: choreviewmodels.Field{
			Value: chore.Deadline.Format(time.DateOnly),
		},
	}
}

func editFormErrorViewModel(
	chore *choremodels.ChoreDetails,
	feedback formFeedback,
) choreviewmodels.EditForm {
	return choreviewmodels.EditForm{
		ActionHref:              feedback.Action.Href,
		CancelHref:              feedback.CancelHref,
		Scheduled:               chore.ScheduleId != nil,
		Name:                    fieldViewModel(feedback, "name"),
		Description:             fieldViewModel(feedback, "description"),
		Deadline:                fieldViewModel(feedback, "deadline"),
		AlsoUpdateChoreTemplate: feedback.Values["also_update_chore_template"] == "true",
		SummaryError:            formFeedbackMessage(feedback),
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
		ActionHref:     feedback.Action.Href,
		CancelHref:     feedback.CancelHref,
		SummaryError:   formFeedbackMessage(feedback),
		Submitted:      true,
		Name:           fieldViewModel(feedback, "name"),
		Description:    fieldViewModel(feedback, "description"),
		Deadline:       fieldViewModel(feedback, "deadline"),
		SaveAsTemplate: feedback.Values["save_as_chore_template"] == "true",
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

func completionHref(chore choreResponse) string {
	if chore.Status != choremodels.ChoreStatusActive {
		return ""
	}
	selfHref := linkHref(chore.Links, "self")
	if selfHref == "" {
		return ""
	}
	return selfHref + "/completion"
}

func editFormHref(chore choreRepresentation) string {
	if href := chore.Actions.Href("edit"); href != "" {
		return href + "/edit"
	}
	return ""
}

func deletionConfirmationHref(chore choreRepresentation) string {
	if href := chore.Actions.Href("delete"); href != "" {
		return href + "/deletion"
	}
	return ""
}

func formFeedbackMessage(feedback formFeedback) string {
	if feedback.Error.Error != "" {
		return feedback.Error.Error
	}
	return "Please correct the highlighted fields and try again."
}
