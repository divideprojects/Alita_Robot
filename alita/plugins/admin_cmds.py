from time import time
from pyrogram import filters, errors
from pyrogram.types import Message, ChatPermissions
from alita import PREFIX_HANDLER, LOGGER, SUPPORT_GROUP
from alita.bot_class import Alita
from alita.utils.localization import GetLang
from alita.utils.admin_check import admin_check
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html
from alita.utils.redishelper import get_key, set_key

__PLUGIN__ = "Admin"
__help__ = """
Lazy to promote or demote someone for admins? Want to see basic information about chat? \
All stuff about chatroom such as admin lists, pinning or grabbing an invite link can be \
done easily using the bot.

 × /adminlist: List of admins in the chat.
*Admin only:*
 × /pin: Silently pins the message replied to - add `loud`, `notify` or `alert` to give notificaton to users.
 × /unpin: Unpins the currently pinned message. - add `all` to unpin all the messages in current chat.
 × /invitelink: Gets private chat's invitelink.
 × /mute: Mute the user replied to or mentioned.
 × /unmute: Unmutes the user mentioned or replied to.
 × /promote: Promotes the user replied to or tagged.
 × /demote: Demotes the user replied to or tagged.
 × /ban: Bans the user replied to or tagged.
 × /unban: Unbans the user replied to or tagged.

An example of promoting someone to admins:
`/promote @username`; this promotes a user to admins.
"""

# Mute permissions
mute_permission = ChatPermissions(
    can_send_messages=False,
    can_send_media_messages=False,
    can_send_stickers=False,
    can_send_animations=False,
    can_send_games=False,
    can_use_inline_bots=False,
    can_add_web_page_previews=False,
    can_send_polls=False,
    can_change_info=False,
    can_invite_users=True,
    can_pin_messages=False,
)

# Unmute permissions
unmute_permissions = ChatPermissions(
    can_send_messages=True,
    can_send_media_messages=True,
    can_send_stickers=True,
    can_send_animations=True,
    can_send_games=True,
    can_use_inline_bots=True,
    can_add_web_page_previews=True,
    can_send_polls=True,
    can_change_info=False,
    can_invite_users=True,
    can_pin_messages=False,
)


@Alita.on_message(filters.command("adminlist", PREFIX_HANDLER) & filters.group)
async def adminlist_show(c: Alita, m: Message):
    _ = GetLang(m).strs
    try:
        me_id = int(await get_key("BOT_ID"))  # Get Bot ID from Redis!
        adminlist = (await get_key("ADMINDICT"))[
            str(m.chat.id)
        ]  # Load ADMINDICT from string
        adminstr = _("admin.adminlist").format(chat_title=m.chat.title)
        for i in adminlist:
            try:
                usr = await c.get_users(i)
            except errors.PeerIdInvalid:
                pass
            if i == me_id:
                adminstr += f"- {(await mention_html(usr.first_name, i))} (Me)\n"
            else:
                usr = await c.get_users(i)
                adminstr += f"- {(await mention_html(usr.first_name, i))} (`{i}`)\n"
        await m.reply_text(adminstr)
    except Exception as ef:

        if str(ef) == str(m.chat.id):
            await m.reply_text(_("admin.useadmincache"))
        else:
            await m.reply_text(
                _("admin.somerror").format(SUPPORT_GROUP=SUPPORT_GROUP, ef=ef)
            )
            LOGGER.error(ef)

    return


@Alita.on_message(filters.command("admincache", PREFIX_HANDLER) & filters.group)
async def reload_admins(c: Alita, m: Message):

    _ = GetLang(m).strs

    res = await admin_check(c, m)
    if not res:
        return

    ADMINDICT = await get_key("ADMINDICT")  # Load ADMINDICT from string

    try:
        adminlist = []
        async for i in m.chat.iter_members(filter="administrators"):
            adminlist.append(i.user.id)
        ADMINDICT[str(m.chat.id)] = adminlist
        await set_key("ADMINDICT", ADMINDICT)
        await m.reply_text(_("admin.reloadedadmins"))
        LOGGER.info(f"Reloaded admins for {m.chat.title}({m.chat.id})")
    except Exception as ef:
        await m.reply_text(_("admin.useadmincache"))
        LOGGER.error(ef)

    return


