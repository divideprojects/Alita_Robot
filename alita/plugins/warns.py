# Copyright (C) 2020 - 2021 Divkix. All rights reserved. Source code available under the AGPL.
#
# This file is part of Alita_Robot.
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as
# published by the Free Software Foundation, either version 3 of the
# License, or (at your option) any later version.

# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.

# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.


from time import time

from pyrogram import filters
from pyrogram.errors import RPCError
from pyrogram.types import (
    CallbackQuery,
    ChatPermissions,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import LOGGER, PREFIX_HANDLER, SUPPORT_STAFF
from alita.bot_class import Alita
from alita.database.rules_db import Rules
from alita.database.users_db import Users
from alita.database.warns_db import Warns, WarnSettings
from alita.tr_engine import tlang
from alita.utils.caching import ADMIN_CACHE, admin_cache_reload
from alita.utils.custom_filters import admin_filter, restrict_filter
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html

warn_db = Warns()
rules_db = Rules()
warn_settings_db = WarnSettings()
users_db = Users()


@Alita.on_message(
    filters.command(["warn", "swarn", "dwarn"], PREFIX_HANDLER) & restrict_filter,
)
async def warn(c: Alita, m: Message):
    from alita import BOT_ID, BOT_USERNAME

    if m.reply_to_message and len(m.text.split()) >= 2:
        reason = m.text.split(None, 1)[1]
    elif not m.reply_to_message and len(m.text.split()) >= 3:
        reason = m.text.split(None, 2)[2]
    else:
        reason = None

    if not len(m.command) > 1 and not m.reply_to_message:
        await m.reply_text("I can't warn nothing! Tell me user whom I should warn")
        return

    user_id, user_first_name, _ = await extract_user(c, m)

    if user_id == BOT_ID:
        await m.reply_text("Huh, why would I warn myself?")
        return

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        LOGGER.info(
            f"{m.from_user.id} trying to warn {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = {i[0] for i in (await admin_cache_reload(m, "warn_user"))}

    if user_id in admins_group:
        await m.reply_text("This user is admin in this chat, I can't warn them!")
        return

    _, num = warn_db.warn_user(m.chat.id, user_id, reason)
    warn_settings = warn_settings_db.get_warnings_settings(m.chat.id)
    if num >= warn_settings["warn_limit"]:
        if warn_settings["warn_mode"] == "kick":
            await m.chat.kick_member(user_id, until_date=int(time() + 45))
            action = "kicked"
        elif warn_settings["warn_mode"] == "ban":
            await m.chat.kick_member(user_id)
            action = "banned"
        elif warn_settings["warn_mode"] == "mute":
            await m.chat.restrict_member(user_id, ChatPermissions())
            action = "muted"
        await m.reply_text(
            (
                f"Warnings {num}/{warn_settings['warn_limit']}!"
                f"\n<b>Reason for last warn</b>:\n{reason}"
                if reason
                else "\n"
                f"{(await mention_html(user_first_name, user_id))} has been <b>{action}!</b>"
            ),
        )
        await m.stop_propagation()

    rules = rules_db.get_rules(m.chat.id)
    if rules:
        kb = InlineKeyboardButton(
            "Rules 📋",
            url=f"https://t.me/{BOT_USERNAME}?start=rules_{m.chat.id}",
        )
    else:
        kb = InlineKeyboardButton(
            "Kick ⚠️",
            callback_data=f"warn.kick.{user_id}",
        )

    if m.text.split()[0] == "/swarn":
        await m.delete()
        await m.stop_propagation()
    if m.text.split()[0] == "/dwarn":
        if not m.reply_to_message:
            await m.reply_text("Reply to a message to delete it and ban the user!")
            await m.stop_propagation()
        await m.reply_to_message.delete()
    txt = f"{(await mention_html(user_first_name, user_id))} has {num} warnings!"
    txt += f"\n<b>Reason for last warn</b>:\n{reason}" if reason else ""
    await m.reply_text(
        txt,
        reply_markup=InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        "Remove Warn ❌",
                        callback_data=f"warn.remove.{user_id}",
                    ),
                ]
                + [kb],
            ],
        ),
    )
    await m.stop_propagation()


