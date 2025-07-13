package extraction

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/google/uuid"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/error_handling"
)

var (
	// Pre-compiled regex patterns for better performance
	quotesPattern = regexp.MustCompile(`(?s)(\s+)?"(.*?)"\s?(.*)?`)
	wordPattern   = regexp.MustCompile(`(?s)(\s+)?([A-Za-z0-9-_+=}\][{;:'",<.>?/|*\\()]+)\s?(.*)?`)
)

/*
ExtractChat extracts the chat from the command arguments or message context.

If a chat ID or username is provided as an argument, it attempts to fetch the chat.
If not, it prompts the user for a chat ID. Returns the chat or nil on failure.
*/
func ExtractChat(b *gotgbot.Bot, ctx *ext.Context) *gotgbot.Chat {
	msg := ctx.EffectiveMessage
	args := ctx.Args()[1:]
	if len(args) != 0 {
		if _, err := strconv.Atoi(args[0]); err == nil {
			chatId, _ := strconv.Atoi(args[0])
			chat, err := b.GetChat(int64(chatId), nil)
			if err != nil {
				_, err := msg.Reply(b, "failed to connect to chat: failed to get chat: unable to getChat: Bad Request: chat not found", nil)
				if err != nil {
					log.Error(err)
					return nil
				}
				return nil
			}
			_chat := chat.ToChat() // need to convert to Chat type
			return &_chat
		} else {
			chat, err := chat_status.GetChat(b, args[0])
			if err != nil {
				_, err := msg.Reply(b, "failed to connect to chat: failed to get chat: unable to getChat: Bad Request: chat not found", nil)
				if err != nil {
					log.Error(err)
					return nil
				}
				return nil
			}
			return chat
		}
	}
	tr := i18n.New(db.GetLanguage(ctx))
	_, err := msg.Reply(b, tr.GetString("strings.Connections.errors.no_chat_id"), nil)
	if err != nil {
		log.Error(err)
		return nil
	}
	return nil
}

/*
ExtractUser extracts the user ID from the message context.

Returns the user ID if found, otherwise returns -1.
*/
func ExtractUser(b *gotgbot.Bot, ctx *ext.Context) int64 {
	userId, _ := ExtractUserAndText(b, ctx)
	return userId
}

/*
ExtractUserAndText extracts the user ID and associated text from the message.

Handles various input formats, including replies, mentions, and direct IDs.
Returns the user ID and the extracted text, or -1 and an empty string on failure.
*/
func ExtractUserAndText(b *gotgbot.Bot, ctx *ext.Context) (int64, string) {
	msg := ctx.EffectiveMessage
	args := ctx.Args()
	prevMessage := msg.ReplyToMessage

	splitText := strings.SplitN(msg.Text, " ", 2)

	if len(splitText) < 2 {
		return IdFromReply(msg)
	}

	textToParse := splitText[1]

	// func used to trim newlines from the text, fixes the pasring issues of '\n' before and after text
	trimTextNewline := func(str string) string {
		return strings.Trim(str, "\n")
	}

	text := ""

	var userId int64
	accepted := make(map[string]struct{})
	accepted["text_mention"] = struct{}{}

	entities := msg.ParseEntityTypes(accepted)

	var ent *gotgbot.ParsedMessageEntity
	isId := false
	if len(entities) > 0 {
		ent = &entities[0]
	} else {
		ent = nil
	}

	// only parse if the entity is a text mention
	if entities != nil && ent != nil && int(ent.Offset) == (len(msg.Text)-len(textToParse)) {
		ent = &entities[0]
		userId = ent.User.Id
		text = msg.Text[ent.Offset+ent.Length:]
	} else if len(args) >= 1 && args[1][0] == '@' {
		user := args[1]
		userId = GetUserId(user)
		if userId == 0 {
			_, err := msg.Reply(b, "Could not find a user by this name; are you sure I've seen them before?", nil)
			error_handling.HandleErr(err)
			return -1, ""
		} else {
			res := strings.SplitN(msg.Text, " ", 3)
			if len(res) >= 3 {
				text = res[2]
			}
		}
	} else if len(args) >= 1 {
		isId = true
		if !strings.HasPrefix(args[1], "-100") {
			for _, arg := range args[1] {
				if unicode.IsDigit(arg) {
					continue
				}
				isId = false
				break
			}
		}
		if isId {
			userId, _ = strconv.ParseInt(args[1], 10, 64)
			res := strings.SplitN(msg.Text, " ", 3)
			if len(res) >= 3 {
				text = res[2]
			}
		}
	}
	if !isId && prevMessage != nil {
		_, parseErr := uuid.Parse(args[1])
		userId, text = IdFromReply(msg)
		if parseErr == nil {
			return userId, trimTextNewline(text)
		}
	} else if !isId {
		_, parseErr := uuid.Parse(args[1])
		if parseErr == nil {
			return userId, trimTextNewline(text)
		}
	}

	_, _, found := GetUserInfo(userId)
	if !found {
		_, err := msg.Reply(b, "Failed to get user: unable to getChatMember: Bad Request: chat not found", nil)
		error_handling.HandleErr(err)
		return -1, ""
	}

	return userId, trimTextNewline(text)
}

