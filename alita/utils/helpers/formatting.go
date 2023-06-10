package helpers

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/db"
)

func FormattingReplacer(b *gotgbot.Bot, chat *gotgbot.Chat, user *gotgbot.User, oldMsg string, buttons []db.Button) (res string, btns []db.Button) {
	var (
		firstName     string
		fullName      string
		username      string
		rulesBtnRegex = `(?s){rules(:(same|up))?}`
	)

	firstName = user.FirstName
	if len(user.FirstName) <= 0 {
		firstName = "PersonWithNoName"
	}

	if user.LastName != "" {
		fullName = firstName + " " + user.LastName
	} else {
		fullName = firstName
	}
	count, _ := chat.GetMemberCount(b, nil)
	mention := MentionHtml(user.Id, firstName)

	if user.Username != "" {
		username = "@" + html.EscapeString(user.Username)
	} else {
		username = mention
	}
	r := strings.NewReplacer(
		"{first}", html.EscapeString(firstName),
		"{last}", html.EscapeString(user.LastName),
		"{fullname}", html.EscapeString(fullName),
		"{username}", username,
		"{mention}", mention,
		"{count}", strconv.Itoa(int(count)),
		"{chatname}", html.EscapeString(chat.Title),
		"{id}", strconv.Itoa(int(user.Id)),
	)
	res = r.Replace(oldMsg)
	btns = buttons // copies the buttons over to format rules btn

	rulesDb := db.GetChatRulesInfo(chat.Id)
	rulesBtnText := rulesDb.RulesBtn
	if rulesBtnText == "" {
		rulesBtnText = "Rules"
	}

	// only add rules btn when rules are added in chat
	if rulesDb.Rules != "" {
		pattern, err := regexp.Compile(rulesBtnRegex)
		if err != nil {
			log.Error(err)
		}
		if pattern.MatchString(res) {
			response := pattern.FindStringSubmatch(res)

			sameline := false
			if response[2] == "same" {
				sameline = true
			}

			rulesButton := db.Button{
				Name:     rulesBtnText,
				Url:      fmt.Sprintf("https://t.me/%s?start=rules_%d", b.Username, chat.Id),
				SameLine: sameline,
			}

			if response[2] == "up" {
				// this adds the button on top of all buttons
				btns = []db.Button{rulesButton}
				btns = append(btns, buttons...)
			} else {
				// this adds the button to bottom (default behaviour)
				btns = buttons
				btns = append(btns, rulesButton)

			}
			res = pattern.ReplaceAllString(res, "")
		}
	}

	return res, btns
}
