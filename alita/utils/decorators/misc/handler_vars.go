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
	arr = append(arr, val...)
	mu.Unlock()
	return arr
}

// AddAdminCmd AddAdminCm: Add Connection Commands for Admin
func AddAdminCmd(cmd ...string) {
	AdminCmds = addToArray(AdminCmds, cmd...)
}

// AddUserCmd Add Connection Commands for User
func AddUserCmd(cmd ...string) {
	UserCmds = addToArray(UserCmds, cmd...)
}

// AddCmdToDisableable - func to add cmd to list of disableable
func AddCmdToDisableable(cmd string) {
	DisableCmds = addToArray(DisableCmds, cmd)
}
