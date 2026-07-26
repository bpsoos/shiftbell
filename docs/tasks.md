# v1 tasks

## phase 1 - initial reconciliation

1. DONE rename chore types to chore templates in the endpoint and view layers
2. DONE rename chore types to chore templates in models, routing, and application wiring
3. DONE rename chore types to chore templates in persistence APIs, fixtures, and tests
4. DONE add missing chore names and align completion terminology with active, completed, and `completed_on`
5. DONE fix nullable-description reads, completion-state loading, and active ordering by deadline and ID
6. DONE load and validate `APP_TIMEZONE` with a UTC default
7. DONE cover unset, valid, and invalid `APP_TIMEZONE` values with Ginkgo specs

## phase 2 - service layer

1. add chore template creation and editing
2. add chore template browsing and details
3. add one-off chore creation
4. add scheduled chore creation
5. add active and completed chore browsing and details
6. add active chore editing
7. add chore completion and scheduled successors
8. add completed chore date correction
9. add chore deletion
10. add schedule browsing and details
11. add schedule editing
12. add chore template deactivation
13. add schedule deactivation

## phase 3 - persistence layer

1. DONE run database migrations on startup
2. replace the provisional schema with the v1 schema
3. extend chore template creation, browsing, details, and editing persistence
4. add schedule browsing, details, and editing persistence
5. extend chore browsing, details, editing, and deletion persistence
6. add atomic chore creation persistence
7. extend atomic completion persistence with successors and corrections
8. add atomic template and schedule deactivation persistence

## phase 4 - endpoint and view layer

1. route handlers through services and map domain errors to HTTP and validation feedback
2. add schedule routes and navigation plus active chore-template selectors
3. build active and completed chore lists with filter-preserving pagination
4. build chore details and active editing with optional template updates and a shared-template warning
5. build the four-variant chore creation flow and template-selection modal
6. direct duplicate-template creation to the existing active template
7. add completion confirmation with an application-local date default
8. add completion-date correction with all other fields read-only
9. add permanent chore deletion flows
10. keep scheduled deadlines and schedule template links read-only
11. build paginated template search, filters, selection, detail, create, and edit flows
12. show active and deactivated schedule counts on chore-template details
13. add template deactivation and link to schedules that block it
14. build schedule search, filters, details, editing, and permanent-deactivation flows
15. link schedule views to their template and current active chore and show the current deadline
16. exclude deactivated resources from normal lists and show read-only deactivated filters
17. label irreversible actions `Deactivate permanently` and require confirmation
18. display the resolved application timezone and use it for UI date defaults
19. cover handlers and rendered user flows with Ginkgo specs
