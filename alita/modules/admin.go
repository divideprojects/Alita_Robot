package modules

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/divideprojects/Alita_Robot/alita/utils/debug_bot"
	"github.com/divideprojects/Alita_Robot/alita/utils/decorators/misc"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
	"github.com/divideprojects/Alita_Robot/alita/utils/permissions"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"

	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
	"github.com/divideprojects/Alita_Robot/alita/utils/extraction"
)

var adminModule = moduleStruct{
	moduleName: "Admin",
	cfg:        nil, // will be set during LoadAdmin
}

/*
	Used to list all the admin in a group

Connection - false, false
*/
/*
adminlist lists all the admins in a group chat.

It checks for required permissions, retrieves the admin list (using cache if available), and formats a message listing all non-bot, non-anonymous admins. It also indicates whether the data is cached or up-to-date.

Connection: false, false
*/
func (m moduleStruct) adminlist(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	cached := true

	// if command is disabled, return
	if chat_status.CheckDisabledCmd(b, msg, "adminlist") {
		return ext.EndGroups
	}

	tr := i18n.New(db.GetLanguage(ctx))

	// permission checks
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}

	text := fmt.Sprintf(tr.GetString("strings."+m.moduleName+".adminlist"), chat.Title)

	adminsAvail, admins := cache.GetAdminCacheList(chat.Id)
	if !adminsAvail {
		admins = cache.LoadAdminCache(b, chat.Id)
		cached = false
	}

	for i := range admins.UserInfo {
		admin := &admins.UserInfo[i]
		user := admin.User
		if user.IsBot || admin.IsAnonymous {
			// don't list bots and anonymous admins
			continue
		}
		if user.Username != "" {
			text += fmt.Sprintf("\n- @%s", user.Username)
		} else {
			text += fmt.Sprintf("\n- %s", helpers.MentionHtml(user.Id, user.FirstName))
		}
	}
	if !cached {
		text += "\n\nNote: These are up-to-date values"
	} else {
		text += "\n\nNote: These values are cached and may not be up-to-date"
	}
	_, err := msg.Reply(b, text, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/* Used to Demote a member in chat

connection = true, true

Bot can only Demote people it promoted! */

/*
demote removes admin privileges from a user in the chat.

Performs permission checks, extracts the target user, and demotes them if possible. Only users promoted by the bot can be demoted. Handles edge cases such as anonymous users, the bot itself, and chat owners.

Connection: true, true
*/
func (m moduleStruct) demote(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	tr := i18n.New(db.GetLanguage(ctx))

	// Use helper for permission checks and user extraction
	userId, _, ok := permissions.PerformCommonPromotionChecks(b, ctx, permissions.CommonPromotionPerms)
	if !ok {
		return ext.EndGroups
	}

	if chat_status.RequireUserOwner(b, ctx, nil, userId, true) {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".demote.is_owner"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}
	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".demote.is_bot_itself"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}
	if !chat_status.IsUserAdmin(b, chat.Id, userId) {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".demote.is_admin"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}

	bb, err := chat.PromoteMember(b,
		userId,
		&gotgbot.PromoteChatMemberOpts{
			CanDeleteMessages:   false,
			CanRestrictMembers:  false,
			CanChangeInfo:       false,
			CanInviteUsers:      false,
			CanPinMessages:      false,
			CanManageVideoChats: false,
			CanManageTopics:     false,
		},
	)

	if err != nil || !bb {
		log.Error(err)
		_, err = msg.Reply(b,
			tr.GetString("strings."+m.moduleName+".errors.err_cannot_demote"),
			nil,
		)
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}

	userMember, err := chat.GetMember(b, userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}
	mem := userMember.MergeChatMember().User
	_, err = msg.Reply(b,
		fmt.Sprintf(tr.GetString("strings."+m.moduleName+".demote.success_demote"), helpers.MentionHtml(mem.Id, mem.FirstName)),
		helpers.Shtml(),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/* Used to Promote a member in chat

connection = true, true

Bot will give promoted user permissions of bot*/

/*
promote grants admin privileges to a user in the chat.

Checks permissions, extracts the target user and optional custom title, and promotes them with the bot's own permissions. Handles edge cases such as anonymous users, the bot itself, and chat owners. Truncates custom titles to 16 characters as required by Telegram.

Connection: true, true
*/
func (m moduleStruct) promote(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	tr := i18n.New(db.GetLanguage(ctx))

	extraText := ""

	// Use helper for permission checks and user extraction
	userId, customTitle, ok := permissions.PerformCommonPromotionChecks(b, ctx, permissions.CommonPromotionPerms)
	if !ok {
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".promote.is_bot_itself"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}

	// checks if user being promoted is already admin or owner
	if chat_status.RequireUserOwner(b, ctx, nil, userId, true) {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".promote.is_owner"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}
	if chat_status.IsUserAdmin(b, chat.Id, userId) {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".promote.is_admin"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}

	userMember, err := chat.GetMember(b, userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	botMember, err := chat.GetMember(b, b.Id, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	// makes code short
	bMem := botMember.MergeChatMember()
	pMem := userMember.MergeChatMember()

	teamMem := db.GetTeamMemInfo(user.Id)
	teamMemInfo := teamMem.Sudo || teamMem.Dev
	isPromoterOwner := chat_status.RequireUserOwner(b, ctx, nil, user.Id, true)

	checkCommonPerms := isPromoterOwner || teamMemInfo

	_, err = chat.PromoteMember(b,
		userId,
		&gotgbot.PromoteChatMemberOpts{
			CanPostMessages:     bMem.CanPostMessages && (pMem.CanPostMessages || checkCommonPerms),
			CanDeleteMessages:   bMem.CanDeleteMessages && (pMem.CanDeleteMessages || checkCommonPerms),
			CanRestrictMembers:  bMem.CanRestrictMembers && (pMem.CanRestrictMembers || checkCommonPerms),
			CanChangeInfo:       bMem.CanChangeInfo && (pMem.CanChangeInfo || checkCommonPerms),
			CanInviteUsers:      bMem.CanInviteUsers && (pMem.CanInviteUsers || checkCommonPerms),
			CanPinMessages:      bMem.CanPinMessages && (pMem.CanPinMessages || checkCommonPerms),
			CanManageVideoChats: bMem.CanManageVideoChats && (pMem.CanManageVideoChats || checkCommonPerms),
			CanManageChat:       bMem.CanManageChat && (pMem.CanManageChat || checkCommonPerms),
			CanManageTopics:     bMem.CanManageTopics && (pMem.CanManageTopics || checkCommonPerms),
		},
	)
	if err != nil {
		_, _ = msg.Reply(b, tr.GetString("strings."+m.moduleName+".errors.err_cannot_promote"), helpers.Shtml())
		return err
	}

	if len(customTitle) > 16 {
		// trim title to 16 characters (telegram restriction)
		extraText += fmt.Sprintf(tr.GetString("strings."+m.moduleName+".promote.admin_title_truncated"), len(customTitle))
		customTitle = customTitle[0:16]
	}

	// set the custom title
	if customTitle != "" {
		_, err = chat.SetAdministratorCustomTitle(
			b,
			userId,
			customTitle,
			nil,
		)
		if err != nil {
			_, err = msg.Reply(b,
				tr.GetString("strings."+m.moduleName+".errors.err_set_title"),
				nil,
			)
			if err != nil {
				log.Error(err)
			}
			return ext.EndGroups
		}
	}
	mem := userMember.MergeChatMember().User
	_, err = msg.Reply(b,
		fmt.Sprintf(tr.GetString("strings."+m.moduleName+".promote.success_promote"), helpers.MentionHtml(mem.Id, mem.FirstName))+extraText,
		helpers.Shtml(),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
getinvitelink retrieves the invite link for the current chat.

Checks permissions and returns the chat's username as an invite link if available, otherwise fetches the invite link from the API.
*/
func (moduleStruct) getinvitelink(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage

	// permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.Caninvite(b, ctx, nil, msg, false) {
		return ext.EndGroups
	}
	tr := i18n.New(db.GetLanguage(ctx))
	if chat.Username != "" {
		_, _ = msg.Reply(b, fmt.Sprintf(tr.GetString("strings.Admin.here_is_the_invite_link_of_this_chat_percent"), chat.Username), nil)
	} else {
		nchat, err := b.GetChat(chat.Id, nil)
		if err != nil {
			_, _ = msg.Reply(b, err.Error(), nil)
			return ext.EndGroups
		}
		_, _ = msg.Reply(b, fmt.Sprintf(tr.GetString("strings.Admin.here_is_the_invite_link_of_this_chat_percent"), nchat.InviteLink), nil)
	}
	return ext.EndGroups
}

/*
Sets a custom title for an admin.
Only works with admins whom bot has promoted.*/

/*
setTitle sets a custom admin title for a user.

Only works for admins promoted by the bot. Checks permissions, extracts the target user and title, and sets the custom title (truncated to 16 characters if necessary).
*/
func (m moduleStruct) setTitle(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	tr := i18n.New(db.GetLanguage(ctx))

	// permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserPromote(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotPromote(b, ctx, nil, false) {
		return ext.EndGroups
	}

	userId, customTitle := extraction.ExtractUserAndText(b, ctx)
	if userId == -1 {
		return ext.EndGroups
	} else if strings.HasPrefix(fmt.Sprint(userId), "-100") {
		_, err := msg.Reply(b, tr.GetString("Warns.errors.anon_user"), nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		_, err := msg.Reply(b, tr.GetString("strings.CommonStrings.errors.no_user_specified"),
			helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if chat_status.RequireUserOwner(b, ctx, nil, userId, true) {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".title.is_owner"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}
	if !chat_status.IsUserAdmin(b, chat.Id, userId) {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".title.is_admin"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if userId == b.Id {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".title.is_bot_itself"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}

	// for managing custom title
	if customTitle == "" {
		_, err := msg.Reply(b, tr.GetString("strings."+m.moduleName+".errors.title_empty"), helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if len(customTitle) > 16 {
		// trim title to 16 characters (telegram restriction)
		customTitle = customTitle[0:16]
	}

	_, err := chat.SetAdministratorCustomTitle(b,
		userId,
		customTitle,
		nil,
	)
	if err != nil {
		log.Error(err)
		_, _ = msg.Reply(b, tr.GetString("strings."+m.moduleName+".errors.err_set_title"), helpers.Shtml())
		return err
	}

	userMember, err := chat.GetMember(b, userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	mem := userMember.MergeChatMember()

	_, err = msg.Reply(b,
		fmt.Sprintf(tr.GetString("strings."+m.moduleName+".title.success_set"), mem.User.FirstName, mem.CustomTitle),
		helpers.Shtml(),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
anonAdmin toggles or displays the anonymous admin mode for the chat.

Allows the chat owner to enable or disable anonymous admin mode. If called with no arguments, displays the current status.
*/
func (m moduleStruct) anonAdmin(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	args := ctx.Args()

	tr := i18n.New(db.GetLanguage(ctx))
	var text string

	// permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}

	adminSettings := db.GetAdminSettings(chat.Id)

	if len(args) == 1 {
		if adminSettings.AnonAdmin {
			text = fmt.Sprintf(tr.GetString("strings."+m.moduleName+".anon_admin.enabled"), chat.Title)
		} else {
			text = fmt.Sprintf(tr.GetString("strings."+m.moduleName+".anon_admin.disabled"), chat.Title)
		}
	} else {
		// only need owner if you want to change value
		if !chat_status.RequireUserOwner(b, ctx, nil, user.Id, false) {
			return ext.EndGroups
		}
		switch args[1] {
		case "on", "true", "yes":
			if adminSettings.AnonAdmin {
				text = tr.GetString("strings." + m.moduleName + ".anon_admin.already_enabled")
			} else {
				go db.SetAnonAdminMode(chat.Id, true)
				text = fmt.Sprintf(tr.GetString("strings."+m.moduleName+".anon_admin.enabled_now"), chat.Title)
			}
		case "off", "no", "false":
			if !adminSettings.AnonAdmin {
				text = tr.GetString("strings." + m.moduleName + ".anon_admin.already_disabled")
			} else {
				go db.SetAnonAdminMode(chat.Id, false)
				text = fmt.Sprintf(tr.GetString("strings."+m.moduleName+".anon_admin.disabled_now"), chat.Title)
			}
		default:
			text = tr.GetString("strings." + m.moduleName + ".anon_admin.invalid_arg")
		}
	}

	_, err := msg.Reply(b, text, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
adminCache reloads the admin cache for the current chat.

Only available to chat admins. Reloads the admin list from Telegram and updates the cache.
*/
func (moduleStruct) adminCache(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User

	var err error

	tr := i18n.New(db.GetLanguage(ctx))
	debug_bot.PrettyPrintStruct(tr)

	// permission checks
	userMember, _ := chat.GetMember(b, user.Id, nil)
	mem := userMember.MergeChatMember()
	if mem.Status == "member" {
		_, err = msg.Reply(b, tr.GetString("strings.Admin.you_need_to_be_admin_to_do_this"), nil)
		if err != nil {
			log.Error(err)
		}
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}

	cache.LoadAdminCache(b, chat.Id)

	k := tr.GetString("strings.CommonStrings.admin_cache.cache_reloaded")
	debug_bot.PrettyPrintStruct(k)
	_, err = msg.Reply(b, k, helpers.Shtml())
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

/*
LoadAdmin registers all admin-related command handlers with the dispatcher.

This function enables the admin module and adds handlers for admin commands such as promote, demote, adminlist, invitelink, title, anonadmin, and admincache.
*/
func LoadAdmin(dispatcher *ext.Dispatcher, cfg *config.Config) {
	// Store config in the module
	adminModule.cfg = cfg

	HelpModule.AbleMap.Store("Admin", true)

	dispatcher.AddHandler(handlers.NewCommand("admin", adminModule.promote))
	dispatcher.AddHandler(handlers.NewCommand("demote", adminModule.demote))
	dispatcher.AddHandler(handlers.NewCommand("invitelink", adminModule.getinvitelink))
	dispatcher.AddHandler(handlers.NewCommand("title", adminModule.setTitle))
	dispatcher.AddHandler(handlers.NewCommand("adminlist", adminModule.adminlist))
	misc.AddCmdToDisableable("adminlist")
	dispatcher.AddHandler(handlers.NewCommand("anonadmin", adminModule.anonAdmin))
	dispatcher.AddHandler(handlers.NewCommand("admincache", adminModule.adminCache))
	dispatcher.AddHandler(
		handlers.NewCommand(
			"clearadmincache",
			func(b *gotgbot.Bot, ctx *ext.Context) error {
				chat := ctx.EffectiveChat
				err := cache.Marshal.Delete(cache.Context, cache.AdminCache{ChatId: chat.Id})
				if err != nil {
					log.Error(err)
					return err
				}
				log.Info(fmt.Sprintf("Cleared admin cache for %d (%s)", chat.Id, chat.Title))
				return ext.EndGroups
			},
		),
	)
}