@Alita.on_message(filters.command("kick", PREFIX_HANDLER) & filters.group)
async def kick_usr(c: Alita, m: Message):

    _ = GetLang(m).strs

    res = await admin_check(c, m)
    if not res:
        return

    from_user = await m.chat.get_member(m.from_user.id)

    if from_user.can_restrict_members or from_user.status == "creator":
        user_id, user_first_name = await extract_user(m)
        try:
            await c.kick_chat_member(m.chat.id, user_id, int(time() + 45))
            await m.reply_text(
                f"Banned {(await mention_html(user_first_name, user_id))}"
            )
        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except Exception as ef:
            await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
            LOGGER.error(ef)

    return


@Alita.on_message(filters.command("ban", PREFIX_HANDLER) & filters.group)
async def ban_usr(c: Alita, m: Message):

    _ = GetLang(m).strs

    res = await admin_check(c, m)
    if not res:
        return

    from_user = await m.chat.get_member(m.from_user.id)

    if from_user.can_restrict_members or from_user.status == "creator":
        user_id, user_first_name = await extract_user(m)
        try:
            await c.kick_chat_member(m.chat.id, user_id)
            await m.reply_text(
                f"Banned {(await mention_html(user_first_name, user_id))}"
            )
        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except Exception as ef:
            await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
            LOGGER.error(ef)

    return


@Alita.on_message(filters.command("unban", PREFIX_HANDLER) & filters.group)
async def unban_usr(c: Alita, m: Message):

    _ = GetLang(m).strs

    res = await admin_check(c, m)
    if not res:
        return

    from_user = await m.chat.get_member(m.from_user.id)

    if from_user.can_restrict_members or from_user.status == "creator":
        user_id, user_first_name = await extract_user(m)
        try:
            await c.unban_chat_member(m.chat.id, user_id)
            await m.reply_text(
                f"Unbanned {(await mention_html(user_first_name, user_id))}"
            )
        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except Exception as ef:
            await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
            LOGGER.error(ef)

    return


@Alita.on_message(filters.command("mute", PREFIX_HANDLER) & filters.group)
async def mute_usr(c: Alita, m: Message):

    _ = GetLang(m).strs

    res = await admin_check(c, m)
    if not res:
        return

    from_user = await m.chat.get_member(m.from_user.id)

    if from_user.can_restrict_members or from_user.status == "creator":
        user_id, user_first_name = await extract_user(m)
        try:
            await m.chat.restrict_member(user_id, mute_permission)
            await m.reply_text(
                f"<b>Muted</b> {(await mention_html(user_first_name,user_id))}"
            )
        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except Exception as ef:
            await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
            LOGGER.error(ef)

        return

    await m.reply_text("You don't have permissions to restrict users.")
    return


@Alita.on_message(filters.command("unmute", PREFIX_HANDLER) & filters.group)
async def unmute_usr(c: Alita, m: Message):

    _ = GetLang(m).strs

    res = await admin_check(c, m)
    if not res:
        return

    from_user = await m.chat.get_member(m.from_user.id)

    if from_user.can_restrict_members or from_user.status == "creator":
        user_id, user_first_name = await extract_user(m)
        try:
            await m.chat.restrict_member(user_id, unmute_permissions)
            await m.reply_text(
                f"<b>Unmuted</b> {(await mention_html(user_first_name,user_id))}"
            )
        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except Exception as ef:
            await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
            LOGGER.error(ef)
        return

    await m.reply_text("You don't have permissions to restrict users.")
    return


@Alita.on_message(filters.command("promote", PREFIX_HANDLER) & filters.group)
async def promote_usr(c: Alita, m: Message):

    _ = GetLang(m).strs

    res = await admin_check(c, m)
    if not res:
        return

    from_user = await m.chat.get_member(m.from_user.id)

    # If user does not have permission to promote other users, return
    if from_user.can_promote_members or from_user.status == "creator":

        user_id, user_first_name = await extract_user(m)
        try:
            await m.chat.promote_member(
                user_id=user_id,
                can_change_info=False,
                can_delete_messages=True,
                can_restrict_members=True,
                can_invite_users=True,
                can_pin_messages=True,
            )
            await m.reply_text(
                _("admin.promoted").format(
                    promoter=(
                        (await mention_html(m.from_user.first_name, m.from_user.id))
                    ),
                    promoted=(await mention_html(user_first_name, user_id)),
                    chat_title=m.chat.title,
                )
            )

            # ----- Add admin to redis cache! -----
            ADMINDICT = await get_key("ADMINDICT")  # Load ADMINDICT from string
            adminlist = []
            async for i in m.chat.iter_members(filter="administrators"):
                adminlist.append(i.user.id)
            ADMINDICT[str(m.chat.id)] = adminlist
            await set_key("ADMINDICT", ADMINDICT)

        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except errors.RightForbidden:
            await m.reply_text("I don't have enough rights topromote this user.")
        except Exception as ef:
            await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
            LOGGER.error(ef)

        return

    await m.reply_text(_("admin.nopromoteperm"))
    return


