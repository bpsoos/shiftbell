# v1 design

## Goals

Simple CRUD-like interface for chores, chore templates, and schedules.

Complex flows with multiple steps complete synchronously in a single transaction. For example, completing a scheduled chore and creating its successor either both succeed or neither change is persisted.

## Non-goals

- User management and chore ownership
- Topics that separate independent sets of chores and schedules
- Permissions and read-only access


V2 would add users who can own chores and schedules that can rotate ownership.

V3 would add topics such as household and photography.

V4 would add permissions and read-only users.

## Stack

A single-binary go application with embedded sqlite and an htmx and bootstrap-based ui.

## Application time

Deadlines and completion dates are calendar dates rather than timestamps. Date calculations use the configured application timezone.

The application timezone is configured with `APP_TIMEZONE` using an IANA timezone name such as `Europe/Budapest`. It defaults to `UTC` when unset and prevents startup when explicitly configured with an invalid value. The resolved timezone is displayed in the UI.

Changing the configured timezone only affects subsequent calculations of dates.

## Resources

### Chores

- `id`
- `schedule_id`, nullable
- `name`, required and limited to 200 characters
- `description`, nullable and limited to 2,000 characters
- `deadline`, required calendar date
- `completed_on`, nullable calendar date

A chore is active when `completed_on` is null and completed otherwise.


A one-off chore has no schedule.

A scheduled chore references exactly one schedule.


Chores store name and description snapshots. One-off chores created from a chore template do not retain a `chore_template_id`. Scheduled chores reach their chore template through their schedule.


### Chore templates

- `id`
- `name`, required and limited to 200 characters
- `description`, nullable and limited to 2,000 characters
- `deactivated_at`, nullable timestamp

Names are unique case-insensitively among active chore templates. A new active chore template may reuse the name of a permanently deactivated chore template.


### Schedules

- `id`
- `name`, required and case insensitively unique among active schedules
- `chore_template_id`, required and immutable
- `interval_days`, required
- `deactivated_at`, nullable timestamp

A schedule is active when `deactivated_at` is null and deactivated otherwise. Deactivation is permanent.

One chore template may have multiple active schedules. A schedule may have many historical chores but at most one active chore.


## Validation

- Names are trimmed and must not be empty.
- Descriptions are trimmed and stored as null when empty.
- Deadlines are required, default to the application-local current date, and may be in the past.
- Completion dates may be in the past or present but never in the future.
- `interval_days` must be a whole number from 1 through 3,650 and has no default.
- Chore-template names are unique case-insensitively among active chore templates.
- Schedule names are unique among active schedules and default from the selected chore template or manually entered chore name.
- Search input is trimmed and limited to 200 characters.

Names, descriptions and search inputs are allowlisted.

Names allow only characters accepted by Go's `unicode.IsPrint`.

```go
func normalizeName(value string) (string, bool) {
	if !utf8.ValidString(value) {
		return "", false
	}

	value = norm.NFC.String(strings.TrimSpace(value))
	if value == "" || utf8.RuneCountInString(value) > 200 {
		return "", false
	}

	for _, r := range value {
		if !unicode.IsPrint(r) {
			return "", false
		}
	}

	return value, true
}
```

Descriptions allow characters accepted by `unicode.IsPrint` plus tabs and line breaks. Line endings are normalized to `\n`. Leading and trailing whitespace is trimmed while internal whitespace and formatting are preserved.

```go
func normalizeDescription(value string) (string, bool) {
	if !utf8.ValidString(value) {
		return "", false
	}

	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = norm.NFC.String(value)
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > 2000 {
		return "", false
	}

	for _, r := range value {
		if !unicode.IsPrint(r) && r != '\t' && r != '\n' {
			return "", false
		}
	}

	return value, true
}
```

Search inputs allow only characters accepted by `unicode.IsPrint`. An empty normalized search is valid.

