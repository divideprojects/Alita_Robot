package modules

import (
	"fmt"
	"html"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// module struct for all modules
type moduleStruct struct {
	moduleName          string
	handlerGroup        int
	permHandlerGroup    int
	restrHandlerGroup   int
	defaultRulesBtn     string
	overwriteFiltersMap map[string]overwriteFilter
	overwriteNotesMap   map[string]overwriteNote
	antiSpam            map[int64]*antiSpamInfo
	AbleMap             moduleEnabled
	AltHelpOptions      map[string][]string
	helpableKb          map[string][][]gotgbot.InlineKeyboardButton
	cfg                 *config.Config
}

// struct for filters module
type overwriteFilter struct {
	filterWord string
	text       string
	fileid     string
	buttons    []db.Button
	dataType   int
}

// struct for notes module
type overwriteNote struct {
	noteWord    string
	text        string
	fileId      string
	buttons     []db.Button
	dataType    int
	pvtOnly     bool
	grpOnly     bool
	adminOnly   bool
	webPrev     bool
	isProtected bool
	noNotif     bool
}

// struct for antiSpam module - antiSpamInfo
type antiSpamInfo struct {
	Levels []antiSpamLevel
}

// struct for antiSpam module - antiSpamLevel
type antiSpamLevel struct {
	Count    int
	Limit    int
	CurrTime time.Duration
	Expiry   time.Duration
	Spammed  bool
}

// helper functions for help module

// This var is used to add the back button to the help menu
// i.e. where modules are shown
var markup gotgbot.InlineKeyboardMarkup

func listModules() []string {
	// sort the modules alphabetically
	modules := HelpModule.AbleMap.LoadModules()
	sort.Strings(modules) // Sort the modules
	return modules
}

// New menu, used for building help menu in bot!
func initHelpButtons() {
	var kb []gotgbot.InlineKeyboardButton

	for _, i := range listModules() {
		kb = append(kb, gotgbot.InlineKeyboardButton{Text: i, CallbackData: fmt.Sprintf("helpq.%s", i)})
	}
	zb := helpers.ChunkKeyboardSlices(kb, 3)
	zb = append(zb, []gotgbot.InlineKeyboardButton{{Text: "« Back", CallbackData: "helpq.BackStart"}})
	markup = gotgbot.InlineKeyboardMarkup{InlineKeyboard: zb}
}

func getModuleHelpAndKb(module, lang string) (helpText string, replyMarkup gotgbot.InlineKeyboardMarkup) {
	ModName := cases.Title(language.English).String(module)
	helpText = fmt.Sprintf("Here is the help for the *%s* module:\n\n", ModName) +
		i18n.New(lang).GetString(fmt.Sprintf("strings.%s.help_msg", ModName))

	tr := i18n.New(lang)
	backBtnSuffix := getBackBtnSuffix(tr)
	replyMarkup = gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: append(
			HelpModule.helpableKb[ModName],
			backBtnSuffix,
		),
	}
	return
}

func sendHelpkb(b *gotgbot.Bot, ctx *ext.Context, module string) (msg *gotgbot.Message, err error) {
	module = strings.ToLower(module)
	if module == "help" {
		_, err = b.SendMessage(
			ctx.EffectiveMessage.Chat.Id,
			fmt.Sprintf(
				mainhlp,
				html.EscapeString(ctx.EffectiveMessage.From.FirstName),
			),
			&gotgbot.SendMessageOpts{
				ParseMode:   gotgbot.ParseModeHTML,
				ReplyMarkup: &markup,
			},
		)
		return
	}
	helpText, replyMarkup, _parsemode := getHelpTextAndMarkup(ctx, module)

	msg, err = b.SendMessage(
		ctx.EffectiveChat.Id,
		helpText,
		&gotgbot.SendMessageOpts{
			ParseMode:   _parsemode,
			ReplyMarkup: replyMarkup,
		},
	)
	return
}

func getModuleNameFromAltName(altName string) string {
	for _, modName := range listModules() {
		altNames := append(i18n.New("config").GetStringSlice(fmt.Sprintf("alt_names.%s", modName)), strings.ToLower(modName))
		for _, altNameInSlice := range altNames {
			if altNameInSlice == altName {
				return modName
			}
		}
	}
	return ""
}

