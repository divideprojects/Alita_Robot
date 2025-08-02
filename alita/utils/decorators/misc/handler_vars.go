package misc

import "sync"

var (
	AdminCmds   = make([]string, 0)
	UserCmds    = make([]string, 0)
	DisableCmds = make([]string, 0)
	mu          = &sync.Mutex{}
)

// addToArray is a func to add to array of strings
func addToArray(arr []string, val ...string) []string {
	mu.Lock()
	defer mu.Unlock()
	return append(arr, val...)
}

/*
AddAdminCmd adds one or more command names to the list of admin commands.

This function is thread-safe and can be called concurrently.
*/
func AddAdminCmd(cmd ...string) {
	AdminCmds = addToArray(AdminCmds, cmd...)
}

/*
AddUserCmd adds one or more command names to the list of user commands.

This function is thread-safe and can be called concurrently.
*/
func AddUserCmd(cmd ...string) {
	UserCmds = addToArray(UserCmds, cmd...)
}

/*
AddCmdToDisableable adds a command name to the list of disableable commands.

This function is thread-safe and can be called concurrently.
*/
func AddCmdToDisableable(cmd string) {
	DisableCmds = addToArray(DisableCmds, cmd)
}
