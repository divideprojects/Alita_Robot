from traceback import format_exc

from pyrogram.errors import (
    ChatAdminRequired,
    PeerIdInvalid,
    RightForbidden,
    RPCError,
    UserAdminInvalid,
)
from pyrogram.filters import regex
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import LOGGER, OWNER_ID, SUPPORT_GROUP, SUPPORT_STAFF
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils.caching import ADMIN_CACHE, admin_cache_reload
from alita.utils.custom_filters import command, restrict_filter
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html
from alita.utils.string import extract_time
from alita.vars import Config


@Alita.on_message(command("tban") & restrict_filter)
async def tban_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.ban.no_target"))
        await m.stop_propagation()

    try:
        user_id, user_first_name, _ = await extract_user(c, m)
    except Exception:
        return

    if not user_id:
        await m.reply_text("Cannot find user to ban")
        return
    if user_id == Config.BOT_ID:
        await m.reply_text("Huh, why would I ban myself?")
        await m.stop_propagation()

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        LOGGER.info(
            f"{m.from_user.id} trying to ban {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.stop_propagation()

    r_id = m.reply_to_message.message_id if m.reply_to_message else m.message_id

    if m.reply_to_message and len(m.text.split()) >= 2:
        reason = m.text.split(None, 2)[1]
    elif not m.reply_to_message and len(m.text.split()) >= 3:
        reason = m.text.split(None, 2)[2]
    else:
        await m.reply_text("Read /help !!")
        return

    if not reason:
        await m.reply_text("You haven't specified a time to ban this user for!")
        return

    split_reason = reason.split(None, 1)
    time_val = split_reason[0].lower()
    reason = split_reason[1] if len(split_reason) > 1 else ""

    bantime = await extract_time(m, time_val)

    if not bantime:
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "ban")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.ban.admin_cannot_ban"))
        await m.stop_propagation()

    try:
        LOGGER.info(f"{m.from_user.id} tbanned {user_id} in {m.chat.id}")
        await m.chat.ban_member(user_id, until_date=int(bantime))
        txt = (tlang(m, "admin.ban.banned_user")).format(
            admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
            banned=(await mention_html(user_first_name, user_id)),
            chat_title=m.chat.title,
        )
        txt += f"\n<b>Reason</b>: {reason}" if reason else ""
        keyboard = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        "Unban",
                        callback_data=f"unban_={user_id}",
                    ),
                ],
            ],
        )
        await m.reply_text(txt, reply_markup=keyboard, reply_to_message_id=r_id)
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except PeerIdInvalid:
        await m.reply_text(
            "I have not seen this user yet...!\nMind forwarding one of their message so I can recognize them?",
        )
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
    except RightForbidden:
        await m.reply_text(tlang(m, tlang(m, "admin.ban.bot_no_right")))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())
    return


@Alita.on_message(command("stban") & restrict_filter)
async def stban_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.ban.no_target"))
        await m.stop_propagation()

    try:
        user_id, _, _ = await extract_user(c, m)
    except Exception:
        return

    if not user_id:
        await m.reply_text("Cannot find user to ban")
        return
    if user_id == Config.BOT_ID:
        await m.reply_text("Huh, why would I ban myself?")
        await m.stop_propagation()

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        LOGGER.info(
            f"{m.from_user.id} trying to ban {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.stop_propagation()

    if m.reply_to_message and len(m.text.split()) >= 2:
        reason = m.text.split(None, 2)[1]
    elif not m.reply_to_message and len(m.text.split()) >= 3:
        reason = m.text.split(None, 2)[2]
    else:
        await m.reply_text("Read /help !!")
        return

    if not reason:
        await m.reply_text("You haven't specified a time to ban this user for!")
        return

    split_reason = reason.split(None, 1)
    time_val = split_reason[0].lower()
    reason = split_reason[1] if len(split_reason) > 1 else ""

    bantime = await extract_time(m, time_val)

    if not bantime:
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "ban")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.ban.admin_cannot_ban"))
        await m.stop_propagation()

    try:
        LOGGER.info(f"{m.from_user.id} stbanned {user_id} in {m.chat.id}")
        await m.chat.ban_member(user_id, until_date=int(bantime))
        await m.delete()
        if m.reply_to_message:
            await m.reply_to_message.delete()
            return
        return
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except PeerIdInvalid:
        await m.reply_text(
            "I have not seen this user yet...!\nMind forwarding one of their message so I can recognize them?",
        )
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
    except RightForbidden:
        await m.reply_text(tlang(m, tlang(m, "admin.ban.bot_no_right")))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())
    return


