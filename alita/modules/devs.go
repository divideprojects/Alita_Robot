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
	"github.com/divideprojects/Alita_Robot/alita/utils/extraction"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

var devsModule = moduleStruct{moduleName: "Dev"}

// for general purposes for strings in functions below
var txt string

func (moduleStruct) chatInfo(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	memStatus := db.GetTeamMemInfo(user.Id)

	// only devs and owner can access this
	if user.Id != config.OwnerId && !memStatus.Dev {
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
		con, _ := chat.GetMemberCount(b, nil)
		replyText = fmt.Sprintf("<b>Name:</b> %s\n<b>Chat ID</b>: %d\n<b>Users Count:</b> %d\n<b>Link:</b> %s", chat.Title, chat.Id, con, chat.InviteLink)
	}

	_, err := msg.Reply(b, replyText, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.ContinueGroups
}

func (moduleStruct) chatList(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	memStatus := db.GetTeamMemInfo(user.Id)

	// only devs and owner can access this
	if user.Id != config.OwnerId && !memStatus.Dev {
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
		openedFile,
		&gotgbot.SendDocumentOpts{
			Caption:                  "Here is the list of chats in my Database!",
			ReplyToMessageId:         msg.MessageId,
			AllowSendingWithoutReply: true,
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

func (moduleStruct) leaveChat(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	memStatus := db.GetTeamMemInfo(user.Id)

	// only devs and owner can access this
	if user.Id != config.OwnerId && !memStatus.Dev {
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

	_, err = msg.Reply(b, "Okay, I left the chat!", helpers.Shtml())
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
func (moduleStruct) addSudo(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	if user.Id != config.OwnerId {
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
	_, err = msg.Reply(b, txt, &gotgbot.SendMessageOpts{ParseMode: helpers.HTML})
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.ContinueGroups
}

/*
	Function used to add dev users in database of bot

Can only be used by OWNER
*/
func (moduleStruct) addDev(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	if user.Id != config.OwnerId {
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
	_, err = msg.Reply(b, txt, &gotgbot.SendMessageOpts{ParseMode: helpers.HTML})
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.ContinueGroups
}

/*
	Function used to remove sudo users from database of bot

Can only be used by OWNER
*/
func (moduleStruct) remSudo(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	if user.Id != config.OwnerId {
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
	_, err = msg.Reply(b, txt, &gotgbot.SendMessageOpts{ParseMode: helpers.HTML})
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.ContinueGroups
}

/*
	Function used to remove dev users from database of bot

Can only be used by OWNER
*/
func (moduleStruct) remDev(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	if user.Id != config.OwnerId {
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
	_, err = msg.Reply(b, txt, &gotgbot.SendMessageOpts{ParseMode: helpers.HTML})
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.ContinueGroups
}

/*
	Function used to list all members of bot's development team

Can only be used by existing team members
*/
func (moduleStruct) listTeam(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User

	teamUsers := db.GetTeamMembers()
	var teamint64Slice []int64
	for k := range teamUsers {
		teamint64Slice = append(teamint64Slice, k)
	}
	teamint64Slice = append(teamint64Slice, config.OwnerId)

	if !string_handling.FindInInt64Slice(teamint64Slice, user.Id) {
		return ext.EndGroups
	}

	var (
		txt       string
		dev       = "<b>Dev Users:</b>\n"
		sudo      = "<b>Sudo Users:</b>\n"
		sudoUsers = make([]string, 0)
		devUsers  = make([]string, 0)
	)
	msg := ctx.EffectiveMessage

	if len(teamUsers) == 0 {
		txt = "No users are added Added in Team!"
	} else {
		for userId, uPerm := range teamUsers {
			reqUser, err := b.GetChat(userId, nil)
			if err != nil {
				log.Error(err)
				return err
			}

			userMentioned := helpers.MentionHtml(reqUser.Id, helpers.GetFullName(reqUser.FirstName, reqUser.LastName))
			if uPerm == "dev" {
				devUsers = append(devUsers, fmt.Sprintf("• %s", userMentioned))
			} else if uPerm == "sudo" {
				sudoUsers = append(sudoUsers, fmt.Sprintf("• %s", userMentioned))
			}
		}
		if len(sudoUsers) == 0 {
			sudo += "\nNo Users"
		} else {
			sudo += strings.Join(sudoUsers, "\n")
		}
		if len(devUsers) == 0 {
			dev += "\nNo Users"
		} else {
			dev += strings.Join(devUsers, "\n")
		}
		txt = dev + "\n\n" + sudo
	}

	_, err := msg.Reply(b, txt, &gotgbot.SendMessageOpts{ParseMode: helpers.HTML})
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
	Function used to fetch stats of bot

Can only be used by OWNER
*/
func (moduleStruct) getStats(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	memStatus := db.GetTeamMemInfo(user.Id)

	// only devs and owner can access this
	if user.Id != config.OwnerId && !memStatus.Dev {
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

func LoadDev(dispatcher *ext.Dispatcher) {
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
