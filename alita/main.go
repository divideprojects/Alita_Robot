package alita

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/modules"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

// ResourceMonitor monitors system resources
func ResourceMonitor() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			numGoroutines := runtime.NumGoroutine()

			// Log metrics
			log.WithFields(log.Fields{
				"goroutines": numGoroutines,
				"memory_mb":  m.Alloc / 1024 / 1024,
				"sys_mb":     m.Sys / 1024 / 1024,
				"gc_runs":    m.NumGC,
			}).Info("Resource usage stats")

			// Warning thresholds
			if numGoroutines > 1000 {
				log.WithField("goroutines", numGoroutines).Warn("High goroutine count detected")
			}

			if m.Alloc/1024/1024 > 500 { // 500MB
				log.WithField("memory_mb", m.Alloc/1024/1024).Warn("High memory usage detected")
			}
		}
	}
}

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

	// Start resource monitoring
	go ResourceMonitor()
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