@Alita.on_message(command("dtban") & restrict_filter)
async def dtban_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.ban.no_target"))
        await m.stop_propagation()

    if not m.reply_to_message:
        await m.reply_text(
            "Reply to a message with this command to temp ban and delete the message.",
        )
        await m.stop_propagation()

    user_id = m.reply_to_message.from_user.id
    user_first_name = m.reply_to_message.from_user.first_name

    if not user_id:
        await m.reply_text("Cannot find user to ban")
        return
    if user_id == Config.BOT_ID:
        await m.reply_text("Huh, why would I ban myself?")
        await m.stop_propagation()

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        LOGGER.info(
            f"{m.from_user.id} trying to ban {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.stop_propagation()

    if m.reply_to_message and len(m.text.split()) >= 2:
        reason = m.text.split(None, 2)[1]
    elif not m.reply_to_message and len(m.text.split()) >= 3:
        reason = m.text.split(None, 2)[2]
    else:
        await m.reply_text("Read /help !!")
        return

    if not reason:
        await m.reply_text("You haven't specified a time to ban this user for!")
        return

    split_reason = reason.split(None, 1)
    time_val = split_reason[0].lower()
    reason = split_reason[1] if len(split_reason) > 1 else ""

    bantime = await extract_time(m, time_val)

    if not bantime:
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "ban")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.ban.admin_cannot_ban"))
        await m.stop_propagation()

    try:
        LOGGER.info(f"{m.from_user.id} dtbanned {user_id} in {m.chat.id}")
        await m.chat.ban_member(user_id, until_date=int(bantime))
        await m.reply_to_message.delete()
        txt = (tlang(m, "admin.ban.banned_user")).format(
            admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
            banned=(await mention_html(user_first_name, user_id)),
            chat_title=m.chat.title,
        )
        txt += f"\n<b>Reason</b>: {reason}" if reason else ""
        keyboard = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        "Unban",
                        callback_data=f"unban_={user_id}",
                    ),
                ],
            ],
        )
        await c.send_message(m.chat.id, txt, reply_markup=keyboard)
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except PeerIdInvalid:
        await m.reply_text(
            "I have not seen this user yet...!\nMind forwarding one of their message so I can recognize them?",
        )
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
    except RightForbidden:
        await m.reply_text(tlang(m, tlang(m, "admin.ban.bot_no_right")))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())
    return


@Alita.on_message(command("kick") & restrict_filter)
async def kick_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.kick.no_target"))
        return

    reason = None

    if m.reply_to_message:
        r_id = m.reply_to_message.message_id
        if len(m.text.split()) >= 2:
            reason = m.text.split(None, 1)[1]
    else:
        r_id = m.message_id
        if len(m.text.split()) >= 3:
            reason = m.text.split(None, 2)[2]
    try:
        user_id, user_first_name, _ = await extract_user(c, m)
    except Exception:
        return

    if not user_id:
        await m.reply_text("Cannot find user to kick")
        return

    if user_id == Config.BOT_ID:
        await m.reply_text("Huh, why would I kick myself?")
        await m.stop_propagation()

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        LOGGER.info(
            f"{m.from_user.id} trying to kick {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.stop_propagation()

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "kick")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.kick.admin_cannot_kick"))
        await m.stop_propagation()

    try:
        LOGGER.info(f"{m.from_user.id} kicked {user_id} in {m.chat.id}")
        await m.chat.ban_member(user_id)
        txt = (tlang(m, "admin.kick.kicked_user")).format(
            admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
            kicked=(await mention_html(user_first_name, user_id)),
            chat_title=m.chat.title,
        )
        txt += f"\n<b>Reason</b>: {reason}" if reason else ""
        await m.reply_text(txt, reply_to_message_id=r_id)
        await m.chat.unban_member(user_id)
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except PeerIdInvalid:
        await m.reply_text(
            "I have not seen this user yet...!\nMind forwarding one of their message so I can recognize them?",
        )
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.kick.bot_no_right"))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())

    return


