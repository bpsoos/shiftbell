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

1. DONE add chore template creation and editing
2. DONE add chore template browsing and details
3. DONE add one-off chore creation
4. DONE add scheduled chore creation
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

1. display the application timezone and use local date defaults
2. build searchable chore template browsing and details
3. build chore template creation and editing
4. build searchable schedule browsing, details, and editing
5. build active and completed chore browsing and details
6. build four-variant chore creation and template selection
7. build active chore editing and optional template updates
8. build chore completion confirmation
9. build completed chore date correction
10. build confirmed permanent chore deletion
11. build confirmed chore template deactivation with blocking schedule links
12. build confirmed schedule deactivation