@Alita.on_message(filters.command("resetwarns", PREFIX_HANDLER) & restrict_filter)
async def reset_warn(c: Alita, m: Message):
    from alita import BOT_ID

    if not len(m.command) > 1 and not m.reply_to_message:
        await m.reply_text("I can't warn nothing! Tell me user whom I should warn")
        return

    user_id, user_first_name, _ = await extract_user(c, m)

    if user_id == BOT_ID:
        await m.reply_text("Huh, why would I warn myself?")
        return

    if user_id in SUPPORT_STAFF:
        await m.reply_text(
            "They are support users, cannot be restriced, how am I then supposed to unrestrict them?",
        )
        LOGGER.info(
            f"{m.from_user.id} trying to resetwarn {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = {i[0] for i in (await admin_cache_reload(m, "reset_warns"))}

    if user_id in admins_group:
        await m.reply_text("This user is admin in this chat, I can't warn them!")
        return

    warn_db.reset_warns(m.chat.id, user_id)
    await m.reply_text(
        f"Warnings have been reset for {(await mention_html(user_first_name,user_id))}",
    )
    return


@Alita.on_message(filters.command("warns", PREFIX_HANDLER) & filters.group)
async def list_warns(c: Alita, m: Message):
    from alita import BOT_ID

    user_id, user_first_name, _ = await extract_user(c, m)

    if user_id == BOT_ID:
        await m.reply_text("Huh, why would I warn myself?")
        return

    if user_id in SUPPORT_STAFF:
        await m.reply_text("This user has no warns!")
        LOGGER.info(
            f"{m.from_user.id} trying to check warns of {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = {i[0] for i in (await admin_cache_reload(m, "warns"))}

    if user_id in admins_group:
        await m.reply_text(
            "This user is admin in this chat, they don't have any warns!",
        )
        return

    warns, num_warns = warn_db.get_warns(m.chat.id, user_id)
    if not warns:
        await m.reply_text("This user has no warns!")
        return
    msg = f"{(await mention_html(user_first_name,user_id))} has <b>{num_warns}</b> warns!\n\n<b>Reasons:</b>\n"
    msg += "\n".join([("- No reason" if i is None else f" - {i}") for i in warns])
    await m.reply_text(msg)
    return


@Alita.on_message(
    filters.command(["rmwarn", "removewarn"], PREFIX_HANDLER) & restrict_filter,
)
async def remove_warn(c: Alita, m: Message):
    from alita import BOT_ID

    if not len(m.command) > 1 and not m.reply_to_message:
        await m.reply_text(
            "I can't remove warns of nothing! Tell me user whose warn should be removed!",
        )
        return

    user_id, user_first_name, _ = await extract_user(c, m)

    if user_id == BOT_ID:
        await m.reply_text("Huh, why would I warn myself?")
        return

    if user_id in SUPPORT_STAFF:
        await m.reply_text("This user has no warns!")
        LOGGER.info(
            f"{m.from_user.id} trying to remove warns of {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = {i[0] for i in (await admin_cache_reload(m, "rmwarn"))}

    if user_id in admins_group:
        await m.reply_text(
            "This user is admin in this chat, they don't have any warns!",
        )
        return

    warns, _ = warn_db.get_warns(m.chat.id, user_id)
    if not warns:
        await m.reply_text("This user has no warnings!")
        return

    _, num_warns = warn_db.remove_warn(m.chat.id, user_id)
    await m.reply_text(
        (
            f"{(await mention_html(user_first_name,user_id))} now has <b>{num_warns}</b> warnings!\n"
            "Their last warn was removed."
        ),
    )
    return


@Alita.on_callback_query(filters.regex("^warn."))
async def remove_last_warn_btn(c: Alita, q: CallbackQuery):

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[q.message.chat.id]}
    except KeyError:
        admins_group = {i[0] for i in (await admin_cache_reload(q, "warn_btn"))}

    if not q.from_user.id in admins_group:
        await q.answer("You are not allowed to use this!", show_alert=True)
        return

    args = q.data.split(".")
    action = args[1]
    user_id = int(args[2])
    chat_id = int(q.message.chat.id)
    user = users_db.get_user_info(int(user_id))
    user_first_name = user["name"]

    if action == "remove":
        _, num_warns = warn_db.remove_warn(chat_id, user_id)
        await q.message.edit_text(
            (
                f"Admin {(await mention_html(q.from_user.first_name, q.from_user.id))} "
                "removed last warn for "
                f"{(await mention_html(user_first_name, user_id))}\n"
                f"<b>Current Warnings:</b> {num_warns}"
            ),
        )
    if action == "kick":
        try:
            await c.kick_chat_member(chat_id, user_id, until_date=int(time() + 45))
            await q.message.edit_text(
                (
                    f"Admin {(await mention_html(q.from_user.first_name, q.from_user.id))} "
                    "kicked user "
                    f"{(await mention_html(user_first_name, user_id))} for last warning!"
                ),
            )
        except RPCError as err:
            await q.message.edit_text(
                f"🛑 Failed to Kick\n<b>Error:</b>\n</code>{err}</code>",
            )

    await q.answer()
    return


@Alita.on_message(filters.command("warnings", PREFIX_HANDLER) & admin_filter)
async def get_settings(_, m: Message):
    settings = warn_settings_db.get_warnings_settings(m.chat.id)
    await m.reply_text(
        (
            "This group has these following settings:\n"
            f"<b>Warn Limit:</b> <code>{settings['warn_limit']}</code>\n"
            f"<b>Warn Mode:</b> <code>{settings['warn_mode']}</code>"
        ),
    )
    return


@Alita.on_message(filters.command("warnmode", PREFIX_HANDLER) & admin_filter)
async def warnmode(_, m: Message):
    if len(m.text.split()) > 1:
        wm = (m.text.split(None, 1)[1]).lower()
        if wm not in ("kick", "ban", "mute"):
            await m.reply_text(
                (
                    "Please choose a valid warn mode!"
                    "Valid options are: <code>ban</code>,<code>kick</code>,<code>mute</code>"
                ),
            )
            return
        warnmode = warn_settings_db.set_warnmode(m.chat.id, wm)
        await m.reply_text(f"Warn Mode has been set to: {warnmode}")
        return
    warnmode = warn_settings_db.get_warnmode(m.chat.id)
    await m.reply_text(f"This chats current Warn Mode is: {warnmode}")
    return


@Alita.on_message(filters.command("warnlimit", PREFIX_HANDLER) & admin_filter)
async def warnlimit(_, m: Message):
    if len(m.text.split()) > 1:
        wl = int(m.text.split(None, 1)[1]).lower()
        if not isinstance(wl, int):
            await m.reply_text("Warn Limit can only be a number!")
            return
        warnlimit = warn_settings_db.set_warnlimit(m.chat.id, wl)
        await m.reply_text(f"Warn Limit has been set to: {warnlimit}")
        return
    warnlimit = warn_settings_db.get_warnlimit(m.chat.id)
    await m.reply_text(f"This chats current Warn Limit is: {warnlimit}")
    return


__PLUGIN__ = "plugins.warnings.main"
__help__ = "plugins.warnings.help"
__alt_name__ = ["warn", "warning"]