@Alita.on_message(command("skick") & restrict_filter)
async def skick_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.kick.no_target"))
        return

    try:
        user_id, _, _ = await extract_user(c, m)
    except Exception:
        return

    if not user_id:
        await m.reply_text("Cannot find user to kick")
        return

    if user_id == Config.BOT_ID:
        await m.reply_text("Huh, why would I kick myself?")
        await m.stop_propagation()

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        LOGGER.info(
            f"{m.from_user.id} trying to skick {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.stop_propagation()

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "kick")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.kick.admin_cannot_kick"))
        await m.stop_propagation()

    try:
        LOGGER.info(f"{m.from_user.id} skicked {user_id} in {m.chat.id}")
        await m.chat.ban_member(user_id)
        await m.delete()
        if m.reply_to_message:
            await m.reply_to_message.delete()
        await m.chat.unban_member(user_id)
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except PeerIdInvalid:
        await m.reply_text(
            "I have not seen this user yet...!\nMind forwarding one of their message so I can recognize them?",
        )
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.kick.bot_no_right"))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())

    return


@Alita.on_message(command("dkick") & restrict_filter)
async def dkick_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.kick.no_target"))
        return
    if not m.reply_to_message:
        return await m.reply_text("Reply to a message to delete it and kick the user!")

    reason = None

    user_id = m.reply_to_message.from_user.id
    user_first_name = m.reply_to_message.from_user.first_name

    if not user_id:
        await m.reply_text("Cannot find user to kick")
        return

    if user_id == Config.BOT_ID:
        await m.reply_text("Huh, why would I kick myself?")
        await m.stop_propagation()

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        LOGGER.info(
            f"{m.from_user.id} trying to dkick {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.stop_propagation()

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "kick")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.kick.admin_cannot_kick"))
        await m.stop_propagation()

    try:
        LOGGER.info(f"{m.from_user.id} dkicked {user_id} in {m.chat.id}")
        await m.reply_to_message.delete()
        await m.chat.ban_member(user_id)
        txt = (tlang(m, "admin.kick.kicked_user")).format(
            admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
            kicked=(await mention_html(user_first_name, user_id)),
            chat_title=m.chat.title,
        )
        txt += f"\n<b>Reason</b>: {reason}" if reason else ""
        await c.send_message(m.chat.id, txt)
        await m.chat.unban_member(user_id)
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except PeerIdInvalid:
        await m.reply_text(
            "I have not seen this user yet...!\nMind forwarding one of their message so I can recognize them?",
        )
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.kick.bot_no_right"))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())

    return


@Alita.on_message(command("unban") & restrict_filter)
async def unban_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.unban.no_target"))
        await m.stop_propagation()

    if m.reply_to_message and not m.reply_to_message.from_user:
        user_id, user_first_name = (
            m.reply_to_message.sender_chat.id,
            m.reply_to_message.sender_chat.title,
        )
    else:
        try:
            user_id, user_first_name, _ = await extract_user(c, m)
        except Exception:
            return

    if m.reply_to_message and len(m.text.split()) >= 2:
        reason = m.text.split(None, 2)[1]
    elif not m.reply_to_message and len(m.text.split()) >= 3:
        reason = m.text.split(None, 2)[2]
    else:
        reason = None

    try:
        await m.chat.unban_member(user_id)
        txt = (tlang(m, "admin.unban.unbanned_user")).format(
            admin=m.from_user.mention,
            unbanned=(await mention_html(user_first_name, user_id)),
            chat_title=m.chat.title,
        )
        txt += f"\n<b>Reason</b>: {reason}" if reason else ""
        await m.reply_text(txt)
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, tlang(m, "admin.unban.bot_no_right")))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())

    return


