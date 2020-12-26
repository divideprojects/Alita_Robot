from alita.utils.localization import GetLang
from alita.__main__ import Alita
from pyrogram import filters, errors
from pyrogram.types import Message
from alita import PREFIX_HANDLER, LOGGER, SUPPORT_GROUP
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

 × /adminlist: List of admins in the chat
*Admin only:*
 × /pin: Silently pins the message replied to - add `loud`, `notify` or `alert` to give notificaton to users.
 × /unpin: Unpins the currently pinned message.
 × /invitelink: Gets private chat's invitelink.
 × /promote: Promotes the user replied to or tagged.
 × /demote: Demotes the user replied to or tagged.
 × /ban: Bans the user replied to or tagged.
 × /unban: Unbans the user replied to or tagged.

An example of promoting someone to admins:
`/promote @username`; this promotes a user to admins.
"""


@Alita.on_message(filters.command("adminlist", PREFIX_HANDLER) & filters.group)
async def adminlist(c: Alita, m: Message):
    _ = GetLang(m).strs
    try:
        me_id = int(get_key("BOT_ID"))  # Get Bot ID from Redis!
        adminlist = get_key("ADMINDICT")[str(m.chat.id)]  # Load ADMINDICT from string
        adminstr = _("admin.adminlist").format(chat_title=m.chat.title)
        for i in adminlist:
            usr = await c.get_users(i)
            if i == me_id:
                adminstr += f"- {mention_html(usr.first_name, i)} (Me)\n"
            else:
                usr = await c.get_users(i)
                adminstr += f"- {mention_html(usr.first_name, i)} (`{i}`)\n"
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

    ADMINDICT = get_key("ADMINDICT")  # Load ADMINDICT from string

    try:
        adminlist = []
        async for i in m.chat.iter_members(filter="administrators"):
            adminlist.append(i.user.id)
        ADMINDICT[str(m.chat.id)] = adminlist
        set_key("ADMINDICT", ADMINDICT)
        await m.reply_text(_("admin.reloadedadmins"))
        LOGGER.info(f"Reloaded admins for {m.chat.title}({m.chat.id})")
    except Exception as ef:
        await m.reply_text(_("admin.useadmincache"))
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
        user_id, user_first_name = await extract_user(c, m)
        try:
            await c.kick_chat_member(m.chat.id, user_id)
            await m.reply_text(f"Banned {mention_html(user_first_name, user_id)}")
        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except Exception as ef:
            await m.reply_text(f"Error: {ef}\n\nReport it in Support Group!")
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
        user_id, user_first_name = await extract_user(c, m)
        try:
            await c.unban_chat_member(m.chat.id, user_id)
            await m.reply_text(f"Unbanned {mention_html(user_first_name, user_id)}")
        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except Exception as ef:
            await m.reply_text(f"Error: {ef}\n\nReport it in Support Group!")
            LOGGER.error(ef)

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

        user_id, user_first_name = await extract_user(c, m)
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
                    promoter=mention_html(m.from_user.first_name, m.from_user.id),
                    promoted=mention_html(user_first_name, user_id),
                    chat_title=m.chat.title,
                )
            )

            # ----- Add admin to redis cache! -----
            ADMINDICT = get_key("ADMINDICT")  # Load ADMINDICT from string
            adminlist = []
            async for i in m.chat.iter_members(filter="administrators"):
                adminlist.append(i.user.id)
            ADMINDICT[str(m.chat.id)] = adminlist
            set_key("ADMINDICT", ADMINDICT)

        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except Exception as ef:
            await m.reply_text(_("admin.useadmincache"))
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

        user_id, user_first_name = await extract_user(c, m)
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
                    demoter=mention_html(m.from_user.first_name, m.from_user.id),
                    demoted=mention_html(user_first_name, user_id),
                    chat_title=m.chat.title,
                )
            )

            # ----- Add admin to redis cache! -----
            ADMINDICT = get_key("ADMINDICT")  # Load ADMINDICT from string
            adminlist = []
            async for i in m.chat.iter_members(filter="administrators"):
                adminlist.append(i.user.id)
            ADMINDICT[str(m.chat.id)] = adminlist
            set_key("ADMINDICT", ADMINDICT)

        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except Exception as ef:
            await m.reply_text(_("admin.useadmincache"))
            LOGGER.error(ef)

        return

    await m.reply_text(_("admin.nodemoteperm"))
    return


@Alita.on_message(filters.command("invitelink", PREFIX_HANDLER) & filters.group)
async def demote_usr(c: Alita, m: Message):

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
        except Exception as ef:
            await m.reply_text(_("admin.useadmincache"))
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

    pin_loud = m.text.split(" ", 1)
    if m.reply_to_message:
        try:
            disable_notification = True

            if len(pin_loud) >= 2 and pin_loud[1] in ["alert", "notify", "loud"]:
                disable_notification = False

            pinned_event = await c.pin_chat_message(
                m.chat.id,
                m.reply_to_m.message_id,
                disable_notification=disable_notification,
            )
            await m.reply_text(_("admin.pinnedmsg"))

        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except Exception as ef:
            await m.reply_text(_("admin.useadmincache"))
            LOGGER.error(ef)
    else:
        await m.reply_text(_("admin.nopinmsg"))
    return


@Alita.on_message(filters.command("unpin", PREFIX_HANDLER) & filters.me)
async def unpin_message(c: Alita, m: Message):

    _ = GetLang(m).strs

    res = await admin_check(c, m)
    if not res:
        return

    try:
        await m.chat.unpin_chat_message(m.chat.id)
    except errors.ChatAdminRequired:
        await m.reply_text(_("admin.notadmin"))
    except Exception as ef:
        await m.reply_text(
            _("admin.somerror").format(SUPPORT_GROUP=SUPPORT_GROUP, ef=ef)
        )
        LOGGER.error(ef)
    return
