package modules

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
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
	moduleName: autoModuleName(),
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

	adminlistMsg, adminlistErr := tr.GetStringWithError("strings.admin.adminlist")
	if adminlistErr != nil {
		log.Errorf("[admin] missing translation for adminlist: %v", adminlistErr)
		adminlistMsg = "Admins in <b>%s</b>:"
	}
	text := fmt.Sprintf(adminlistMsg, chat.Title)

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
		demoteOwnerMsg, demoteOwnerErr := tr.GetStringWithError("strings.admin.demote.is_owner")
		if demoteOwnerErr != nil {
			log.Errorf("[admin] missing translation for demote.is_owner: %v", demoteOwnerErr)
			demoteOwnerMsg = "This person created this chat, and how would I demote them?"
		}
		_, err := msg.Reply(b, demoteOwnerMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}
	if userId == b.Id {
		demoteBotMsg, demoteBotErr := tr.GetStringWithError("strings.admin.demote.is_bot_itself")
		if demoteBotErr != nil {
			log.Errorf("[admin] missing translation for demote.is_bot_itself: %v", demoteBotErr)
			demoteBotMsg = "I am not going to demote myself."
		}
		_, err := msg.Reply(b, demoteBotMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}
	if !chat_status.IsUserAdmin(b, chat.Id, userId) {
		demoteAdminMsg, demoteAdminErr := tr.GetStringWithError("strings.admin.demote.is_admin")
		if demoteAdminErr != nil {
			log.Errorf("[admin] missing translation for demote.is_admin: %v", demoteAdminErr)
			demoteAdminMsg = "This person is not an admin, and how I am supposed to demote them?"
		}
		_, err := msg.Reply(b, demoteAdminMsg, helpers.Shtml())
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
		demoteErrorMsg, demoteErrorErr := tr.GetStringWithError("strings.admin.errors.err_cannot_demote")
		if demoteErrorErr != nil {
			log.Errorf("[admin] missing translation for errors.err_cannot_demote: %v", demoteErrorErr)
			demoteErrorMsg = "Failed to demote; I might not be the admin, or they may be promoted by another admin."
		}
		_, err = msg.Reply(b, demoteErrorMsg, nil)
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}

	userMember, err := b.GetChatMember(chat.Id, userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}
	mem := userMember.MergeChatMember().User
	demoteMsg, demoteErr := tr.GetStringWithError("strings.admin.demote.success_demote")
	if demoteErr != nil {
		log.Errorf("[admin] missing translation for demote.success_demote: %v", demoteErr)
		demoteMsg = "User %s has been demoted from admin."
	}
	_, err = msg.Reply(b,
		fmt.Sprintf(demoteMsg, helpers.MentionHtml(mem.Id, mem.FirstName)),
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
		promoteBotMsg, promoteBotErr := tr.GetStringWithError("strings.admin.promote.is_bot_itself")
		if promoteBotErr != nil {
			log.Errorf("[admin] missing translation for promote.is_bot_itself: %v", promoteBotErr)
			promoteBotMsg = "If only I could do this to myself ;_;"
		}
		_, err := msg.Reply(b, promoteBotMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}

	// checks if user being promoted is already admin or owner
	if chat_status.RequireUserOwner(b, ctx, nil, userId, true) {
		promoteOwnerMsg, promoteOwnerErr := tr.GetStringWithError("strings.admin.promote.is_owner")
		if promoteOwnerErr != nil {
			log.Errorf("[admin] missing translation for promote.is_owner: %v", promoteOwnerErr)
			promoteOwnerMsg = "This person created this chat, and how would am I supposed to promote them?"
		}
		_, err := msg.Reply(b, promoteOwnerMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}
	if chat_status.IsUserAdmin(b, chat.Id, userId) {
		promoteAdminMsg, promoteAdminErr := tr.GetStringWithError("strings.admin.promote.is_admin")
		if promoteAdminErr != nil {
			log.Errorf("[admin] missing translation for promote.is_admin: %v", promoteAdminErr)
			promoteAdminMsg = "This person is already an admin, and how would am I supposed to promote them?"
		}
		_, err := msg.Reply(b, promoteAdminMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}

	userMember, err := b.GetChatMember(chat.Id, userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	botMember, err := b.GetChatMember(chat.Id, b.Id, nil)
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
		promoteErrorMsg, promoteErrorErr := tr.GetStringWithError("strings.admin.errors.err_cannot_promote")
		if promoteErrorErr != nil {
			log.Errorf("[admin] missing translation for errors.err_cannot_promote: %v", promoteErrorErr)
			promoteErrorMsg = "Failed to promote; I might not be the admin, or they may be promoted by another admin."
		}
		_, _ = msg.Reply(b, promoteErrorMsg, helpers.Shtml())
		return err
	}

	if len(customTitle) > 16 {
		// trim title to 16 characters (telegram restriction)
		titleTruncatedMsg, titleTruncatedErr := tr.GetStringWithError("strings.admin.promote.admin_title_truncated")
		if titleTruncatedErr != nil {
			log.Errorf("[admin] missing translation for promote.admin_title_truncated: %v", titleTruncatedErr)
			titleTruncatedMsg = "Admin title truncated to 16 characters from %d"
		}
		extraText += fmt.Sprintf(titleTruncatedMsg, len(customTitle))
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
			setTitleErrorMsg, setTitleErrorErr := tr.GetStringWithError("strings.admin.errors.err_set_title")
			if setTitleErrorErr != nil {
				log.Errorf("[admin] missing translation for errors.err_set_title: %v", setTitleErrorErr)
				setTitleErrorMsg = "Failed to set custom admin title; The Title may not be correct or may contain emojis."
			}
			_, err = msg.Reply(b, setTitleErrorMsg, nil)
			if err != nil {
				log.Error(err)
			}
			return ext.EndGroups
		}
	}
	mem := userMember.MergeChatMember().User
	promoteMsg, promoteErr := tr.GetStringWithError("strings.admin.promote.success_promote")
	if promoteErr != nil {
		log.Errorf("[admin] missing translation for promote.success_promote: %v", promoteErr)
		promoteMsg = "User %s has been promoted to admin."
	}
	_, err = msg.Reply(b,
		fmt.Sprintf(promoteMsg, helpers.MentionHtml(mem.Id, mem.FirstName))+extraText,
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
	inviteLinkMsg, err := tr.GetStringWithError("strings.Admin.here_is_the_invite_link_of_this_chat_percent")
	if err != nil {
		log.Error(err)
		inviteLinkMsg = "Here is the invite link of this chat: %s"
	}
	if chat.Username != "" {
		_, _ = msg.Reply(b, fmt.Sprintf(inviteLinkMsg, chat.Username), nil)
	} else {
		nchat, err := b.GetChat(chat.Id, nil)
		if err != nil {
			_, _ = msg.Reply(b, err.Error(), nil)
			return ext.EndGroups
		}
		_, _ = msg.Reply(b, fmt.Sprintf(inviteLinkMsg, nchat.InviteLink), nil)
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
		anonUserMsg, anonUserErr := tr.GetStringWithError("strings.Warns.errors.anon_user")
		if anonUserErr != nil {
			log.Errorf("[admin] missing translation for errors.anon_user: %v", anonUserErr)
			anonUserMsg = "Anonymous users cannot be managed."
		}
		_, err := msg.Reply(b, anonUserMsg, nil)
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	} else if userId == 0 {
		noUserMsg, err := tr.GetStringWithError("strings.CommonStrings.errors.no_user_specified")
		if err != nil {
			log.Error(err)
			noUserMsg = "No user specified"
		}
		_, err = msg.Reply(b, noUserMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if chat_status.RequireUserOwner(b, ctx, nil, userId, true) {
		titleOwnerMsg, titleOwnerErr := tr.GetStringWithError("strings.admin.title.is_owner")
		if titleOwnerErr != nil {
			log.Errorf("[admin] missing translation for title.is_owner: %v", titleOwnerErr)
			titleOwnerMsg = "This person created this chat, and how would am I supposed to set a admin title to them?"
		}
		_, err := msg.Reply(b, titleOwnerMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}
	if !chat_status.IsUserAdmin(b, chat.Id, userId) {
		titleAdminMsg, titleAdminErr := tr.GetStringWithError("strings.admin.title.is_admin")
		if titleAdminErr != nil {
			log.Errorf("[admin] missing translation for title.is_admin: %v", titleAdminErr)
			titleAdminMsg = "This person is already an admin, how would I set a custom admin title for them?"
		}
		_, err := msg.Reply(b, titleAdminMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	if userId == b.Id {
		titleBotMsg, titleBotErr := tr.GetStringWithError("strings.admin.title.is_bot_itself")
		if titleBotErr != nil {
			log.Errorf("[admin] missing translation for title.is_bot_itself: %v", titleBotErr)
			titleBotMsg = "If only I could do this to myself ;_;"
		}
		_, err := msg.Reply(b, titleBotMsg, helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}

		return ext.EndGroups
	}

	// for managing custom title
	if customTitle == "" {
		titleEmptyMsg, titleEmptyErr := tr.GetStringWithError("strings.admin.errors.title_empty")
		if titleEmptyErr != nil {
			log.Errorf("[admin] missing translation for errors.title_empty: %v", titleEmptyErr)
			titleEmptyMsg = "You need to give me an admin title to set it."
		}
		_, err := msg.Reply(b, titleEmptyMsg, helpers.Shtml())
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
		setTitleErrorMsg2, setTitleErrorErr2 := tr.GetStringWithError("strings.admin.errors.err_set_title")
		if setTitleErrorErr2 != nil {
			log.Errorf("[admin] missing translation for errors.err_set_title: %v", setTitleErrorErr2)
			setTitleErrorMsg2 = "Failed to set custom admin title; The Title may not be correct or may contain emojis."
		}
		_, _ = msg.Reply(b, setTitleErrorMsg2, helpers.Shtml())
		return err
	}

	userMember, err := b.GetChatMember(chat.Id, userId, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	mem := userMember.MergeChatMember()

	titleSuccessMsg, titleSuccessErr := tr.GetStringWithError("strings.admin.title.success_set")
	if titleSuccessErr != nil {
		log.Errorf("[admin] missing translation for title.success_set: %v", titleSuccessErr)
		titleSuccessMsg = "Successfully set %s's admin title to <b>%s</b>"
	}
	_, err = msg.Reply(b,
		fmt.Sprintf(titleSuccessMsg, mem.User.FirstName, mem.CustomTitle),
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
			anonEnabledMsg, anonEnabledErr := tr.GetStringWithError("strings.admin.anon_admin.enabled")
			if anonEnabledErr != nil {
				log.Errorf("[admin] missing translation for anon_admin.enabled: %v", anonEnabledErr)
				anonEnabledMsg = "AnonAdmin mode is currently <b>enabled</b> for %s.\n\nThis allows all anonymous admin to perform admin actions without restriction."
			}
			text = fmt.Sprintf(anonEnabledMsg, chat.Title)
		} else {
			anonDisabledMsg, anonDisabledErr := tr.GetStringWithError("strings.admin.anon_admin.disabled")
			if anonDisabledErr != nil {
				log.Errorf("[admin] missing translation for anon_admin.disabled: %v", anonDisabledErr)
				anonDisabledMsg = "AnonAdmin mode is currently <b>disabled</b> for %s.\n\nThis requires anonymous admins to press a button to confirm their permissions."
			}
			text = fmt.Sprintf(anonDisabledMsg, chat.Title)
		}
	} else {
		// only need owner if you want to change value
		if !chat_status.RequireUserOwner(b, ctx, nil, user.Id, false) {
			return ext.EndGroups
		}
		switch args[1] {
		case "on", "true", "yes":
			if adminSettings.AnonAdmin {
				anonAlreadyEnabledMsg, anonAlreadyEnabledErr := tr.GetStringWithError("strings.admin.anon_admin.already_enabled")
				if anonAlreadyEnabledErr != nil {
					log.Errorf("[admin] missing translation for anon_admin.already_enabled: %v", anonAlreadyEnabledErr)
					anonAlreadyEnabledMsg = "AnonAdmin mode is already <b>enabled</b> for %s"
				}
				text = anonAlreadyEnabledMsg
			} else {
				go db.SetAnonAdminMode(chat.Id, true)
				anonEnabledNowMsg, anonEnabledNowErr := tr.GetStringWithError("strings.admin.anon_admin.enabled_now")
				if anonEnabledNowErr != nil {
					log.Errorf("[admin] missing translation for anon_admin.enabled_now: %v", anonEnabledNowErr)
					anonEnabledNowMsg = "AnonAdmin mode is now <b>enabled</b> for %s.\n\nFrom now onwards, I will ask the admins to verify permissions from anonymous admins."
				}
				text = fmt.Sprintf(anonEnabledNowMsg, chat.Title)
			}
		case "off", "no", "false":
			if !adminSettings.AnonAdmin {
				anonAlreadyDisabledMsg, anonAlreadyDisabledErr := tr.GetStringWithError("strings.admin.anon_admin.already_disabled")
				if anonAlreadyDisabledErr != nil {
					log.Errorf("[admin] missing translation for anon_admin.already_disabled: %v", anonAlreadyDisabledErr)
					anonAlreadyDisabledMsg = "AnonAdmin mode is already <b>disabled</b> for %s"
				}
				text = anonAlreadyDisabledMsg
			} else {
				go db.SetAnonAdminMode(chat.Id, false)
				anonDisabledNowMsg, anonDisabledNowErr := tr.GetStringWithError("strings.admin.anon_admin.disabled_now")
				if anonDisabledNowErr != nil {
					log.Errorf("[admin] missing translation for anon_admin.disabled_now: %v", anonDisabledNowErr)
					anonDisabledNowMsg = "AnonAdmin mode is now <b>disabled</b> for %s.\n\nFrom now onwards, I won't ask the admins to verify for permissions anymore from anonymous admins."
				}
				text = fmt.Sprintf(anonDisabledNowMsg, chat.Title)
			}
		default:
			anonInvalidArgMsg, anonInvalidArgErr := tr.GetStringWithError("strings.admin.anon_admin.invalid_arg")
			if anonInvalidArgErr != nil {
				log.Errorf("[admin] missing translation for anon_admin.invalid_arg: %v", anonInvalidArgErr)
				anonInvalidArgMsg = "Invalid argument, I only understand <code>on</code>, <code>off</code>, <code>yes</code>, <code>no</code>"
			}
			text = anonInvalidArgMsg
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

	// permission checks
	userMember, _ := b.GetChatMember(chat.Id, user.Id, nil)
	mem := userMember.MergeChatMember()
	if mem.Status == "member" {
		adminMsg, err := tr.GetStringWithError("strings.Admin.you_need_to_be_admin_to_do_this")
		if err != nil {
			log.Error(err)
			adminMsg = "You need to be admin to do this"
		}
		_, err = msg.Reply(b, adminMsg, nil)
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

	cacheMsg, err := tr.GetStringWithError("strings.CommonStrings.admin_cache.cache_reloaded")
	if err != nil {
		log.Error(err)
		cacheMsg = "Admin cache reloaded"
	}
	_, err = msg.Reply(b, cacheMsg, helpers.Shtml())
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