@Alita.on_message(filters.command("demote", PREFIX_HANDLER) & filters.group)
async def demote_usr(c: Alita, m: Message):

    _ = GetLang(m).strs

    res = await admin_check(c, m)
    if not res:
        return

    from_user = await m.chat.get_member(m.from_user.id)

    # If user does not have permission to demote other users, return
    if from_user.can_promote_members or from_user.status == "creator":

        user_id, user_first_name = await extract_user(m)
        try:
            await m.chat.promote_member(
                user_id=user_id,
                can_change_info=False,
                can_delete_messages=False,
                can_restrict_members=False,
                can_invite_users=False,
                can_pin_messages=False,
            )
            await m.reply_text(
                _("admin.demoted").format(
                    demoter=(
                        await mention_html(m.from_user.first_name, m.from_user.id)
                    ),
                    demoted=(await mention_html(user_first_name, user_id)),
                    chat_title=m.chat.title,
                )
            )

            # ----- Add admin to redis cache! -----
            ADMINDICT = await get_key("ADMINDICT")  # Load ADMINDICT from string
            adminlist = []
            async for i in m.chat.iter_members(filter="administrators"):
                adminlist.append(i.user.id)
            ADMINDICT[str(m.chat.id)] = adminlist
            await set_key("ADMINDICT", ADMINDICT)

        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except errors.RightForbidden:
            await m.reply_text("I don't have enough rights to demote this user.")
        except Exception as ef:
            await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
            LOGGER.error(ef)

        return

    await m.reply_text(_("admin.nodemoteperm"))
    return


@Alita.on_message(filters.command("invitelink", PREFIX_HANDLER) & filters.group)
async def get_invitelink(c: Alita, m: Message):

    _ = GetLang(m).strs

    res = await admin_check(c, m)
    if not res:
        return

    from_user = await m.chat.get_member(m.from_user.id)

    # If user does not have permission to invite other users, return
    if from_user.can_invite_users or from_user.status == "creator":

        try:
            link = await c.export_chat_invite_link(m.chat.id)
            await m.reply_text(_("admin.invitelink").format(link=link))
        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except errors.ChatAdminInviteRequired:
            await m.reply_text(_("admin.noinviteperm"))
        except errors.RightForbidden:
            await m.reply_text("I don't have enough rights to view invitelink.")
        except Exception as ef:
            await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
            LOGGER.error(ef)

        return

    await m.reply_text(_("admin.nouserinviteperm"))
    return


@Alita.on_message(filters.command("pin", PREFIX_HANDLER) & filters.group)
async def pin_message(c: Alita, m: Message):

    _ = GetLang(m).strs

    res = await admin_check(c, m)
    if not res:
        return

    pin_loud = m.text.split(None, 1)
    if m.reply_to_message:
        try:
            disable_notification = True

            if len(pin_loud) >= 2 and pin_loud[1] in ["alert", "notify", "loud"]:
                disable_notification = False

            await c.pin_chat_message(
                m.chat.id,
                m.reply_to_message.message_id,
                disable_notification=disable_notification,
            )
            await m.reply_text(_("admin.pinnedmsg"))

        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except errors.RightForbidden:
            await m.reply_text("I don't have enough rights to pin messages.")
        except Exception as ef:
            await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
            LOGGER.error(ef)
    else:
        await m.reply_text(_("admin.nopinmsg"))
    return


@Alita.on_message(filters.command("unpin", PREFIX_HANDLER) & filters.group)
async def unpin_message(c: Alita, m: Message):

    _ = GetLang(m).strs

    res = await admin_check(c, m)
    if not res:
        return

    try:
        if len(m.command) > 1 and m.command[1] == "all":
            await c.unpin_all_chat_messages(m.chat.id)
            await m.reply_text("Unpinned all messages!")
        else:
            await m.chat.unpin_chat_message(m.chat.id)
            await m.reply_text("Unpinned last message!")
    except errors.ChatAdminRequired:
        await m.reply_text(_("admin.notadmin"))
    except errors.RightForbidden:
        await m.reply_text("I don't have enough rights to unpin messages")
    except Exception as ef:
        await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
        LOGGER.error(ef)

    return
