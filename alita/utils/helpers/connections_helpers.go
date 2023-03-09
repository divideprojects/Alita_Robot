package helpers

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	log "github.com/sirupsen/logrus"

	"github.com/Divkix/Alita_Robot/alita/db"
	"github.com/Divkix/Alita_Robot/alita/i18n"
	"github.com/Divkix/Alita_Robot/alita/utils/chat_status"
	"github.com/Divkix/Alita_Robot/alita/utils/parsemode"
)

/*
	Used to return the chat to which user is connected

If user is connected to a chat, chat is returned else nil is returned
*/
func IsUserConnected(b *gotgbot.Bot, ctx *ext.Context, chatAdmin, botAdmin bool) (chat *gotgbot.Chat) {
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveUser
	tr := i18n.I18n{LangCode: db.GetLanguage(ctx)}
	var err error

	if ctx.Update.Message.Chat.Type == "private" {
		conn := db.Connection(user.Id)
		if conn.Connected && conn.ChatId != 0 {
			chat, err = b.GetChat(conn.ChatId, nil)
			if err != nil {
				log.Error(err)
				return nil
			}
		} else {
			_, err := msg.Reply(b,
				tr.GetString("strings.Connections.is_user_connected.need_group"),
				&gotgbot.SendMessageOpts{
					ReplyToMessageId:         msg.MessageId,
					AllowSendingWithoutReply: true,
				},
			)
			if err != nil {
				log.Error(err)
				return nil
			}

			return nil
		}
	} else {
		chat = ctx.EffectiveChat
	}
	if botAdmin {
		if !chat_status.IsUserAdmin(b, chat.Id, b.Id) {
			_, err := msg.Reply(b, tr.GetString("strings.Connections.is_user_connected.bot_not_admin"), parsemode.Shtml())
			if err != nil {
				log.Error(err)
				return nil
			}

			return nil
		}
	}
	if chatAdmin {
		if !chat_status.IsUserAdmin(b, chat.Id, user.Id) {
			_, err := msg.Reply(b, tr.GetString("strings.Connections.is_user_connected.user_not_admin"), parsemode.Shtml())
			if err != nil {
				log.Error(err)
				return nil
			}

			return nil
		}
	}
	ctx.EffectiveChat = chat
	return chat
}