```go
func normalizeSearch(value string) (string, bool) {
	if !utf8.ValidString(value) {
		return "", false
	}

	value = norm.NFC.String(strings.TrimSpace(value))
	if utf8.RuneCountInString(value) > 200 {
		return "", false
	}

	for _, r := range value {
		if !unicode.IsPrint(r) {
			return "", false
		}
	}

	return value, true
}
```


## List behavior

Active chores are displayed by default. Users may switch between `active` and `completed`; v1 does not provide a combined `all` filter.

Active chores are ordered by deadline ascending and then ID ascending. Completed chores are ordered by completion date descending.

Completed chores are removed from the default active list immediately after completion.

Schedules and chore templates are reverse ordered by id and searchable.

Deactivated schedules and chore templates are excluded from normal lists and selectors. They remain available through a read-only `deactivated` filter.

## Recurrence behavior

The first deadline of a schedule is required and defaults to the application-local current date.

Completing a scheduled chore creates exactly one successor with:

`deadline = completed_on + interval_days`

The calculation intentionally uses the actual completion date rather than the previous deadline. Repeated completion requests are successful no-ops and never create duplicate successors.

Each schedule has at most one active chore. Completing a scheduled chore and creating its successor occur in the same transaction, so a successful completion immediately leaves the schedule with its next active chore.

Correcting the completion date of a scheduled chore recalculates the deadline of its immediate successor in the same transaction while that successor is still active. Corrections do not cascade once the successor has been completed.

Changing `interval_days` does not modify the current active chore. The new interval applies when calculating the next successor. Scheduled chore deadlines cannot be edited manually after creation. An active scheduled chore's deadline changes only when correction of its predecessor's completion date causes the application to recalculate it.


## Transactional behavior

Operations that change multiple resources are synchronous and atomic:

- Manual scheduled creation creates the chore template, schedule, and first chore together.
- Chore-template-based scheduled creation creates the schedule and first chore together.
- Scheduled completion records the completion and creates exactly one successor together.
- Completion-date correction updates the completion date and active successor deadline together.
- Permanent schedule deactivation records the deactivation and removes the active chore together.

If any step fails, the request returns an error and none of its changes are persisted. Repeated completion requests remain successful no-ops and do not create duplicate successors.


## Deletion and permanent deactivation

One-off chores may be permanently deleted. Completed scheduled chores may be permanently deleted. A successful scheduled completion guarantees that its successor already exists.

An active scheduled chore cannot be deleted directly. The user must permanently deactivate its schedule.

Permanently deactivating a schedule:

- Sets `deactivated_at`
- Removes its current active chore
- Preserves completed chores linked to the schedule
- Prevents reactivation and future successor creation

A chore template is permanently deactivated rather than hard-deleted. Deactivation is rejected while an active schedule references it. Deactivated chore templates disappear from creation choices but preserve historical references.

The UI labels these actions `Deactivate permanently` and requires confirmation.

## User flows

### Chore creation

1. The user clicks `New chore` and navigates to the creation flow.
2. Under `Source`, the user chooses `Specify new` or `Select from template`.
3. Selecting a template opens the chore-template search and selection modal. After selection, the template name and description are displayed read-only.
4. Under `Recurrence`, the user chooses `One-off` or `Scheduled`.
5. The final form displays the fields required by the selected source and recurrence.
6. The user clicks `Create` and returns to the chores page.

The final form has four variants:

- `Specify new` and `One-off`:
    - required name
    - optional description
    - required deadline defaulted to today.
    - `Save as chore template` is optional.
- `Select from template` and `One-off`:
    - selected template name (read-only)
    - selected template description (read-only)
    - required deadline defaults to today.
    - The resulting chore is an independent snapshot and does not retain a link to the template.
- `Specify new` and `Scheduled`:
    - required name
    - optional description
    - required schedule name defaulted from the chore name
    - required first deadline defaulted to today
    - required `interval_days`.
    - A chore template is created automatically because every schedule requires one.
- `Select from template` and `Scheduled`:
    - selected template name (read-only)
    - selected template description (read-only)
    - required schedule name defaults from the template name
    - required first deadline defaults to today.
    - `interval_days` is required.

