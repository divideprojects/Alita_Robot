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
from traceback import format_exc

from pyrogram import filters
from pyrogram.errors import PeerIdInvalid, RPCError, Unauthorized, UserIsBlocked
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import LOGGER, PREFIX_HANDLER, SUPPORT_STAFF
from alita.bot_class import Alita
from alita.database.reporting_db import Reporting
from alita.utils.custom_filters import admin_filter
from alita.utils.parser import mention_html

#  initialise
db = Reporting()


@Alita.on_message(
    filters.command("reports", PREFIX_HANDLER) & (filters.private | admin_filter),
)
async def report_setting(_, m: Message):
    args = m.text.split()

    if m.chat.type == "private":
        if len(args) >= 2:
            option = args[1].lower()
            if option in ("yes", "on", "true"):
                db.set_settings(m.chat.id, True)
                LOGGER.info(f"{m.from_user.id} enabled reports for them")
                await m.reply_text(
                    "Turned on reporting! You'll be notified whenever anyone reports something in groups you are admin.",
                )

            elif option in ("no", "off", "false"):
                db.set_settings(m.chat.id, False)
                LOGGER.info(f"{m.from_user.id} disabled reports for them")
                await m.reply_text("Turned off reporting! You wont get any reports.")
        else:
            await m.reply_text(
                f"Your current report preference is: `{(db.get_settings(m.chat.id))}`",
            )
    else:
        if len(args) >= 2:
            option = args[1].lower()
            if option in ("yes", "on", "true"):
                db.set_settings(m.chat.id, True)
                LOGGER.info(f"{m.from_user.id} enabled reports in {m.chat.id}")
                await m.reply_text(
                    "Turned on reporting! Admins who have turned on reports will be notified when /report "
                    "or @admin is called.",
                    quote=True,
                )

            elif option in ("no", "off", "false"):
                db.set_settings(m.chat.id, False)
                LOGGER.info(f"{m.from_user.id} disabled reports in {m.chat.id}")
                await m.reply_text(
                    "Turned off reporting! No admins will be notified on /report or @admin.",
                    quote=True,
                )
        else:
            await m.reply_text(
                f"This group's current setting is: `{(db.get_settings(m.chat.id))}`",
            )


@Alita.on_message(
    (filters.command("report", PREFIX_HANDLER) | filters.regex(r"(?i)@admin(s)?"))
    & filters.group,
)
async def report_watcher(c: Alita, m: Message):

    if m.chat.type != "supergroup":
        return

    if not m.from_user:
        return

    me = await c.get_me()

    if (m.chat and m.reply_to_message) and (db.get_settings(m.chat.id)):
        reported_msg_id = m.reply_to_message.message_id
        reported_user = m.reply_to_message.from_user
        chat_name = m.chat.title or m.chat.username
        admin_list = await c.get_chat_members(m.chat.id, filter="administrators")

        if reported_user.id == me.id:
            await m.reply_text("Nice try.")
            return

        if reported_user.id in SUPPORT_STAFF:
            await m.reply_text("Uh? You reporting my support team?")
            return

        if m.chat.username:
            msg = (
                f"<b>⚠️ Report: </b>{m.chat.title}\n"
                f"<b> • Report by:</b> {(await mention_html(m.from_user.first_name, m.from_user.id))} (<code>{m.from_user.id}</code>)\n"
                f"<b> • Reported user:</b> {(await mention_html(reported_user.first_name, reported_user.id))} (<code>{reported_user.id}</code>)\n"
            )

            should_forward = False
        else:
            msg = f"{(await mention_html(m.from_user.first_name, m.from_user.id))} is calling for admins in '{chat_name}'!\n"
            should_forward = True

        link_chat_id = str(m.chat.id).replace("-100", "")
        link = f"https://t.me/c/{link_chat_id}/{reported_msg_id}"  # message link

        reply_markup = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        "➡ Message",
                        url=link,
                    ),
                ],
                [
                    InlineKeyboardButton(
                        "⚠ Kick",
                        callback_data=f"report_{m.chat.id}=kick={reported_user.id}={reported_msg_id}",
                    ),
                    InlineKeyboardButton(
                        "⛔️ Ban",
                        callback_data=f"report_{m.chat.id}=ban={reported_user.id}={reported_msg_id}",
                    ),
                ],
                [
                    InlineKeyboardButton(
                        "❎ Delete Message",
                        callback_data=f"report_{m.chat.id}=del={reported_user.id}={reported_msg_id}",
                    ),
                ],
            ],
        )

        LOGGER.info(
            f"{m.from_user.id} reported msgid-{m.reply_to_message.message_id} to admins in {m.chat.id}",
        )
        await m.reply_text(
            (
                f"{(await mention_html(m.from_user.first_name, m.from_user.id))} "
                "reported the message to the admins."
            ),
            quote=True,
        )

        for admin in admin_list:
            if (
                admin.user.is_bot or admin.user.is_deleted
            ):  # can't message bots or deleted accounts
                continue

            if db.get_settings(admin.user.id):
                try:
                    await c.send_message(
                        admin.user.id,
                        msg,
                        reply_markup=reply_markup,
                        disable_web_page_preview=True,
                    )

                    # TODO - fix message.forward
                    # if should_forward:
                    #     # forward the reported message
                    #     await m.reply_to_message.forward(admin.user.id)

                    #     if len(m.text.split()) > 1:
                    #         # If user is giving a reason, send his message too
                    #         await m.forward(admin.user.id)

                except (Unauthorized, UserIsBlocked, PeerIdInvalid):
                    pass
                except RPCError as ef:
                    LOGGER.error(ef)
                    LOGGER.error(format_exc())
    return


@Alita.on_callback_query(filters.regex("^report_"))
async def report_buttons(c: Alita, q: CallbackQuery):
    splitter = (str(q.data).replace("report_", "")).split("=")
    chat_id = int(splitter[0])
    action = str(splitter[1])
    user_id = int(splitter[2])
    message_id = int(splitter[3])
    if action == "kick":
        try:
            await c.kick_chat_member(chat_id, user_id, until_date=int(time() + 45))
            await q.answer("✅ Succesfully kicked")
            return
        except RPCError as err:
            await q.answer(
                f"🛑 Failed to Kick\n<b>Error:</b>\n</code>{err}</code>",
                show_alert=True,
            )
    elif action == "ban":
        try:
            await c.kick_chat_member(chat_id, user_id)
            await q.answer("✅ Succesfully Banned")
            return
        except RPCError as err:
            await q.answer(f"🛑 Failed to Ban\n<b>Error:</b>\n`{err}`", show_alert=True)
    elif action == "del":
        try:
            await c.delete_messages(chat_id, message_id)
            await q.answer("✅ Message Deleted")
            return
        except RPCError as err:
            await q.answer(
                f"🛑 Failed to delete message!\n<b>Error:</b>\n`{err}`",
                show_alert=True,
            )
    return


__PLUGIN__ = "plugins.reporting.main"
__help__ = "plugins.reporting.help"
__alt_name__ = ["reports", "report"]
