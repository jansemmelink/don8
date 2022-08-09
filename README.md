# Don8

Application to manage donations to an organisation, for example donations from parents to the school for a fund raising event.

Written with Affies Wildsfees in mind...

# Status
## Progress
* Auth is kind of working: can register, get email, can activate and can login. Need to test a bit more...
* Can create groups and child groups and navigate between then
* Can list own groups and open from the list

## Next
* Dev: Create more users and switch between users in my dev app
* Disable [logout] nav link when user opted for random password - as user won't be able to login again? Or check that can login again with cached password... switch on auto complete in the login form...
* Generate a link to invite others to a group, send in email and prompt to join the group
* Send link to an email list
* Attach named email lists to a group (e.g. to Affies) and update the list from time to time.
* Dynamic list of those who joined and those who did not join...
* Create group requests
* Promise something
* Receive something
* Show request status
* Generate a report to send to all members or send as PDF from school to mailing list

* Fix Group title/desc updates
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