@Alita.on_message(command("sban") & restrict_filter)
async def sban_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.ban.no_target"))
        await m.stop_propagation()

    if m.reply_to_message and not m.reply_to_message.from_user:
        user_id = m.reply_to_message.sender_chat.id
    else:
        try:
            user_id, _, _ = await extract_user(c, m)
        except Exception:
            return

    if not user_id:
        await m.reply_text("Cannot find user to ban")
        return
    if user_id == m.chat.id:
        await m.reply_text("That's an admin!")
        await m.stop_propagation()
    if user_id == Config.BOT_ID:
        await m.reply_text("Huh, why would I ban myself?")
        await m.stop_propagation()

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        LOGGER.info(
            f"{m.from_user.id} trying to sban {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.stop_propagation()

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "ban")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.ban.admin_cannot_ban"))
        await m.stop_propagation()

    try:
        LOGGER.info(f"{m.from_user.id} sbanned {user_id} in {m.chat.id}")
        await m.chat.ban_member(user_id)
        await m.delete()
        if m.reply_to_message:
            await m.reply_to_message.delete()
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except PeerIdInvalid:
        await m.reply_text(
            "I have not seen this user yet...!\nMind forwarding one of their message so I can recognize them?",
        )
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
    except RightForbidden:
        await m.reply_text(tlang(m, tlang(m, "admin.ban.bot_no_right")))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())
    return


@Alita.on_message(command("dban") & restrict_filter)
async def dban_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.ban.no_target"))
        await m.stop_propagation()

    if not m.reply_to_message:
        return await m.reply_text("Reply to a message to delete it and ban the user!")

    if m.reply_to_message and not m.reply_to_message.from_user:
        user_id, user_first_name = (
            m.reply_to_message.sender_chat.id,
            m.reply_to_message.sender_chat.title,
        )
    else:
        user_id, user_first_name = (
            m.reply_to_message.from_user.id,
            m.reply_to_message.from_user.first_name,
        )

    if not user_id:
        await m.reply_text("Cannot find user to ban")
        return
    if user_id == m.chat.id:
        await m.reply_text("That's an admin!")
        await m.stop_propagation()
    if user_id == Config.BOT_ID:
        await m.reply_text("Huh, why would I ban myself?")
        await m.stop_propagation()

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        LOGGER.info(
            f"{m.from_user.id} trying to dban {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.stop_propagation()

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "ban")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.ban.admin_cannot_ban"))
        await m.stop_propagation()

    reason = None
    if len(m.text.split()) >= 2:
        reason = m.text.split(None, 1)[1]

    try:
        LOGGER.info(f"{m.from_user.id} dbanned {user_id} in {m.chat.id}")
        await m.reply_to_message.delete()
        await m.chat.ban_member(user_id)
        txt = (tlang(m, "admin.ban.banned_user")).format(
            admin=m.from_user.mention,
            banned=m.reply_to_message.from_user.mention,
            chat_title=m.chat.title,
        )
        txt += f"\n<b>Reason</b>: {reason}" if reason else ""
        keyboard = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        "Unban",
                        callback_data=f"unban_={user_id}",
                    ),
                ],
            ],
        )
        await c.send_message(m.chat.id, txt, reply_markup=keyboard)
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except PeerIdInvalid:
        await m.reply_text(
            "I have not seen this user yet...!\nMind forwarding one of their message so I can recognize them?",
        )
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
    except RightForbidden:
        await m.reply_text(tlang(m, tlang(m, "admin.ban.bot_no_right")))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())
    return


