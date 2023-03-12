package helpers

import (
	"fmt"
	"html"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"github.com/Divkix/Alita_Robot/alita/utils/chat_status"
)

const MaxMessageLength int = 4096

func SplitMessage(msg string) []string {
	if len(msg) > MaxMessageLength {
		tmp := make([]string, 1)
		tmp[0] = msg
		return tmp
	} else {
		lines := strings.Split(msg, "\n")
		smallMsg := ""
		result := make([]string, 0)
		for _, line := range lines {
			if len(smallMsg)+len(line) < MaxMessageLength {
				smallMsg += line + "\n"
			} else {
				result = append(result, smallMsg)
				smallMsg = line + "\n"
			}
		}
		result = append(result, smallMsg)
		return result
	}
}

func MentionHtml(userId int64, name string) string {
	return MentionUrl(fmt.Sprintf("tg://user?id=%d", userId), name)
}

func MentionUrl(url, name string) string {
	return fmt.Sprintf("<a href=\"%s\">%s</a>", url, html.EscapeString(name))
}

func GetFullName(FirstName, LastName string) string {
	var name string
	if LastName != "" {
		name = FirstName + " " + LastName
	} else {
		name = FirstName
	}
	return name
}

func InitButtons(b *gotgbot.Bot, chatId, userId int64) gotgbot.InlineKeyboardMarkup {
	var connButtons [][]gotgbot.InlineKeyboardButton
	if chat_status.IsUserAdmin(b, chatId, userId) {
		connButtons = [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "Admin commands",
					CallbackData: "connbtns.Admin",
				},
			},
			{
				{
					Text:         "User commands",
					CallbackData: "connbtns.User",
				},
			},
		}
	} else {
		connButtons = [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "User commands",
					CallbackData: "connbtns.User",
				},
			},
		}
	}
	connKeyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: connButtons}
	return connKeyboard
}

// GetMessageLinkFromMessageId Gets the message link via chat Id and message Id
// maybe replace in future by msg.GetLink()
func GetMessageLinkFromMessageId(chat *gotgbot.Chat, messageId int64) (messageLink string) {
	messageLink = "https://t.me/"
	chatIdStr := fmt.Sprint(chat.Id)
	if chat.Username == "" {
		var linkId string
		if strings.HasPrefix(chatIdStr, "-100") {
			linkId = strings.ReplaceAll(chatIdStr, "-100", "")
		} else if strings.HasPrefix(chatIdStr, "-") && !strings.HasPrefix(chatIdStr, "-100") {
			// this is for non-supergroups
			linkId = strings.ReplaceAll(chatIdStr, "-", "")
		}
		messageLink += fmt.Sprintf("c/%s/%d", linkId, messageId)
	} else {
		messageLink += fmt.Sprintf("%s/%d", chat.Username, messageId)
	}
	return
}
