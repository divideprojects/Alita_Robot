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
	"github.com/divideprojects/Alita_Robot/alita/i18n"
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

// report handles the /report command and @admin mentions to notify
// administrators about problematic messages with action buttons.
func (moduleStruct) report(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := ctx.EffectiveMessage

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "report") {
		return ext.EndGroups
	}

	if msg.ReplyToMessage == nil {
		_, _ = msg.Reply(b,
			translator.Message("reports_need_reply", nil),
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
		_, _ = msg.Reply(b, translator.Message("reports_cant_report_own", nil), nil)
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
		_, _ = msg.Reply(b, translator.Message("reports_expose_yourself", nil), nil)
		return ext.EndGroups
	}
	if msg.ReplyToMessage.From.Id == 1087968824|777000|136817688 {
		_, _ = msg.Reply(b, translator.Message("reports_special_account", nil), nil)
		return ext.EndGroups
	}

	if chat_status.IsUserAdmin(b, chat.Id, user.Id) {
		_, err := msg.Reply(b, translator.Message("reports_admin_cant_report", nil), helpers.Shtml())
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
		admins = cache.LoadAdminCache(b, chat.Id)
	}

	for i := range admins.UserInfo {
		admin := &admins.UserInfo[i]
		adminArray = append(adminArray, admin.User.Id)
	}

	reportedUser := msg.ReplyToMessage.From
	reportedMsgId := msg.ReplyToMessage.MessageId

	if reportedUser.Id == b.Id {
		_, err := msg.Reply(b, translator.Message("reports_cant_report_bot", nil), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}
	if string_handling.FindInInt64Slice(adminArray, reportedUser.Id) {
		_, err := msg.Reply(b, translator.Message("reports_cant_report_admin", nil), helpers.Shtml())
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
			ParseMode: helpers.HTML,
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId:                replyMsgId,
				AllowSendingWithoutReply: true,
			},
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{
							Text: translator.Message("reports_button_message", nil),
							Url:  helpers.GetMessageLinkFromMessageId(chat, reportedMsgId),
						},
					},
					{
						{
							Text:         translator.Message("reports_button_kick", nil),
							CallbackData: fmt.Sprintf(callbackData, "kick"),
						},
						{
							Text:         translator.Message("reports_button_ban", nil),
							CallbackData: fmt.Sprintf(callbackData, "ban"),
						},
					},
					{
						{
							Text:         translator.Message("reports_button_delete_message", nil),
							CallbackData: fmt.Sprintf(callbackData, "delete"),
						},
					},
					{
						{
							Text:         translator.Message("reports_button_mark_resolved", nil),
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

// reports handles the /reports command to manage reporting settings
// for both users and chats, including blocking and status changes.
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

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

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
				replyText = translator.Message("reports_user_enabled", nil)
				go db.SetUserReportSettings(user.Id, true)
			} else {
				replyText = translator.Message("reports_chat_enabled", nil)
				go db.SetChatReportStatus(chat.Id, true)
			}
		case "off", "no", "false":
			if msg.Chat.Type == "private" {
				replyText = translator.Message("reports_user_disabled", nil)
				go db.SetUserReportSettings(user.Id, false)
			} else {
				replyText = translator.Message("reports_chat_disabled", nil)
				go db.SetChatReportStatus(chat.Id, false)
			}
		case "block":
			if msg.Chat.Type == "private" {
				replyText = translator.Message("reports_group_only", nil)
			} else {
				if reply := msg.ReplyToMessage; reply != nil {
					bUser := reply.From
					go db.BlockReportUser(chat.Id, bUser.Id)
					replyText = translator.Message("reports_blocked_user", i18n.Params{
						"mention": helpers.MentionHtml(bUser.Id, bUser.FirstName),
					})
				} else {
					replyText = translator.Message("reports_reply_to_block", nil)
				}
			}
		case "unblock":
			if msg.Chat.Type == "private" {
				replyText = translator.Message("reports_group_only", nil)
			} else {
				if reply := msg.ReplyToMessage; reply != nil {
					bUser := reply.From
					go db.UnblockReportUser(chat.Id, bUser.Id)
					replyText = translator.Message("reports_unblocked_user", i18n.Params{
						"mention": helpers.MentionHtml(bUser.Id, bUser.FirstName),
					})
				} else {
					replyText = translator.Message("reports_reply_to_unblock", nil)
				}
			}
		case "showblocklist":
			if msg.Chat.Type == "private" {
				replyText = translator.Message("reports_group_only", nil)
			} else {
				blockedUsers := db.GetChatReportSettings(chat.Id).BlockedList
				if len(blockedUsers) == 0 {
					replyText = translator.Message("reports_no_blocked_users", nil)
				} else {
					var builder strings.Builder
					builder.Grow(256) // Pre-allocate capacity
					builder.WriteString(translator.Message("reports_blocked_users_list", nil))
					for _, blockUserId := range blockedUsers {
						bUser, err := b.GetChat(blockUserId, nil)
						if err != nil {
							log.Error(err)
							continue
						}
						builder.WriteString("\n - ")
						builder.WriteString(helpers.MentionHtml(blockUserId, bUser.FirstName))
					}
					replyText = builder.String()
				}
			}
		default:
			replyText = translator.Message("reports_invalid_option", nil)
		}
	} else {
		if msg.Chat.Type == "private" {
			rStatus := db.GetUserReportSettings(chat.Id).Status
			if rStatus {
				replyText = translator.Message("reports_user_status_enabled", nil)
			} else {
				replyText = translator.Message("reports_user_status_disabled", nil)
			}
		} else {
			rStatus := db.GetChatReportSettings(chat.Id).Status
			if rStatus {
				replyText = translator.Message("reports_chat_status_enabled", nil)
			} else {
				replyText = translator.Message("reports_chat_status_disabled", nil)
			}
		}
		replyText += "\n\n" + translator.Message("reports_change_setting_help", nil)
	}

	_, err = msg.Reply(b, replyText, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

// markResolvedButtonHandler processes callback queries from report action buttons
// to kick, ban, delete messages, or mark reports as resolved.
func (moduleStruct) markResolvedButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	msg := query.Message
	var replyQuery, replyText string

	// Get translator for callback queries
	translator := i18n.MustNewTranslator("en") // fallback for callback queries

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
		replyQuery = translator.Message("reports_action_kicked", nil)
		replyText = translator.Message("reports_user_kicked_by", i18n.Params{
			"admin": helpers.MentionHtml(user.Id, user.FirstName),
		})
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
		replyQuery = translator.Message("reports_action_banned", nil)
		replyText = translator.Message("reports_user_banned_by", i18n.Params{
			"admin": helpers.MentionHtml(user.Id, user.FirstName),
		})
		_, err := chat.BanMember(b, userId, nil)
		if err != nil {
			log.Error(err)
			return err
		}

	case "delete":
		replyQuery = translator.Message("reports_action_deleted", nil)
		replyText = translator.Message("reports_message_deleted_by", i18n.Params{
			"admin": helpers.MentionHtml(user.Id, user.FirstName),
		})
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
		replyQuery = translator.Message("reports_action_resolved", nil)
		replyText = translator.Message("reports_resolved_by", i18n.Params{
			"admin": helpers.MentionHtml(user.Id, user.FirstName),
		})

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

// LoadReports registers all reports module handlers with the dispatcher,
// including report commands and @admin mention monitoring.
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
