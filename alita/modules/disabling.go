package modules

import (
	"fmt"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/misc"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

var disablingModule = moduleStruct{moduleName: "Disabling"}

/*
	To disable a command

# Connection - true, true

Only Admin can use this command to disable usage of a command in the chat
*/
func (moduleStruct) disable(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	args := ctx.Args()[1:]

	if len(args) >= 1 {
		toDisable := make([]string, 0)

		for _, i := range args {
			i = strings.ToLower(i)
			if string_handling.FindInStringSlice(misc.DisableCmds, i) {
				toDisable = append(toDisable, i)
				_, err := msg.Reply(b, fmt.Sprintf("Disabled the use of the following in this chat:"+
					"%s",
					strings.Join(toDisable, "\n - ")),
					helpers.Smarkdown())
				if err != nil {
					log.Error(err)
					return err
				}
			} else {
				_, err := msg.Reply(b,
					fmt.Sprintf("Unknown command to disable:\n-%s\nCheck /disableable!", i), nil)
				if err != nil {
					log.Error(err)
					return err
				}
			}
		}
		// finally disable all cmds
		for _, i := range toDisable {
			db.DisableCMD(chat.Id, i)
		}

	} else {
		_, err := msg.Reply(b, "You haven't specified a command to disable.", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

/*
	To check the disableable commands

Anyone can use this command to check the disableable commands
*/
func (moduleStruct) disableable(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage

	text := "The following commands can be disabled:"
	for _, cmds := range misc.DisableCmds {
		text += fmt.Sprintf("\n - `%s`", cmds)
	}

	_, err := msg.Reply(b, text, helpers.Smarkdown())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
	To list all disabled commands in chat

# Connection - false, true

Any user in can use this command to check the disabled commands in the current chat.
*/
func (moduleStruct) disabled(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "adminlist") {
		return ext.EndGroups
	}
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, false, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat

	var replyMsgId int64

	if reply := msg.ReplyToMessage; reply != nil {
		replyMsgId = reply.MessageId
	} else {
		replyMsgId = msg.MessageId
	}

	disabled := db.GetChatDisabledCMDs(chat.Id)

	if len(disabled) == 0 {
		_, err := msg.Reply(b,
			"There are no disabled commands in this chat.",
			&gotgbot.SendMessageOpts{
				ReplyToMessageId:         replyMsgId,
				AllowSendingWithoutReply: true,
			},
		)
		if err != nil {
			log.Error(err)
			return err
		}
	} else {
		text := "The following commands are disabled in this chat:"
		sort.Strings(disabled)
		for _, cmds := range disabled {
			text += fmt.Sprintf("\n - `%s`", cmds)
		}
		_, err := msg.Reply(b, text, helpers.Smarkdown())
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.EndGroups
}

/*
	To either delete or not to delete the disbaled command in the chat

# Connection - true, true

Only admins can use this command to either choose to delete the disabled command
or not to. If no argument is given, the current chat setting is returned
*/
func (moduleStruct) disabledel(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	args := ctx.Args()[1:]

	var text string

	if len(args) >= 1 {
		param := strings.ToLower(args[0])
		switch param {
		case "on", "true", "yes":
			go db.ToggleDel(chat.Id, true)
			text = "Disabled messages will now be deleted."
		case "off", "false", "no":
			go db.ToggleDel(chat.Id, false)
			text = "Disabled messages will no longer be deleted."
		default:
			text = "Your input was not recognised as one of: yes/no/on/off"
		}
	} else {
		currStatus := db.ShouldDel(chat.Id)
		if currStatus {
			text = "Disabled Command deleting is *enabled*, disabled commands from users will be deleted!"
		} else {
			text = "Disabled Command deleting is *disabled*, disabled commands from users will *not* be deleted!"
		}
	}
	_, err := msg.Reply(b, text, helpers.Smarkdown())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
	To re-enable a command

# Connection - true, true

Only Admin can use this command to re-enable usage of a disabled command in the chat
*/
func (moduleStruct) enable(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	args := ctx.Args()[1:]

	if len(args) >= 1 {
		toEnable := make([]string, 0)

		for _, i := range args {
			i = strings.ToLower(i)
			if string_handling.FindInStringSlice(misc.DisableCmds, i) {
				toEnable = append(toEnable, i)
				_, err := msg.Reply(b, fmt.Sprintf("Re-Enabled the use of the following in this chat:"+
					"%s",
					strings.Join(toEnable, "\n - ")),
					helpers.Smarkdown())
				if err != nil {
					log.Error(err)
					return err
				}
			} else {
				_, err := msg.Reply(b,
					fmt.Sprintf("Unknown command to Re-Enable:\n-%s\nCheck /disableable!", i), nil)
				if err != nil {
					log.Error(err)
					return err
				}
			}
		}
		// finally disable all cmds
		for _, i := range toEnable {
			db.EnableCMD(chat.Id, i)
		}

	} else {
		_, err := msg.Reply(b, "You haven't specified a command to disable.", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return ext.EndGroups
}

func LoadDisabling(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(disablingModule.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("disable", disablingModule.disable))
	dispatcher.AddHandler(handlers.NewCommand("disableable", disablingModule.disableable))
	dispatcher.AddHandler(handlers.NewCommand("disabled", disablingModule.disabled))
	misc.AddCmdToDisableable("disabled")
	dispatcher.AddHandler(handlers.NewCommand("disabledel", disablingModule.disabledel))
	dispatcher.AddHandler(handlers.NewCommand("enable", disablingModule.enable))
}