The backend treats a non-empty `interval_days` as the scheduling request. The recurrence choice controls whether that field is presented and submitted.

One-off chore creation persists the new chore immediately. When `Save as chore template` is selected, the chore and template are created atomically. Scheduled creation creates the template when needed, the schedule, and the first active chore atomically. A duplicate active chore-template name produces validation feedback and directs the user to select the existing template. A failed request creates no partial records.

### Active chore update

1. The user opens an active chore and clicks `Edit`.
2. The user may update the name and description.
3. For a one-off chore, the user may also update the deadline.
4. For a scheduled chore, the deadline remains read-only and the user may explicitly choose `Also update chore template`, which defaults off.
5. The UI warns that updating a shared chore template affects future instances of every schedule using it.
6. The user clicks `Update` and sees the updated chore.

Recurrence settings are not edited from the chore form. They belong to the schedule page.

### Chore completion

1. The user clicks `Mark complete` on an active chore.
2. A confirmation modal pops up asking the user to confirm the completion day (defaulted to application-local "today").
3. The user confirms the completion day.
3. The chore disappears from the default active list and becomes available through the completed filter.
4. Completion is irreversible, but the completion date may be corrected.

For a scheduled chore, the same successful request creates exactly one successor and the schedule immediately links to it.

### Completed chore correction

1. The user opens a chore from the completed list.
2. The user clicks `Correct completion date`.
3. The name, description, and deadline remain read-only.
4. The user enters a completion date that is not in the future.
5. The user clicks `Update` and sees the corrected completed chore.
6. For a scheduled chore, the active successor deadline is recalculated in the same transaction.

### Chore deletion

1. The user opens a deletable chore and clicks `Delete permanently`.
2. The UI asks for confirmation.
3. The user confirms and returns to the chores page.
4. The deleted chore no longer appears in active or completed lists.

One-off chores and completed scheduled chores are deletable. Active scheduled chores must be removed by permanently deactivating their schedule.

### Chore-template browsing and search

1. The user navigates to the chore-template page.
2. The page displays active templates reverse ordered by id and paginated.
3. The user may enter a case-insensitive substring search across name and description.
4. An empty search returns the normal active list.
5. The user may switch to the read-only deactivated filter.
6. The user opens a template to view its details and linked active and deactivated schedule counts.

### Chore-template creation and update

1. The user navigates to the chore-template page.
2. The user clicks `New chore template` or opens an active template and clicks `Edit`.
3. The user enters a required name and optional description.
4. The user clicks `Create` or `Update` and sees the persisted template.

A duplicate active chore-template name produces validation feedback. Chore-template edits affect only future scheduled instances.

### Chore-template permanent deactivation

1. The user opens an active chore template and clicks `Deactivate permanently`.
2. If an active schedule references the template, deactivation is rejected and the referencing schedules are shown.
3. Otherwise, the UI asks for confirmation.
4. The user confirms and returns to the chore-template page.
5. The template disappears from active lists and creation selectors and remains available through the deactivated filter.

### Schedule browsing

1. The user navigates to the schedules page.
2. The page displays each schedule's template name with a link, schedule name, interval, and active or deactivated state.
3. An active schedule displays its current deadline and a link to its active chore.
4. The user may switch to the read-only deactivated filter.
5. The user opens a schedule to view its details.

### Schedule update

1. The user opens an active schedule and clicks `Edit`.
2. The user may update the schedule name and `interval_days`.
3. The linked chore template remains read-only and cannot be replaced.
4. The user clicks `Update` and sees the updated schedule.

An interval change does not modify the current active chore. It applies when the next successor is created.

### Schedule permanent deactivation

1. The user opens an active schedule and clicks `Deactivate permanently`.
2. The UI asks for confirmation and explains that the action cannot be reversed.
3. The user confirms and returns to the schedules page.
4. The schedule and its active chore disappear from active lists.
5. The schedule remains available through the deactivated filter with its completed history preserved.

Deactivation records the schedule deactivation and removes its active chore in the same transaction.
