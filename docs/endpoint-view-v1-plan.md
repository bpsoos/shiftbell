# Endpoint and view v1 plan

## Goal

Add a small server-rendered HTML surface without changing the established
`application/vnd.shiftbell+json` API. The browser UI covers navigation,
read-only browsing, pagination, and chore creation. Existing API mutations remain
available to vendor-JSON clients but are not exposed as HTML controls or forms.

## Compatibility boundary

- Preserve existing vendor-JSON fields, links, actions, status codes, and
  acceptance behavior.
- Keep the existing API routes registered.
- HTML must not expose search, collection filters, save-as-template,
  chore-template creation, chore editing, completion, completion correction,
  deletion, or chore-template deactivation.
- Mutation routes other than `POST /chores` are vendor-JSON-only. An HTML
  request to one of those routes is not an alternate mutation UI.
- Chore and chore-template detail pages are read-only even when their JSON
  representations advertise mutation actions.
- No schedule collection, detail, mutation, route, or navigation item is added.

## HTML surface

| Route | HTML behavior |
|---|---|
| `GET /` | Redirect to `/chores`. |
| `GET /chores` | Render the paginated active chore collection and its new-chore link. |
| `GET /chores/new` | Render source selection, recurrence selection, and manual or template-based creation forms. |
| `POST /chores` | Submit a one-off chore form, redisplay validation feedback, or redirect to `/chores` after success. |
| `GET /chores/:id` | Render a read-only chore detail. |
| `GET /chore-templates` | Render the paginated active chore-template collection. |
| `GET /chore-templates?picker=1` | Render the paginated active-template picker used by chore creation. |
| `GET /chore-templates/:id` | Render a read-only chore-template detail. |
| `GET/HEAD /assets/*` | Serve embedded local assets. |

The HTML collections use endpoint-provided next and previous links. They do not
render search boxes, status/state tabs, or other filter controls. The
vendor-JSON collection endpoints retain their existing search and filter
behavior.

The chore-template collection ignores its JSON create action when rendering
HTML. Detail views likewise ignore edit, complete, correct-completion, delete,
and deactivate actions.

## Chore creation

The browser flow supports:

1. Choosing `Specify new` or `Select template`.
2. Selecting an active template through the paginated picker when applicable.
3. Choosing `One-off` or `Scheduled`.
4. Creating a manual one-off chore with name, description, and deadline, or a
   template-based one-off chore with template ID and deadline.

The HTML form never renders or enables `save_as_chore_template`. That field and
its existing behavior remain part of the vendor-JSON API only.

Initial and failed creation forms use typed view models under
`internal/models/view/chores`. The endpoint owns request binding, validation
mapping, response status, redirects, and preservation of submitted form values.
The view owns only escaped HTML rendering.

### Scheduled recurrence

Scheduled recurrence remains visible as a creation choice but is not
implemented:

- `GET /chores/new?source=manual&recurrence=scheduled` returns
  `501 Not Implemented`.
- `GET /chores/new?template_id={id}&recurrence=scheduled` validates and loads
  the active template, then returns `501 Not Implemented`.
- `POST /chores` with a non-null `interval_days` returns
  `501 Not Implemented` before calling the chore service.

The response uses the negotiated JSON or HTML error representation. A failed
scheduled request must not create a chore, chore template, schedule, or partial
record.

## Endpoint and view boundary

Endpoints negotiate the response, parse input, call the core, construct the
canonical API representation, and map the small amount of HTML-only
presentation state. Views do not call services, persistence, Echo, or response
writers.

The view interfaces contain only methods needed by the supported HTML surface:
collection, read-only detail, creation steps, manual and template one-off forms,
picker, and error rendering. They do not contain mutation form or confirmation
methods.

Pagination links come from endpoint representations and are rendered verbatim.
Views do not reconstruct pagination, action eligibility, or domain rules.

## HTML and HTMX behavior

- Normal browser navigation renders the complete shell with one
  `<main id="main">`, basic chore and chore-template navigation, and local
  assets.
- `HX-Request: true` returns the replaceable main fragment.
- Navigation and pagination use HTMX while retaining ordinary `href`
  navigation.
- First-party JavaScript may enable swaps for server-rendered error responses,
  synchronize navigation, focus replaced content, and close mobile navigation.
  There is no modal lifecycle or mutation-specific browser behavior.
- The shell loads only embedded Bootstrap, HTMX, and ShiftBell assets.

## Configuration

Production wiring injects the resolved application timezone and `time.Now`
where required. Constructors assign the dependencies and configuration they are
given; they do not invent nil fallbacks or default clocks or timezones.

Date defaults and deadline presentation use the injected application timezone.
Incoming civil dates remain strict `YYYY-MM-DD` values.

## Testing

- Endpoint unit tests exercise public handler methods through Echo and
  `httptest`, with services and views mocked.
- Endpoint HTML tests stop at the view boundary and assert only endpoint-owned
  status, headers, model selection, redirects, and the component returned by the
  mock.
- Do not write unit tests for the view layer.
- Keep the existing vendor-JSON acceptance suite unchanged as the API
  compatibility gate.
- Do not remove JSON endpoint, service, persistence, or acceptance coverage for
  mutations merely because those mutations have no HTML UI.
- Cover all three scheduled-recurrence boundaries and assert that the service is
  not called for direct scheduled creation.

## Completion criteria

- The HTML shell, basic navigation, active collections, pagination, read-only
  details, picker, and one-off chore creation work for direct and HTMX requests.
- Scheduled creation returns `501` without side effects.
- No forbidden search, filter, save-as-template, chore-template creation, edit,
  completion, correction, deletion, or deactivation control or HTML mutation
  path remains.
- JSON-only mutation routes continue to satisfy their established vendor-JSON
  contract and acceptance tests.
- Generated Templ output is current, formatting and lint pass, tests pass, and
  the application builds.
