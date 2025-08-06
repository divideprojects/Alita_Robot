package misc

import "sync"

var (
	AdminCmds   = make([]string, 0)
	UserCmds    = make([]string, 0)
	DisableCmds = make([]string, 0)
	mu          = &sync.Mutex{}
)

// addToArray safely appends multiple strings to a string slice using mutex protection.
// Thread-safe function for concurrent access to shared command arrays.
func addToArray(arr []string, val ...string) []string {
	mu.Lock()
	arr = append(arr, val...)
	mu.Unlock()
	return arr
}

// AddAdminCmd adds commands to the admin commands list for connection menu.
// These commands will be shown in the admin commands section of the connection interface.
func AddAdminCmd(cmd ...string) {
	AdminCmds = addToArray(AdminCmds, cmd...)
}

// AddUserCmd adds commands to the user commands list for connection menu.
// These commands will be shown in the user commands section of the connection interface.
func AddUserCmd(cmd ...string) {
	UserCmds = addToArray(UserCmds, cmd...)
}

// AddCmdToDisableable adds a command to the list of commands that can be disabled in chats.
// Administrators can use this to control which commands are available to regular users.
func AddCmdToDisableable(cmd string) {
	DisableCmds = addToArray(DisableCmds, cmd)
}
