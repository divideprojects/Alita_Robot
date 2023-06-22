package modules

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/misc"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

var reportsModule = moduleStruct{
	moduleName:   "Reports",
	handlerGroup: 8,
}

func (moduleStruct) report(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage

	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "report") {
		return ext.EndGroups
	}

	if msg.ReplyToMessage == nil {
		_, _ = msg.Reply(b,
			"You need to reply to a message to report it.",
			nil,
		)
		return ext.EndGroups
	}

	var (
		replyMsgId int64
		adminArray []int64
		err        error
	)

	if msg.ReplyToMessage.From.Id == user.Id {
		_, _ = msg.Reply(b, "You can't report your own message!", nil)
		return ext.EndGroups
	}

	if replyMsg := msg.ReplyToMessage; replyMsg != nil {
		replyMsgId = replyMsg.MessageId
	} else {
		replyMsgId = msg.MessageId
	}
	reportprefs := db.GetChatReportSettings(chat.Id)

	// don't let blocked users report
	if string_handling.FindInInt64Slice(reportprefs.BlockedList, user.Id) {
		if chat_status.CanBotDelete(b, ctx, nil, true) {
			_, err := msg.Delete(b, nil)
			if err != nil {
				log.Error(err)
			}

		}
		return ext.EndGroups
	}

	if user.Id == 1087968824|777000|136817688 {
		_, _ = msg.Reply(b, "You need to expose yourself first!", nil)
		return ext.EndGroups
	}
	if msg.ReplyToMessage.From.Id == 1087968824|777000|136817688 {
		_, _ = msg.Reply(b, "It's a special account of telegram!", nil)
		return ext.EndGroups
	}

	if chat_status.IsUserAdmin(b, chat.Id, user.Id) {
		_, err := msg.Reply(b, "You're an admin, whom will I report your issues to?", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if !reportprefs.Status {
		return ext.EndGroups
	}

	adminsAvail, admins := cache.GetAdminCacheList(chat.Id)
	if !adminsAvail {
		admins = cache.LoadAdminCache(b, chat)
	}

	for i := range admins.UserInfo {
		admin := &admins.UserInfo[i]
		adminArray = append(adminArray, admin.User.Id)
	}

	reportedUser := msg.ReplyToMessage.From
	reportedMsgId := msg.ReplyToMessage.MessageId

	if reportedUser.Id == b.Id {
		_, err := msg.Reply(b, "Why would I report myself?", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}
	if string_handling.FindInInt64Slice(adminArray, reportedUser.Id) {
		_, err := msg.Reply(b, "Why would I report an admin?", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	reported := fmt.Sprintf(
		"<b>⚠️ Report:</b>"+
			"\n<b> • Report by:</b> %s"+
			"\n<b> • Reported user:</b> %s"+
			"\n<b>Status:</b> <i>Pending...</i>",
		helpers.MentionHtml(user.Id, user.FirstName),
		helpers.MentionHtml(reportedUser.Id, reportedUser.FirstName),
	)
	for _, adminUserId := range adminArray {
		if !db.GetUserReportSettings(adminUserId).Status {
			continue
		}
		reported += helpers.MentionHtml(adminUserId, "\u2063")
	}

	callbackData := "report." + "%s=" + fmt.Sprint(user.Id) + "=" + fmt.Sprint(reportedMsgId)
	_, err = msg.Reply(b,
		reported,
		&gotgbot.SendMessageOpts{
			ParseMode:                helpers.HTML,
			ReplyToMessageId:         replyMsgId,
			AllowSendingWithoutReply: true,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text: "➡ Message",
							Url:  helpers.GetMessageLinkFromMessageId(chat, reportedMsgId),
						},
					},
					{
						{
							Text:         "⚠ Kick",
							CallbackData: fmt.Sprintf(callbackData, "kick"),
						},
						{
							Text:         "⛔️ Ban",
							CallbackData: fmt.Sprintf(callbackData, "ban"),
						},
					},
					{
						{
							Text:         "❎ Delete Message",
							CallbackData: fmt.Sprintf(callbackData, "delete"),
						},
					},
					{
						{
							Text:         "✔️ Mark Resolved",
							CallbackData: fmt.Sprintf(callbackData, "resolved"),
						},
					},
				},
			},
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

func (moduleStruct) reports(b *gotgbot.Bot, ctx *ext.Context) error {
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage
	args := ctx.Args()[1:]

	var (
		err       error
		replyText string
	)

	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}

	if len(args) >= 1 {
		action := strings.ToLower(args[0])
		switch action {
		case "on", "yes", "true":
			if msg.Chat.Type == "private" {
				replyText = "Turned on reporting! You'll be notified whenever anyone reports something in groups you are admin."
				go db.SetUserReportSettings(user.Id, true)
			} else {
				replyText = "Users will now be able to report messages."
				go db.SetChatReportStatus(chat.Id, true)
			}
		case "off", "no", "false":
			if msg.Chat.Type == "private" {
				replyText = "Turned off reporting! You'll no longer be notified whenever anyone reports something in groups you are admin."
				go db.SetUserReportSettings(user.Id, false)
			} else {
				replyText = "Users will no longer be able to report via @admin or /report."
				go db.SetChatReportStatus(chat.Id, false)
			}
		case "block":
			if msg.Chat.Type == "private" {
				replyText = "This command can only be used in a group!"
			} else {
				if reply := msg.ReplyToMessage; reply != nil {
					bUser := reply.From
					go db.BlockReportUser(chat.Id, bUser.Id)
					replyText = fmt.Sprintf("Blocked user %s from reporting.", helpers.MentionHtml(bUser.Id, bUser.FirstName))
				} else {
					replyText = "You must reply to a user to block them."
				}
			}
		case "unblock":
			if msg.Chat.Type == "private" {
				replyText = "This command can only be used in a group!"
			} else {
				if reply := msg.ReplyToMessage; reply != nil {
					bUser := reply.From
					go db.UnblockReportUser(chat.Id, bUser.Id)
					replyText = fmt.Sprintf("Unblocked user %s from reporting.", helpers.MentionHtml(bUser.Id, bUser.FirstName))
				} else {
					replyText = "You must reply to a user to unblock them."
				}
			}
		case "showblocklist":
			if msg.Chat.Type == "private" {
				replyText = "This command can only be used in a group!"
			} else {
				blockedUsers := db.GetChatReportSettings(chat.Id).BlockedList
				if len(blockedUsers) == 0 {
					replyText = "No users are currently blocked from using report commands!"
				} else {
					replyText = "Users blocked from using report commands: "
					for _, blockUserId := range blockedUsers {
						bUser, err := b.GetChat(blockUserId, nil)
						if err != nil {
							log.Error(err)
							continue
						}
						replyText += "\n - " + helpers.MentionHtml(blockUserId, bUser.FirstName)
					}
				}
			}
		default:
			replyText = "Your input was not recognised as one of: <code><yes/on/no/off> or <block/unblock/showblocklist></code>"
		}
	} else {
		if msg.Chat.Type == "private" {
			rStatus := db.GetUserReportSettings(chat.Id).Status
			if rStatus {
				replyText = "Your current preference is true, You'll be notified whenever anyone reports something in groups you are admin."
			} else {
				replyText = "You'll have nt enabled reports, You won't be notified."
			}
		} else {
			rStatus := db.GetChatReportSettings(chat.Id).Status
			if rStatus {
				replyText = "Reports are currently enabled in this chat.\nUsers can use the /report command, or mention @admin, to tag all admins."
			} else {
				replyText = "Reports are currently disabled in this chat."
			}
		}
		replyText += "\n\nTo change this setting, try this command again, with one of the following args: yes/no/on/off"
	}

	_, err = msg.Reply(b, replyText, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

func (moduleStruct) markResolvedButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := query.Message
	var replyQuery, replyText string

	// permissions check
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	args := strings.Split(strings.Split(query.Data, ".")[1], "=")
	action := args[0]
	_userId, _ := strconv.Atoi(args[1])
	_msgId, _ := strconv.Atoi(args[2])

	userId := int64(_userId)
	msgId := int64(_msgId)

	switch action {
	case "kick":
		replyQuery = "✅ Successfully Kicked"
		replyText = fmt.Sprintf(
			"User kicked!"+
				"Action taken by %s",
			helpers.MentionHtml(user.Id, user.FirstName),
		)
		_, err := chat.BanMember(b, userId, nil)
		if err != nil {
			log.Error(err)
			return err
		}

		time.Sleep(1 * time.Second) // wait for sometime before unbanning

		_, err = chat.UnbanMember(b, userId, nil)
		if err != nil {
			log.Error(err)
			return err
		}
	case "ban":
		replyQuery = "✅ Successfully Banned"
		replyText = fmt.Sprintf(
			"User banned!"+
				"Action taken by %s",
			helpers.MentionHtml(user.Id, user.FirstName),
		)
		_, err := chat.BanMember(b, userId, nil)
		if err != nil {
			log.Error(err)
			return err
		}

	case "delete":
		replyQuery = "✅ Successfully Deleted"
		replyText = fmt.Sprintf(
			"Message Deleted!"+
				"Action taken by %s",
			helpers.MentionHtml(user.Id, user.FirstName),
		)
		_, err := b.DeleteMessage(chat.Id, msgId, nil)
		if err.Error() == "unable to deleteMessage: Bad Request: message to delete not found" {
			log.WithFields(
				log.Fields{
					"chat": chat.Id,
				},
			).Error("error deleting message")
			return ext.EndGroups
		} else if err != nil {
			log.Error(err)
			return err
		}
	default:
		replyQuery = "✅ Resolved Report Successfully!"
		replyText = fmt.Sprintf(
			"<b>Resolved by:</b> %s",
			helpers.MentionHtml(user.Id, user.FirstName),
		)

	}
	_, _, err := msg.EditText(
		b,
		replyText,
		&gotgbot.EditMessageTextOpts{
			ChatId:    chat.Id,
			ParseMode: helpers.HTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	_, err = query.Answer(b,
		&gotgbot.AnswerCallbackQueryOpts{
			Text: replyQuery,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

func LoadReports(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(reportsModule.moduleName, true)

	dispatcher.AddHandlerToGroup(
		handlers.NewMessage(
			func(msg *gotgbot.Message) bool {
				r, _ := regexp.Compile("(?i)@admin(s)?")
				return r.MatchString(msg.Text)
			},
			reportsModule.report,
		),
		reportsModule.handlerGroup,
	)
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("report."), reportsModule.markResolvedButtonHandler))
	dispatcher.AddHandler(handlers.NewCommand("report", reportsModule.report))
	misc.AddCmdToDisableable("report")
	dispatcher.AddHandler(handlers.NewCommand("reports", reportsModule.reports))
}
