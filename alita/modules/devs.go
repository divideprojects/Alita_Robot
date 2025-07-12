package modules

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/extraction"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
)

/*
devsModule provides developer and admin commands for bot management.

Implements commands for team management, chat info, stats, and database cleanup.
*/
var devsModule = moduleStruct{
	moduleName: "Dev",
	cfg:        nil, // will be set during LoadDev
}

// for general purposes for strings in functions below
var txt string

/*
chatInfo retrieves information about a specified chat.

Only accessible by the owner or devs. Replies with chat name, ID, user count, and invite link.
*/
func (m moduleStruct) chatInfo(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	memStatus := db.GetTeamMemInfo(user.Id)
	cfg := m.cfg

	// only devs and owner can access this
	if user.Id != cfg.OwnerId && !memStatus.Dev {
		return ext.ContinueGroups
	}

	msg := ctx.EffectiveMessage
	var replyText string

	args := ctx.Args()

	if len(args) == 0 {
		replyText = "You must specify a user to get info on"
	} else {
		_chatId := args[1]
		chatId, _ := strconv.Atoi(_chatId)
		chat, err := b.GetChat(int64(chatId), nil)
		if err != nil {
			_, _ = msg.Reply(b, err.Error(), nil)
			return ext.EndGroups
		}
		// need to convert chat to group chat to use GetMemberCount
		_chat := chat.ToChat()
		gChat := &_chat
		con, _ := gChat.GetMemberCount(b, nil)
		replyText = fmt.Sprintf("<b>Name:</b> %s\n<b>Chat ID</b>: %d\n<b>Users Count:</b> %d\n<b>Link:</b> %s", chat.Title, chat.Id, con, chat.InviteLink)
	}

	_, err := msg.Reply(b, replyText, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.ContinueGroups
}

/*
chatList generates and sends a list of all chats the bot is in.

Only accessible by the owner or devs. Sends the list as a text file.
*/
func (m moduleStruct) chatList(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	memStatus := db.GetTeamMemInfo(user.Id)
	cfg := m.cfg

	// only devs and owner can access this
	if user.Id != cfg.OwnerId && !memStatus.Dev {
		return ext.ContinueGroups
	}

	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat

	rMsg, err := msg.Reply(
		b,
		"Getting list of chats I'm in...",
		nil,
	)
	if err != nil {
		log.Error(err)
		return err
	}

	var writeString string
	fileName := "chatlist.txt"

	allChats := db.GetAllChats()

	for chatId, v := range allChats {
		if !v.IsInactive {
			writeString += fmt.Sprintf("%d: %s\n", chatId, v.ChatName)
		}
	}

	// If the file doesn't exist, create it or re-write it
	err = os.WriteFile(fileName, []byte(writeString), 0644)
	if err != nil {
		log.Error(err)
		return err
	}

	openedFile, _ := os.Open(fileName)

	_, err = rMsg.Delete(b, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	_, err = b.SendDocument(
		chat.Id,
		gotgbot.InputFileByReader(fileName, openedFile),
		&gotgbot.SendDocumentOpts{
			Caption: "Here is the list of chats in my Database!",
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId:                msg.MessageId,
				AllowSendingWithoutReply: true,
			},
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	err = openedFile.Close()
	if err != nil {
		log.Error(err)
	}
	err = os.Remove(fileName)
	if err != nil {
		log.Error(err)
	}

	return ext.EndGroups
}

/*
leaveChat makes the bot leave a specified chat.

Only accessible by the owner or devs. Takes the chat ID as an argument.
*/
func (m moduleStruct) leaveChat(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	memStatus := db.GetTeamMemInfo(user.Id)
	tr := i18n.New(db.GetLanguage(ctx))
	cfg := m.cfg

	// only devs and owner can access this
	if user.Id != cfg.OwnerId && !memStatus.Dev {
		return ext.ContinueGroups
	}

	msg := ctx.EffectiveMessage
	args := ctx.Args()
	chatId, _ := strconv.ParseInt(args[1], 10, 64)

	_, err := b.LeaveChat(chatId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	_, err = msg.Reply(b, tr.GetString("Dev.leavechat.success"), helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.ContinueGroups
}

/*
	Function used to add sudo users in database of bot

Can only be used by OWNER
*/
/*
addSudo adds a user to the sudo list in the database.

Only the owner can use this command. Replies with the result.
*/
func (m moduleStruct) addSudo(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	cfg := m.cfg
	if user.Id != cfg.OwnerId {
		return ext.ContinueGroups
	}

	msg := ctx.EffectiveMessage
	userId := extraction.ExtractUser(b, ctx)
	if userId == -1 {
		return ext.ContinueGroups
	} else if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		return ext.ContinueGroups
	}

	reqUser, err := b.GetChat(userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}
	memStatus := db.GetTeamMemInfo(userId)

	if memStatus.Sudo {
		txt = "User is already Sudo!"
	} else {
		txt = fmt.Sprintf("Added %s to Sudo List!", helpers.MentionHtml(reqUser.Id, reqUser.FirstName))
		go db.AddSudo(userId)
	}

	_, err = msg.Reply(b, txt, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.ContinueGroups
}

/*
addDev adds a user to the dev list in the database.

Only the owner can use this command. Replies with the result.
*/
func (m moduleStruct) addDev(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	cfg := m.cfg
	if user.Id != cfg.OwnerId {
		return ext.ContinueGroups
	}

	msg := ctx.EffectiveMessage
	userId := extraction.ExtractUser(b, ctx)
	if userId == -1 {
		return ext.ContinueGroups
	} else if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		return ext.ContinueGroups
	}

	reqUser, err := b.GetChat(userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}
	memStatus := db.GetTeamMemInfo(userId)

	if memStatus.Dev {
		txt = "User is already Dev!"
	} else {
		txt = fmt.Sprintf("Added %s to Dev List!", helpers.MentionHtml(reqUser.Id, reqUser.FirstName))
		go db.AddDev(userId)
	}

	_, err = msg.Reply(b, txt, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.ContinueGroups
}

/*
remSudo removes a user from the sudo list in the database.

Only the owner can use this command. Replies with the result.
*/
func (m moduleStruct) remSudo(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	cfg := m.cfg
	if user.Id != cfg.OwnerId {
		return ext.ContinueGroups
	}

	msg := ctx.EffectiveMessage
	userId := extraction.ExtractUser(b, ctx)
	if userId == -1 {
		return ext.ContinueGroups
	} else if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		return ext.ContinueGroups
	}

	reqUser, err := b.GetChat(userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}
	memStatus := db.GetTeamMemInfo(userId)

	if !memStatus.Sudo {
		txt = "User is not Sudo!"
	} else {
		txt = fmt.Sprintf("Removed %s from Sudo List!", helpers.MentionHtml(reqUser.Id, reqUser.FirstName))
		go db.RemSudo(userId)
	}

	_, err = msg.Reply(b, txt, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.ContinueGroups
}

/*
remDev removes a user from the dev list in the database.

Only the owner can use this command. Replies with the result.
*/
func (m moduleStruct) remDev(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	cfg := m.cfg
	if user.Id != cfg.OwnerId {
		return ext.ContinueGroups
	}

	msg := ctx.EffectiveMessage
	userId := extraction.ExtractUser(b, ctx)
	if userId == -1 {
		return ext.ContinueGroups
	} else if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		return ext.ContinueGroups
	}

	reqUser, err := b.GetChat(userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}
	memStatus := db.GetTeamMemInfo(userId)

	if !memStatus.Dev {
		txt = "User is not Dev!"
	} else {
		txt = fmt.Sprintf("Removed %s from Dev List!", helpers.MentionHtml(reqUser.Id, reqUser.FirstName))
		go db.RemDev(userId)
	}

	_, err = msg.Reply(b, txt, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.ContinueGroups
}

/*
listTeam lists all team members (sudos and devs) in the database.

Only accessible by the owner or devs. Replies with the list.
*/
func (m moduleStruct) listTeam(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	memStatus := db.GetTeamMemInfo(user.Id)
	cfg := m.cfg

	// only devs and owner can access this
	if user.Id != cfg.OwnerId && !memStatus.Dev {
		return ext.ContinueGroups
	}

	msg := ctx.EffectiveMessage
	edits, err := msg.Reply(
		b,
		"<code>Getting team members...</code>",
		&gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	teamMembers := db.GetTeamMembers()
	var teamText string
	if len(teamMembers) == 0 {
		teamText = "No team members found!"
	} else {
		teamText = "<b>Team Members:</b>\n"
		for userId, role := range teamMembers {
			reqUser, err := b.GetChat(userId, nil)
			if err != nil {
				teamText += fmt.Sprintf("• %d (%s)\n", userId, role)
			} else {
				teamText += fmt.Sprintf("• %s (%s)\n", helpers.MentionHtml(reqUser.Id, helpers.GetFullName(reqUser.FirstName, reqUser.LastName)), role)
			}
		}
	}

	_, _, err = edits.EditText(
		b,
		teamText,
		&gotgbot.EditMessageTextOpts{
			ParseMode: helpers.HTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.ContinueGroups
}

/*
getStats fetches and displays bot statistics.

Only accessible by the owner or devs. Replies with database stats.
*/
func (m moduleStruct) getStats(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	memStatus := db.GetTeamMemInfo(user.Id)
	cfg := m.cfg

	// only devs and owner can access this
	if user.Id != cfg.OwnerId && !memStatus.Dev {
		return ext.ContinueGroups
	}

	msg := ctx.EffectiveMessage
	edits, err := msg.Reply(
		b,
		"<code>Fetching bot stats...</code>",
		&gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	stats := db.LoadAllStats()
	_, _, err = edits.EditText(
		b,
		stats,
		&gotgbot.EditMessageTextOpts{
			ParseMode: helpers.HTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.ContinueGroups
}

/*
LoadDev registers all developer/admin command handlers with the dispatcher.

Enables the dev module and adds handlers for team management, chat info, stats, and database cleanup.
*/
func LoadDev(dispatcher *ext.Dispatcher, cfg *config.Config) {
	// Store config in the module
	devsModule.cfg = cfg

	dispatcher.AddHandler(handlers.NewCommand("stats", devsModule.getStats))
	dispatcher.AddHandler(handlers.NewCommand("addsudo", devsModule.addSudo))
	dispatcher.AddHandler(handlers.NewCommand("adddev", devsModule.addDev))
	dispatcher.AddHandler(handlers.NewCommand("remsudo", devsModule.remSudo))
	dispatcher.AddHandler(handlers.NewCommand("remdev", devsModule.remDev))
	dispatcher.AddHandler(handlers.NewCommand("teamusers", devsModule.listTeam))
	dispatcher.AddHandler(handlers.NewCommand("chatinfo", devsModule.chatInfo))
	dispatcher.AddHandler(handlers.NewCommand("chatlist", devsModule.chatList))
	dispatcher.AddHandler(handlers.NewCommand("leavechat", devsModule.leaveChat))
	dispatcher.AddHandler(handlers.NewCommand("dbclean", devsModule.dbClean))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("dbclean."), devsModule.dbCleanButtonHandler))
}
