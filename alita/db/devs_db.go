package db

import (
	"errors"
	"fmt"
	"runtime"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/dustin/go-humanize"
)

func GetTeamMemInfo(userID int64) (devrc *DevSettings) {
	devrc = &DevSettings{}
	err := GetRecord(devrc, DevSettings{UserId: userID})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		devrc = &DevSettings{UserId: userID, IsDev: false}
	} else if err != nil {
		devrc = &DevSettings{UserId: userID, IsDev: false}
		log.Errorf("[Database] GetTeamMemInfo: %v - %d", err, userID)
	}
	log.Infof("[Database] GetTeamMemInfo: %d", userID)
	return
}

func GetTeamMembers() map[int64]string {
	var teamArray []*DevSettings
	array := make(map[int64]string)

	err := GetRecords(&teamArray, DevSettings{IsDev: true})
	if err != nil {
		log.Error(err)
		return nil
	}

	for _, result := range teamArray {
		if result.IsDev {
			array[result.UserId] = "dev"
		}
	}

	return array
}

func AddDev(userID int64) {
	devSettings := &DevSettings{UserId: userID, IsDev: true}

	// Try to update existing record first
	err := UpdateRecord(&DevSettings{}, DevSettings{UserId: userID}, DevSettings{IsDev: true})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new record if not exists
		err = CreateRecord(devSettings)
	}

	if err != nil {
		log.Errorf("[Database] AddDev: %v - %d", err, userID)
		return
	}
	log.Infof("[Database] AddDev: %d", userID)
}

func RemDev(userID int64) {
	err := DB.Where("user_id = ?", userID).Delete(&DevSettings{}).Error
	if err != nil {
		log.Errorf("[Database] RemDev: %v - %d", err, userID)
	}
}

func LoadAllStats() string {
	totalUsers := LoadUsersStats()
	activeChats, inactiveChats := LoadChatStats()
	AcCount, ClCount := LoadPinStats()
	uRCount, gRCount := LoadReportStats()
	antiCount := LoadAntifloodStats()
	setRules, pvtRules := LoadRulesStats()
	blacklistTriggers, blacklistChats := LoadBlacklistsStats()
	connectedUsers, connectedChats := LoadConnectionStats()
	disabledCmds, disableEnabledChats := LoadDisableStats()
	filtersNum, filtersChats := LoadFilterStats()
	enabledWelcome, enabledGoodbye, cleanServiceEnabled, cleanWelcomeEnabled, cleanGoodbyeEnabled := LoadGreetingsStats()
	notesNum, notesChats := LoadNotesStats()
	numChannels := LoadChannelStats()

	result := "<u>Alita's Stats:</u>" +
		fmt.Sprintf("\n\nGo Version: %s", runtime.Version()) +
		fmt.Sprintf("\nGoroutines: %s", humanize.Comma(int64(runtime.NumGoroutine()))) +
		fmt.Sprintf("\n<b>Antiflood:</b> enabled in %s chats", humanize.Comma(antiCount)) +
		fmt.Sprintf(
			"\n<b>Users:</b> %s users found in %s active Chats (%s Inactive, %s Total)",
			humanize.Comma(totalUsers),
			humanize.Comma(int64(activeChats)),
			humanize.Comma(int64(inactiveChats)),
			humanize.Comma(int64(activeChats+inactiveChats)),
		) +
		"\n<b>Pins:</b>" +
		fmt.Sprintf("\n    <b>CleanLinked Enabled:</b> %s", humanize.Comma(ClCount)) +
		fmt.Sprintf("\n    <b>AntiChannelPin Enabled:</b> %s", humanize.Comma(AcCount)) +
		fmt.Sprintf(
			"\n<b>Reports:</b> %s users enabled reports in %s Chats",
			humanize.Comma(uRCount),
			humanize.Comma(gRCount),
		) +
		"\n<b>Rules:</b>" +
		fmt.Sprintf("\n    <b>Set:</b> %s", humanize.Comma(setRules)) +
		fmt.Sprintf("\n    <b>Private:</b> %s", humanize.Comma(pvtRules)) +
		fmt.Sprintf(
			"\n<b>Blacklists:</b> %s triggers in %s chats",
			humanize.Comma(blacklistTriggers),
			humanize.Comma(blacklistChats),
		) +
		"\n<b>Connections:</b>" +
		fmt.Sprintf("\n    %s users connected to chats", humanize.Comma(connectedUsers)) +
		fmt.Sprintf("\n    %s chats allow user connections", humanize.Comma(connectedChats)) +
		fmt.Sprintf(
			"\n<b>Disabling:</b> %s commands disabled in %s chats",
			humanize.Comma(disabledCmds),
			humanize.Comma(disableEnabledChats),
		) +
		fmt.Sprintf(
			"\n<b>Filters:</b> %s filters saved in %s chats",
			humanize.Comma(filtersNum),
			humanize.Comma(filtersChats),
		) +
		"\n<b>Greetings:</b>" +
		fmt.Sprintf("\n    <b>Welcome Enabled:</b> %s", humanize.Comma(enabledWelcome)) +
		fmt.Sprintf("\n    <b>Goodbye Enabled:</b> %s", humanize.Comma(enabledGoodbye)) +
		fmt.Sprintf("\n    <b>CleanService:</b> %s", humanize.Comma(cleanServiceEnabled)) +
		fmt.Sprintf("\n    <b>CleanWelcome:</b> %s", humanize.Comma(cleanWelcomeEnabled)) +
		fmt.Sprintf("\n    <b>CleanGoodbye:</b> %s", humanize.Comma(cleanGoodbyeEnabled)) +
		fmt.Sprintf(
			"\n<b>Notes:</b> %s notes saved in %s chats",
			humanize.Comma(notesNum),
			humanize.Comma(notesChats),
		) +
		fmt.Sprintf("\n<b>Channels Stored</b>: %s", humanize.Comma(numChannels))

	return result
}

// AddSudo - Note: The new DevSettings model only supports Dev role, not Sudo
func AddSudo(userID int64) {
	log.Warnf("[Database] AddSudo: Sudo role not supported in new model, adding as Dev instead for user %d", userID)
	AddDev(userID)
}

// RemSudo - Note: The new DevSettings model only supports Dev role, not Sudo
func RemSudo(userID int64) {
	log.Warnf("[Database] RemSudo: Sudo role not supported in new model, removing Dev instead for user %d", userID)
	RemDev(userID)
}