func startHelpPrefixHandler(b *gotgbot.Bot, ctx *ext.Context, user *gotgbot.User, arg string) error {
	msg := ctx.EffectiveMessage
	chat := ctx.EffectiveChat
	tr := i18n.New(db.GetLanguage(ctx))

	if strings.HasPrefix(arg, "help_") {
		helpModule := strings.Split(arg, "_")[1]
		_, err := sendHelpkb(b, ctx, helpModule)
		if err != nil {
			log.Errorf("[Start]: %v", err)
			return err
		}
	} else if strings.HasPrefix(arg, "connect_") {
		chatID, _ := strconv.Atoi(strings.Split(arg, "_")[1])
		cochat, _ := b.GetChat(int64(chatID), nil)
		go db.ConnectId(user.Id, cochat.Id)

		Text := fmt.Sprintf("You have been connected to %s!", cochat.Title)
		connKeyboard := helpers.InitButtons(b, cochat.Id, user.Id)

		_, err := ctx.EffectiveMessage.Reply(b, Text,
			&gotgbot.SendMessageOpts{
				ReplyMarkup: connKeyboard,
			},
		)
		if err != nil {
			log.Error(err)
			return err
		}
	} else if strings.HasPrefix(arg, "rules_") {
		chatID, _ := strconv.Atoi(strings.Split(arg, "_")[1])
		chatinfo, _ := b.GetChat(int64(chatID), nil)
		rulesrc := db.GetChatRulesInfo(int64(chatID))

		if rulesrc.Rules == "" {
			noRulesMsg, noRulesErr := tr.GetStringWithError("strings.Helpers.this_chat_does_not_have_any_rules")
			if noRulesErr != nil {
				log.Errorf("[helpers] missing translation for Helpers.this_chat_does_not_have_any_rules: %v", noRulesErr)
				noRulesMsg = "This chat doesn't have any rules set."
			}
			_, err := msg.Reply(b, noRulesMsg, helpers.Shtml())
			if err != nil {
				log.Error(err)
				return err
			}
			return ext.EndGroups
		}

		Text := fmt.Sprintf("Rules for <b>%s</b>:\n\n%s", chatinfo.Title, rulesrc.Rules)
		_, err := msg.Reply(b, Text, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
	} else if strings.HasPrefix(arg, "note") {
		nArgs := strings.SplitN(arg, "_", 3)
		chatID, _ := strconv.Atoi(nArgs[1])
		chatinfo, _ := b.GetChat(int64(chatID), nil)

		if strings.HasPrefix(arg, "notes_") {
			// check if feth admin notes or not
			admin := chat_status.IsUserAdmin(b, int64(chatID), user.Id)
			noteKeys := db.GetNotesList(chatinfo.Id, admin)
			info := "There are no notes in this chat!"
			if len(noteKeys) > 0 {
				info = "These are the current notes in this Chat:\n"
				for _, note := range noteKeys {
					info += fmt.Sprintf(" - <a href='https://t.me/%s?start=note_%d_%s'>%s</a>\n", b.Username, int64(chatID), note, note)
				}
			}

			_, err := msg.Reply(b, info, helpers.Shtml())
			if err != nil {
				log.Error(err)
				return err
			}
		} else if strings.HasPrefix(arg, "note_") {
			noteName := strings.ToLower(nArgs[2])
			noteData := db.GetNote(chatinfo.Id, noteName)
			if noteData == nil {
				noteNotExistMsg, noteNotExistErr := tr.GetStringWithError("strings.Helpers.this_note_does_not_exist")
				if noteNotExistErr != nil {
					log.Errorf("[helpers] missing translation for Helpers.this_note_does_not_exist: %v", noteNotExistErr)
					noteNotExistMsg = "This note doesn't exist."
				}
				_, err := msg.Reply(b, noteNotExistMsg, helpers.Shtml())
				if err != nil {
					log.Error(err)
					return err
				}
				return ext.EndGroups
			}
			if noteData.AdminOnly {
				if !chat_status.IsUserAdmin(b, int64(chatID), user.Id) {
					adminOnlyMsg, adminOnlyErr := tr.GetStringWithError("strings.Notes.errors.admin_only")
					if adminOnlyErr != nil {
						log.Errorf("[helpers] missing translation for Notes.errors.admin_only: %v", adminOnlyErr)
						adminOnlyMsg = "This note is only available to admins."
					}
					_, err := msg.Reply(b, adminOnlyMsg, helpers.Shtml())
					if err != nil {
						log.Error(err)
						return err
					}
					return ext.ContinueGroups
				}
			}
			_chat := chatinfo.ToChat() // need to convert to chat
			_, err := helpers.SendNote(b, &_chat, ctx, noteData, msg.MessageId)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	} else if arg == "about" {
		tr := i18n.New(db.GetLanguage(ctx))
		aboutKb := getAboutKb(tr)
		_, err := b.SendMessage(chat.Id,
			aboutText,
			&gotgbot.SendMessageOpts{
				ParseMode: "Markdown",
				LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
					IsDisabled: true,
				},
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId:                msg.MessageId,
					AllowSendingWithoutReply: true,
				},
				ReplyMarkup: &aboutKb,
			},
		)
		if err != nil {
			log.Error(err)
			return err
		}
	} else {
		// This sends the normal help block
		tr := i18n.New(db.GetLanguage(ctx))
		startMarkup := getStartMarkup(tr)
		_, err := b.SendMessage(chat.Id,
			startHelp,
			&gotgbot.SendMessageOpts{
				ParseMode: gotgbot.ParseModeHTML,
				LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
					IsDisabled: true,
				},
				ReplyMarkup: &startMarkup,
			},
		)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.EndGroups
}

func getAltNamesOfModule(moduleName string) []string {
	return append(i18n.New("config").GetStringSlice(fmt.Sprintf("alt_names.%s", moduleName)), strings.ToLower(moduleName))
}

func getHelpTextAndMarkup(ctx *ext.Context, module string) (helpText string, kbmarkup gotgbot.InlineKeyboardMarkup, _parsemode string) {
	var moduleName string
	userOrGroupLanguage := db.GetLanguage(ctx)

	for _, ModName := range listModules() {
		// add key as well to this array
		altnames := getAltNamesOfModule(ModName)

		if string_handling.FindInStringSlice(altnames, module) {
			moduleName = ModName
			break
		}
	}

	// compare and check if module name is not empty
	if moduleName != "" {
		_parsemode = helpers.Markdown
		helpText, kbmarkup = getModuleHelpAndKb(moduleName, userOrGroupLanguage)
	} else {
		_parsemode = gotgbot.ParseModeHTML
		helpText = fmt.Sprintf(mainhlp, html.EscapeString(ctx.EffectiveUser.FirstName))
		kbmarkup = markup
	}

	return
}
