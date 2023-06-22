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
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/cmdDecorator"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/misc"
)

var rulesModule = moduleStruct{
	moduleName:      "Rules",
	defaultRulesBtn: "Rules",
}

func (moduleStruct) clearRules(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat

	go db.SetChatRules(chat.Id, "")
	_, err := msg.Reply(bot, "Successfully cleared rules!", nil)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

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

	if len(args) >= 2 {
		switch strings.ToLower(args[1]) {
		case "on", "yes", "true":
			go db.SetPrivateRules(chat.Id, true)
			text = "Use of /rules will send the rules to the user's PM."
		case "off", "no", "false":
			go db.SetPrivateRules(chat.Id, false)
			text = fmt.Sprintf("All /rules commands will send the rules to %s.", chat.Title)
		default:
			text = "Your input was not recognised as one of: yes/no/on/off"
		}
	} else {
		rulesprefs := db.GetChatRulesInfo(chat.Id)
		if rulesprefs.Private {
			text = "Use of /rules will send the rules to the user's PM."
		} else {
			text = fmt.Sprintf("All /rules commands will send the rules to %s.", chat.Title)
		}
	}

	_, err := msg.Reply(bot, text, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

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
		Text += fmt.Sprintf("The rules for <b>%s</b> are:\n\n", chat.Title)
		Text += rules.Rules
	} else {
		Text += "This chat doesn't seem to have had any rules set yet... I wouldn't take that as an invitation though."
	}

	if chat_status.RequireGroup(bot, ctx, nil, true) && rules.Private {
		Text = "Click on the button to see the chat rules!"
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
			ReplyMarkup:              rulesKb,
			ReplyToMessageId:         replyMsgId,
			AllowSendingWithoutReply: true,
			ParseMode:                helpers.HTML,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

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

	if len(args) == 1 && msg.ReplyToMessage == nil {
		text = "You need to give me rules to set!\nEither reply to a message to provide them with command."
	} else {
		if msg.ReplyToMessage != nil {
			text = msg.ReplyToMessage.OriginalMDV2()
		} else {
			text = strings.SplitN(msg.OriginalMDV2(), " ", 2)[1]
		}
		go db.SetChatRules(chat.Id, tgmd2html.MD2HTMLV2(text))
		text = "Successfully set rules for this group!"
	}

	_, err := msg.Reply(bot, text, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

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

	if len(args) >= 2 {
		rulesBtnCustomText := strings.Join(args[1:], " ")
		if len(rulesBtnCustomText) > 30 {
			text = "The custom button name you entered is too long. Please enter a text with less than 30 characters."
		} else {
			text = fmt.Sprintf("Successfully set the rules button to: <b>%s</b>", rulesBtnCustomText)
			go db.SetChatRulesButton(chat.Id, rulesBtnCustomText)
		}
	} else {
		customRulesBtn := db.GetChatRulesInfo(chat.Id).RulesBtn
		if customRulesBtn == "" {
			text = fmt.Sprintf("You haven't set a custom rules button yet. The default text \"%s\" will be used.", m.defaultRulesBtn)
		} else {
			text = fmt.Sprintf("The rules button is currently set to the following text:\n %s", customRulesBtn)
		}
	}

	_, err = msg.Reply(bot, text, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

func (moduleStruct) resetRulesBtn(bot *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(bot, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat

	go db.SetChatRulesButton(chat.Id, "")
	_, err := msg.Reply(bot, "Successfully cleared custom rules button text!", nil)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

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
