# v1 tasks

## phase 1 - initial reconciliation

1. DONE rename chore types to chore templates in the endpoint and view layers
2. DONE rename chore types to chore templates in models, routing, and application wiring
3. rename chore types to chore templates in persistence APIs, fixtures, and tests
4. add missing chore names and align completion terminology with active, completed, and `completed_on`
5. fix nullable-description reads, completion-state loading, and active ordering by deadline and ID
6. load and validate `APP_TIMEZONE` with a UTC default
7. cover unset, valid, and invalid `APP_TIMEZONE` values with Ginkgo specs

## phase 2 - service layer

1. add service inputs, results, repository ports, transaction ports, and domain errors
2. normalize and validate names, descriptions, searches, dates, and intervals
3. inject the application clock and timezone into date-dependent services
4. implement active and completed chore browsing and detail services
5. implement all four creation variants with atomic template, schedule, and first-chore writes
6. keep template-based one-offs unlinked and apply schedule-name and first-deadline defaults
7. implement active chore updates with immutable scheduled deadlines and optional template updates
8. complete chores idempotently and atomically create successors from the current template
9. calculate successor deadlines as `completed_on + interval_days`
10. implement completion correction and linked active-successor recalculation in one transaction
11. allow one-off and completed scheduled chore deletion and reject active scheduled deletion
12. implement template browsing, search, counts, creation, updates, and permanent deactivation
13. reject template deactivation with its referencing active schedules
14. implement schedule browsing, search, details, and name and interval updates
15. keep schedule templates immutable and leave active chores unchanged by interval updates
16. deactivate schedules atomically while removing active chores and preserving completed chores
17. cover validation, transaction orchestration, recurrence, and lifecycle rules with Ginkgo specs

## phase 3 - persistence layer

1. replace the provisional migrations with the exact v1 fields and relationships
2. remove direct chore-template links, `is_complete`, and `completed_at`
3. enforce canonical calendar dates, required deadlines, nullability, and interval bounds
4. add case-insensitive partial unique indexes for active template and schedule names
5. enforce at most one active chore per schedule
6. enforce at most one successor per predecessor
7. make every persister usable through a shared SQLite transaction
8. implement chore create, detail, ordered active/completed list, permitted edit, and delete queries
9. implement reverse-ID template search, filters, details, counts, writes, and deactivation
10. implement reverse-ID schedule search, filters, details, active-chore reads, writes, and deactivation
11. add predecessor links and conditional completion and correction primitives
12. cover schema constraints, query behavior, and name reuse with Ginkgo specs
13. integration-test creation, completion, correction, and deactivation transactions
14. run embedded migrations before serving and remove the migrate entrypoint

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
