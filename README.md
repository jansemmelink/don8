# Don8

Application to manage donations to an organisation, for example donations from parents to the school for a fund raising event.

Written with Affies Wildsfees in mind...

# Status
## Progress
* Authing the user on certain API e.g. to create group else unauthorized + error message
* Created API to create group with user as member with all permissions

## Next
* App: Coordinator: Create groups and list my groups
* App: Coordinator: Create child groups and list my child groups
* App: Filter group list by text
* App: Open Group - show nr of followers (later more) + sub-groups
* App: Group Invite Link for other to join
* App: Open group invite and join
* App/Api: Leave a group or delete a group (if you're the only admin)
* App/Api: Invite others to a group as coordinators (search by phone nr)

* User check when app start
* Still need to figure out authentication - ideally by SMS to the user, or may be start with email which is free... just need to confirm we can contact the user.
* Call Log() on each db function that changes the db
* Write more db/*_test.go modules to test these modules
* Create an API
* Start the React App with very simple screens and strict access control...
    * Start with user side to register then find groups from link or search,
    * Let user Like/follow group and show in home screen
    * Make a promise
    * Let member receive a promise
    * See member reports
    * Let user manage own promises (list promised|donated|all, and option to withdraw a promise or adjust a promise)
    * Notify members of promises and donations
    * Let members move donations to other locations
