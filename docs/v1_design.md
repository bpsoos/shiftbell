# v1 design

## goals

Provide a simple CRUD-like interface for the managed resources both on browser client and the db side.

Simlicity can be achieved for functionality that requries a complex flow using async patterns.
For example when a scheduled chore is marked complete a new one should be created automatically with the proper deadline based on the schedule. The mark complete should not take on all of this complexity. A worker goroutine should notice that a new instance has to be created and perform a simple create operation.

## non-goals

user management, "topic" management, permissions
v2 will add users who can own chores and smart schedules that can rotate who owns the next chore
v3 will add high level "topics" that enables to create completely separated sets of chores and schedules for example "household" and "photgraphy side project"
v4 will add permissions that will allow the creation of read-only users

## stack

single binary golang application with embedded sqlite, and a simple htmx + bootstrap based UI

## resources

### chores
  - id,name,description,deadline,is_complete,completed_at

### choretypes
  - id,name,description

### schedules
  - id,chore_type_id,interval_days

## user flows

### one-off chore creation with manual input

1. user needs new one-off chore
    1. flow:
        1. users clicks "new chore"
        2. user is navigated to the create page
        3 user selects "manual" input type
        4. user enters parameters: name, description, deadline(optional)
        5. user clicks "create"
        6. user is navigated back to the chores page
        7. new one-off chore is displayed
2. user needs chore type based on one-off chore during creation
    1. flow:
        1. user completes the manual creation flow
        2. before clicking "create", ticks the "save as chore type"

### one-off chore creation based on chore type

1. user needs new one-off chore
    1. flow:
        1. users clicks "new chore"
        2. user is navigated to the create page
        3. user selects "chore type based" input type
        4. "chore type search and select" modal opens
        5. user selects chore type
        6. "chore type search and select" modal closes
        7. user enters parameters: name, description, deadline(optional)
        8. user clicks "create"
        9. user is navigated back to the chores page
        10. new one-off chore is displayed

### one-off chore update

1. user needs to update one-off chore
    1. flow:
        1. user navigates to the chore on the chores page
        2. user clicks "edit"
        3. user enter new params
            1. name, description, deadline changes allowed for all one-off chores
            2. completion date changtes allowed for finished one-off chores
        4. user clicks "update"
        5. the updated chore is displayed

### one-off chore completion

1. user finished one-off chore
    1. flow:
        1. user navigates to the chore on the chores page
        2. user clicks "mark complete"
        3. chore is pushed down and greyed out

### scheduled chore creation

1. user needs new scheduled chore
    1. flow:
        1. users clicks "new chore"
        2. user is navigated to the create page
        3 user selects either input type
        4. user enters parameters: name, description, deadline(optional)
        5. user fills the "repeat on an interval" input (number of days)
        6. user clicks "create"
        7. user is navigated back to the chores page
        8. the first instance of the scheduled chore is eventually displayed
        9. [background]: a chore type is created if input type was manual
        10. [background]: a schedule is created configured with the given interval days and chore type
        11. [background]: a chore instance is created with a deadline calculated by the interval days of the schedule, name and description based on the chore type

### scheduled chore update

1. user needs to update a scheduled chore
    1. flow:
        1. user navigates to the completed chore on the chores page
        2. user clicks "edit"
        3. user enter new params
            1. name, description, deadline changes allowed for all scheduled chores
            2. interval days changes allowed on unfinished chores
            3. user can tick to also update the related chore type (default on)
            4. completion date changes allowed for finished scheduled chores
        4. user clicks "update"
        5. the updated chore is displayed

### schedule update

1. user needs to update the interval days of a schedule
    1. flow:
        1. user navigates to the schedules page
        2. user clicks "edit"
        3. user enter new params
            1. interval days changes allowed
            2. chore_type_id reference change disallowed
        4. user clicks "update"
        5. the updated schedule is displayed
        6. the next chore created based on this schedule will have a deadline calculated by the new interval days
