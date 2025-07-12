package misc

import "sync"

var (
	AdminCmds   = make([]string, 0)
	UserCmds    = make([]string, 0)
	DisableCmds = make([]string, 0)
	mu          = &sync.RWMutex{}
)

// This function is no longer needed as we handle concurrency directly in the specific functions

/*
AddAdminCmd adds one or more command names to the list of admin commands.

This function is thread-safe and can be called concurrently.
*/
func AddAdminCmd(cmd ...string) {
	mu.Lock()
	defer mu.Unlock()
	AdminCmds = append(AdminCmds, cmd...)
}

/*
AddUserCmd adds one or more command names to the list of user commands.

This function is thread-safe and can be called concurrently.
*/
func AddUserCmd(cmd ...string) {
	mu.Lock()
	defer mu.Unlock()
	UserCmds = append(UserCmds, cmd...)
}

/*
AddCmdToDisableable adds a command name to the list of disableable commands.

This function is thread-safe and can be called concurrently.
*/
func AddCmdToDisableable(cmd string) {
	mu.Lock()
	defer mu.Unlock()
	DisableCmds = append(DisableCmds, cmd)
}

/*
GetAdminCmds returns a copy of the admin commands slice.

This function is thread-safe for reading.
*/
func GetAdminCmds() []string {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]string, len(AdminCmds))
	copy(result, AdminCmds)
	return result
}

/*
GetUserCmds returns a copy of the user commands slice.

This function is thread-safe for reading.
*/
func GetUserCmds() []string {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]string, len(UserCmds))
	copy(result, UserCmds)
	return result
}

/*
GetDisableCmds returns a copy of the disableable commands slice.

This function is thread-safe for reading.
*/
func GetDisableCmds() []string {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]string, len(DisableCmds))
	copy(result, DisableCmds)
	return result
}