@Alita.on_message(command("ban") & restrict_filter)
async def ban_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.ban.no_target"))
        await m.stop_propagation()

    if m.reply_to_message and not m.reply_to_message.from_user:
        user_id, user_first_name = (
            m.reply_to_message.sender_chat.id,
            m.reply_to_message.sender_chat.title,
        )
    else:
        try:
            user_id, user_first_name, _ = await extract_user(c, m)
        except Exception:
            return

    if not user_id:
        await m.reply_text("Cannot find user to ban")
        await m.stop_propagation()
    if user_id == m.chat.id:
        await m.reply_text("That's an admin!")
        await m.stop_propagation()
    if user_id == Config.BOT_ID:
        await m.reply_text("Huh, why would I ban myself?")
        await m.stop_propagation()

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        LOGGER.info(
            f"{m.from_user.id} trying to ban {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.stop_propagation()

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "ban")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.ban.admin_cannot_ban"))
        await m.stop_propagation()

    reason = None
    if m.reply_to_message:
        r_id = m.reply_to_message.message_id
        if len(m.text.split()) >= 2:
            reason = m.text.split(None, 1)[1]
    else:
        r_id = m.message_id
        if len(m.text.split()) >= 3:
            reason = m.text.split(None, 2)[2]

    try:
        LOGGER.info(f"{m.from_user.id} banned {user_id} in {m.chat.id}")
        await m.chat.ban_member(user_id)
        txt = (tlang(m, "admin.ban.banned_user")).format(
            admin=m.from_user.mention,
            banned=(await mention_html(user_first_name, user_id)),
            chat_title=m.chat.title,
        )
        txt += f"\n<b>Reason</b>: {reason}" if reason else ""
        keyboard = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        "Unban",
                        callback_data=f"unban_={user_id}",
                    ),
                ],
            ],
        )
        await m.reply_text(txt, reply_markup=keyboard, reply_to_message_id=r_id)
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except PeerIdInvalid:
        await m.reply_text(
            "I have not seen this user yet...!\nMind forwarding one of their message so I can recognize them?",
        )
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
    except RightForbidden:
        await m.reply_text(tlang(m, tlang(m, "admin.ban.bot_no_right")))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())
    return


@Alita.on_callback_query(regex("^unban_"))
async def unbanbutton(c: Alita, q: CallbackQuery):
    splitter = (str(q.data).replace("unban_", "")).split("=")
    user_id = int(splitter[1])
    user = await q.message.chat.get_member(q.from_user.id)

    if not user.can_restrict_members and q.from_user.id != OWNER_ID:
        await q.answer(
            "You don't have enough permission to do this!\nStay in your limits!",
            show_alert=True,
        )
        return
    whoo = await c.get_chat(user_id)
    doneto = whoo.first_name if whoo.first_name else whoo.title
    try:
        await q.message.chat.unban_member(user_id)
    except RPCError as e:
        await q.message.edit_text(f"Error: {e}")
        return
    await q.message.edit_text(f"{q.from_user.mention} unbanned {doneto}!")
    return


@Alita.on_message(command("kickme"))
async def kickme(_, m: Message):
    reason = None
    if len(m.text.split()) >= 2:
        reason = m.text.split(None, 1)[1]
    try:
        LOGGER.info(f"{m.from_user.id} kickme used by {m.from_user.id} in {m.chat.id}")
        await m.chat.ban_member(m.from_user.id)
        txt = "Why not let me help you!"
        txt += f"\n<b>Reason</b>: {reason}" if reason else ""
        await m.reply_text(txt)
        await m.chat.unban_member(m.from_user.id)
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
    return


__PLUGIN__ = "bans"

_DISABLE_CMDS_ = ["kickme"]

__alt_name__ = [
    "ban",
    "unban",
    "kickme",
    "kick",
    "tban",
]