/*
GetUserId returns the user or channel ID for a given username.

Strips the '@' prefix and searches both users and channels in the database.
Returns 0 if not found.
*/
func GetUserId(username string) int64 {
	if len(username) <= 5 {
		return 0
	}

	// remove '@' from the username
	username = strings.ReplaceAll(username, "@", "")

	user := db.GetUserIdByUserName(username)
	if user != 0 {
		return user
	}

	channel := db.GetChannelIdByUserName(username)
	if channel != 0 {
		return channel
	}

	return 0
}

/*
GetUserInfo retrieves the username and name for a given user or channel ID.

Returns the username, name, and a boolean indicating if the user/channel was found.
*/
func GetUserInfo(userId int64) (username, name string, found bool) {
	username, name, found = db.GetUserInfoById(userId)
	if found {
		return username, name, found
	}

	username, name, found = db.GetChannelInfoById(userId)
	if found {
		return username, name, found
	}

	return "", "", false
}

/*
IdFromReply extracts the user ID and text from a replied-to message.

Returns the user ID and the text following the command, or 0 and an empty string if not a reply.
*/
func IdFromReply(m *gotgbot.Message) (int64, string) {
	prevMessage := m.ReplyToMessage

	var userId int64

	if prevMessage == nil {
		return 0, ""
	}

	// get's the Id for both user and channel
	userId = prevMessage.GetSender().Id()

	res := strings.SplitN(m.Text, " ", 2)
	if len(res) < 2 {
		return userId, ""
	}
	return userId, res[1]
}

/*
ExtractQuotes extracts text between quotes or the next word from a sentence.

If matchQuotes is true, extracts quoted text; if matchWord is true, extracts the next word.
Returns the extracted text and any remaining text after the match.
*/
func ExtractQuotes(sentence string, matchQuotes, matchWord bool) (inQuotes, afterWord string) {
	// if first character starts with '""' and matchQutes is true
	if sentence[0] == '"' && matchQuotes {
		if quotesPattern.MatchString(sentence) {
			pat := quotesPattern.FindStringSubmatch(sentence)
			// pat[0] would be the whole matched string
			// pat[1] is the spaces
			inQuotes, afterWord = pat[2], pat[3]
			return
		}
	} else if matchWord {
		if wordPattern.MatchString(sentence) {
			pat := wordPattern.FindStringSubmatch(sentence)
			// pat[0] would be the whole matched string
			// pat[1] is the spaces
			inQuotes, afterWord = pat[2], pat[3]
			return
		}
	}

	return
}

/*
ExtractTime parses a time duration and reason from a string.

Supports 'm', 'h', 'd', and 'w' suffixes for minutes, hours, days, and weeks.
Returns the calculated Unix timestamp, a human-readable string, and the reason.
Returns -1 and an error message if parsing fails or the duration exceeds one year.
*/
func ExtractTime(b *gotgbot.Bot, ctx *ext.Context, inputVal string) (banTime int64, timeStr, reason string) {
	msg := ctx.EffectiveMessage
	timeNow := time.Now().Unix()
	yearTime := timeNow + int64(365*24*60*60)

	args := strings.Fields(inputVal)
	timeVal := args[0] // first word will be the time specification
	if len(args) >= 2 {
		reason = strings.Join(args[1:], " ")
	}

	if strings.ContainsAny(timeVal, "m & d & h & w") {
		t := timeVal[:len(timeVal)-1]
		timeNum, err := strconv.Atoi(t)
		if err != nil {
			_, err := msg.Reply(b, "Invalid time amount specified.", nil)
			error_handling.HandleErr(err)
			return -1, "", ""
		}

		switch string(timeVal[len(timeVal)-1]) {
		case "m":
			banTime = timeNow + int64(timeNum*60)
			timeStr = fmt.Sprintf("%d minutes", timeNum)
		case "h":
			banTime = timeNow + int64(timeNum*60*60)
			timeStr = fmt.Sprintf("%d hours", timeNum)
		case "d":
			banTime = timeNow + int64(timeNum*24*60*60)
			timeStr = fmt.Sprintf("%d days", timeNum)
		case "w":
			banTime = timeNow + int64(timeNum*7*24*60*60)
			timeStr = fmt.Sprintf("%d weeks", timeNum)
		default:
			return -1, "", ""
		}

		if banTime >= yearTime {
			_, err := msg.Reply(b, "Cannot set time to more than 1 year!", nil)
			error_handling.HandleErr(err)
			return -1, "", ""
		}

		return banTime, timeStr, reason
	} else {
		_, err := msg.Reply(b, fmt.Sprintf("Invalid time type specified. Expected m, h, d or w got: %s", timeVal), nil)
		error_handling.HandleErr(err)
		return -1, "", ""
	}
}
