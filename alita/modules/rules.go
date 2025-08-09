package modules

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	tgmd2html "github.com/PaulSonOfLars/gotg_md2html"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/cmdDecorator"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/misc"
)

var rulesModule = moduleStruct{
	moduleName:      "Rules",
	defaultRulesBtn: "Rules",
}

// clearRules handles commands to completely remove all rules
// from the chat, requiring admin permissions.
func (moduleStruct) clearRules(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

	go db.SetChatRules(chat.Id, "")
	_, err := msg.Reply(bot, translator.Message("rules_cleared_success", nil), nil)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// privaterules handles the /privaterules command to toggle whether
// rules are sent privately or in the group chat.
func (moduleStruct) privaterules(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	args := ctx.Args()
	var text string

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

	if len(args) >= 2 {
		switch strings.ToLower(args[1]) {
		case "on", "yes", "true":
			go db.SetPrivateRules(chat.Id, true)
			text = translator.Message("rules_private_enabled", nil)
		case "off", "no", "false":
			go db.SetPrivateRules(chat.Id, false)
			text = translator.Message("rules_private_disabled", i18n.Params{
				"chat_title": chat.Title,
			})
		default:
			text = translator.Message("rules_invalid_option", nil)
		}
	} else {
		rulesprefs := db.GetChatRulesInfo(chat.Id)
		if rulesprefs.Private {
			text = translator.Message("rules_private_enabled", nil)
		} else {
			text = translator.Message("rules_private_disabled", i18n.Params{
				"chat_title": chat.Title,
			})
		}
	}

	_, err := msg.Reply(bot, text, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// sendRules handles the /rules command to display chat rules
// either in the group or privately based on settings.
func (m moduleStruct) sendRules(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// if command is disabled, return
	if chat_status.CheckDisabledCmd(bot, msg, "adminlist") {
		return ext.EndGroups
	}
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, false, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	var (
		replyMsgId int64
		Text       = ""
		err        error
		rulesKb    gotgbot.InlineKeyboardMarkup
		rulesBtn   string
	)

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

	if reply := msg.ReplyToMessage; reply != nil {
		replyMsgId = reply.MessageId
	} else {
		replyMsgId = msg.MessageId
	}

	rules := db.GetChatRulesInfo(chat.Id)
	rulesBtn = rules.RulesBtn
	if rulesBtn == "" {
		rulesBtn = m.defaultRulesBtn
	}
	if rules.Rules != "" {
		Text += translator.Message("rules_for_chat", i18n.Params{
			"chat_title": chat.Title,
		}) + "\n\n"
		Text += rules.Rules
	} else {
		Text += translator.Message("rules_no_rules_set", nil)
	}

	if chat_status.RequireGroup(bot, ctx, nil, true) && rules.Private {
		Text = translator.Message("rules_click_for_rules", nil)
		rulesKb = gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{
						Text: rulesBtn,
						Url:  fmt.Sprintf("t.me/%s?start=rules_%d", bot.Username, chat.Id),
					},
				},
			},
		}
	}

	_, err = msg.Reply(bot, Text,
		&gotgbot.SendMessageOpts{
			ReplyMarkup: rulesKb,
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId:                replyMsgId,
				AllowSendingWithoutReply: true,
			},
			ParseMode: helpers.HTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// setRules handles the /setrules command to create or update
// chat rules with markdown formatting support.
func (moduleStruct) setRules(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	args := ctx.Args()
	var text string

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))

	if len(args) == 1 && msg.ReplyToMessage == nil {
		text = translator.Message("rules_need_rules_to_set", nil)
	} else {
		if msg.ReplyToMessage != nil {
			text = msg.ReplyToMessage.OriginalMDV2()
		} else {
			text = strings.SplitN(msg.OriginalMDV2(), " ", 2)[1]
		}
		go db.SetChatRules(chat.Id, tgmd2html.MD2HTMLV2(text))
		text = translator.Message("rules_set_success", nil)
	}

	_, err := msg.Reply(bot, text, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// rulesBtn handles the /rulesbutton command to set or view
// the custom button text for private rules links.
func (m moduleStruct) rulesBtn(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()
	var err error
	var text string

	if !chat_status.IsUserAdmin(bot, chat.Id, user.Id) {
		return ext.EndGroups
	}

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))
	
	if len(args) >= 2 {
		rulesBtnCustomText := strings.Join(args[1:], " ")
		if len(rulesBtnCustomText) > 30 {
			text = translator.Message("rules_button_too_long", nil)
		} else {
			text = translator.Message("rules_button_set_success", i18n.Params{
				"button_text": rulesBtnCustomText,
			})
			go db.SetChatRulesButton(chat.Id, rulesBtnCustomText)
		}
	} else {
		customRulesBtn := db.GetChatRulesInfo(chat.Id).RulesBtn
		if customRulesBtn == "" {
			text = translator.Message("rules_button_not_set", i18n.Params{
				"default_text": m.defaultRulesBtn,
			})
		} else {
			text = translator.Message("rules_button_current", i18n.Params{
				"button_text": customRulesBtn,
			})
		}
	}

	_, err = msg.Reply(bot, text, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// resetRulesBtn handles commands to reset the custom rules button
// text back to the default value.
func (moduleStruct) resetRulesBtn(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat

	// Get translator for the chat
	translator := i18n.MustNewTranslator(db.GetLanguage(ctx))
	
	go db.SetChatRulesButton(chat.Id, "")
	_, err := msg.Reply(bot, translator.Message("rules_button_reset_success", nil), nil)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// LoadRules registers all rules module handlers with the dispatcher,
// including rules management and display commands.
func LoadRules(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(rulesModule.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("rules", rulesModule.sendRules))
	misc.AddCmdToDisableable("rules")
	dispatcher.AddHandler(handlers.NewCommand("setrules", rulesModule.setRules))
	cmdDecorator.MultiCommand(dispatcher, []string{"resetrules", "clearrules"}, rulesModule.clearRules)
	dispatcher.AddHandler(handlers.NewCommand("privaterules", rulesModule.privaterules))
	dispatcher.AddHandler(handlers.NewCommand("rulesbutton", rulesModule.rulesBtn))
	dispatcher.AddHandler(handlers.NewCommand("rulesbtn", rulesModule.rulesBtn))
	dispatcher.AddHandler(handlers.NewCommand("clearrulesbutton", rulesModule.resetRulesBtn))
	dispatcher.AddHandler(handlers.NewCommand("clearrulesbtn", rulesModule.resetRulesBtn))
	dispatcher.AddHandler(handlers.NewCommand("resetrulesbutton", rulesModule.resetRulesBtn))
	dispatcher.AddHandler(handlers.NewCommand("resetrulesbtn", rulesModule.resetRulesBtn))
}
