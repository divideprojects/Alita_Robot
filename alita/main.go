package alita

import (
	"fmt"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/modules"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

// ListModules list all modules loaded in the bot
func ListModules() string {
	modSlice := modules.HelpModule.AbleMap.LoadModules()
	sort.Strings(modSlice)
	return fmt.Sprintf("[%s]", strings.Join(modSlice, ", "))
}

// InitialChecks some initial checks before running bot
func InitialChecks(b *gotgbot.Bot) {
	// Create bot in db if not already created
	go db.EnsureBotInDb(b)
	checkDuplicateAliases()
	go cache.InitCache()
}

// check duplicate aliases of commands in the bot
func checkDuplicateAliases() {
	var althelp []string

	for _, i := range modules.HelpModule.AltHelpOptions {
		althelp = append(althelp, i...)
	}

	duplicateAlias, val := string_handling.IsDuplicateInStringSlice(althelp)
	if val {
		log.Fatalf("Found duplicate alias: %s", duplicateAlias)
	}
}

// LoadModules load the modules one-by-one using dispatcher
func LoadModules(dispatcher *ext.Dispatcher) {
	// Initialize Inner Map
	modules.HelpModule.AbleMap.Init()

	// Load this at last because it loads all the modules
	defer modules.LoadHelp(dispatcher)

	modules.LoadBotUpdates(dispatcher)
	modules.LoadAntispam(dispatcher)
	modules.LoadLanguage(dispatcher)
	modules.LoadAdmin(dispatcher)
	modules.LoadPin(dispatcher)
	modules.LoadMisc(dispatcher)
	modules.LoadBans(dispatcher)
	modules.LoadMutes(dispatcher)
	modules.LoadPurges(dispatcher)
	modules.LoadUsers(dispatcher)
	modules.LoadReports(dispatcher)
	modules.LoadDev(dispatcher)
	modules.LoadLocks(dispatcher)
	modules.LoadFilters(dispatcher)
	modules.LoadAntiflood(dispatcher)
	modules.LoadNotes(dispatcher)
	modules.LoadConnections(dispatcher)
	modules.LoadDisabling(dispatcher)
	modules.LoadRules(dispatcher)
	modules.LoadWarns(dispatcher)
	modules.LoadGreetings(dispatcher)
	modules.LoadBlacklists(dispatcher)
	modules.LoadMkdCmd(dispatcher)
}
