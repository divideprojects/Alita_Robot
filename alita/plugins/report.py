from html import escape
from pyrogram import filters
from pyrogram.types import (
    Message,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    CallbackQuery,
)
from pyrogram.errors import BadRequest, Unauthorized
from alita import LOGGER, PREFIX_HANDLER, SUPPORT_STAFF
from alita.bot_class import Alita
from alita.db import reporting_db as db
from alita.utils.parser import mention_html
from alita.utils.admin_check import admin_check


__PLUGIN__ = "Reporting"

__help__ = """
 × /report <reason>: reply to a message to report it to admins.
 - @admin: reply to a message to report it to admins.
**NOTE:** Neither of these will get triggered if used by admins.

**Admins Only:**
 × /reports <on/off>: change report setting, or view current status.
   - If done in PM, toggles your status.
   - If in group, toggles that groups's status.
"""


@Alita.on_message(filters.command("reports", PREFIX_HANDLER))
async def report_setting(c: Alita, m: Message):
    args = m.text.split()

    if m.chat.type == "private":
        if len(args) >= 2:
            if args[1] in ("yes", "on"):
                db.set_user_setting(m.chat.id, True)
                await m.reply_text(
                    "Turned on reporting! You'll be notified whenever anyone reports something."
                )

            elif args[1] in ("no", "off"):
                db.set_user_setting(m.chat.id, False)
                await m.reply_text("Turned off reporting! You wont get any reports.")
        else:
            await m.reply_text(
                f"Your current report preference is: `{db.user_should_report(m.chat.id)}`"
            )

    else:
        res = await admin_check(c, m)
        if not res:
            return

        if len(args) >= 2:
            if args[1] in ("yes", "on"):
                db.set_chat_setting(m.chat.id, True)
                await m.reply_text(
                    "Turned on reporting! Admins who have turned on reports will be notified when /report "
                    "or @admin is called.",
                    reply_to_message_id=m.message_id,
                )

            elif args[1] in ("no", "off"):
                db.set_chat_setting(m.chat.id, False)
                await m.reply_text(
                    "Turned off reporting! No admins will be notified on /report or @admin.",
                    reply_to_message_id=m.message_id,
                )
        else:
            await m.reply_text(
                f"This group's current setting is: `{db.chat_should_report(m.chat.id)}`"
            )


@Alita.on_message(filters.command("report", PREFIX_HANDLER))
async def report(c: Alita, m: Message):
    me = await c.get_me()

    if m.chat and m.reply_to_message and db.chat_should_report(m.chat.id):
        reported_user = m.reply_to_message.from_user
        chat_name = m.chat.title or m.chat.username
        admin_list = await c.get_chat_members(m.chat.id, filter="administrators")

        if m.from_user.id == reported_user.id:
            await m.reply_text(
                "Uh yeah, Sure sure...you don't need to report yourself!"
            )
            return

        if reported_user.id == me.id:
            await m.reply_text("Nice try.")
            return

        if reported_user.id in SUPPORT_STAFF:
            await m.reply_text("Uh? You reporting whitelisted users?")
            return

        if m.chat.username and m.chat.type == "supergroup":

            # reported = f"{(await mention_html(m.from_user.first_name, m.from_user.id)} reported {mention_html(reported_user.first_name, reported_user.id))} to the admins!"

            msg = (
                f"<b>⚠️ Report: </b>{escape(m.chat.title)}\n"
                f"<b> • Report by:</b> {(await mention_html(m.from_user.first_name, m.from_user.id))} (<code>{m.from_user.id}</code>)\n"
                f"<b> • Reported user:</b> {(await mention_html(reported_user.first_name, reported_user.id))} (<code>{reported_user.id}</code>)\n"
            )
            link = f'<b> • Reported message:</b> <a href="https://t.me/{m.chat.username}/{m.reply_to_message.message_id}">click here</a>'
            should_forward = False
            keyboard = [
                [
                    InlineKeyboardButton(
                        "➡ Message",
                        url=f"https://t.me/{m.chat.username}/{m.reply_to_message.message_id}",
                    )
                ],
                [
                    InlineKeyboardButton(
                        "⚠ Kick",
                        callback_data=f"report_{m.chat.id}=kick={reported_user.id}={reported_user.first_name}",
                    ),
                    InlineKeyboardButton(
                        "⛔️ Ban",
                        callback_data=f"report_{m.chat.id}=banned={reported_user.id}={reported_user.first_name}",
                    ),
                ],
                [
                    InlineKeyboardButton(
                        "❎ Delete Message",
                        callback_data=f"report_{m.chat.id}=delete={reported_user.id}={m.reply_to_message.message_id}",
                    )
                ],
            ]
            reply_markup = InlineKeyboardMarkup(keyboard)
        else:
            # reported = f"{(await mention_html(m.from_user.first_name, m.from_user.id)} reported {mention_html(reported_user.first_name, reported_user.id))} to the admins!"

            msg = f'{(await mention_html(m.from_user.first_name, m.from_user.id))} is calling for admins in f"{escape(chat_name)}"!'
            link = ""
            should_forward = True

        for admin in admin_list:
            if admin.user.is_bot:  # can't message bots
                continue

            if db.user_should_report(admin.user.id):
                try:
                    if not m.chat.type == "supergroup":
                        await c.send_message(admin.user.id, msg + link)

                        if should_forward:
                            await m.reply_to_message.forward(admin.user.id)

                            if (
                                len(m.text.split()) > 1
                            ):  # If user is giving a reason, send his message too
                                await m.reply_to_message.forward(admin.user.id)
                    if not m.chat.username:
                        await c.send_message(admin.user.id, msg + link)

                        if should_forward:
                            await m.reply_to_message.forward(admin.user.id)

                            if (
                                len(m.text.split()) > 1
                            ):  # If user is giving a reason, send his message too
                                await m.forward(admin.user.id)

                    if m.chat.username and m.chat.type == "supergroup":
                        await c.send_message(
                            admin.user.id, msg + link, reply_markup=reply_markup
                        )

                        if should_forward:
                            await m.reply_to_message.forward(admin.user.id)

                            if (
                                len(m.text.split()) > 1
                            ):  # If user is giving a reason, send his message too
                                await m.forward(admin.user.id)

                except Unauthorized:
                    pass
                except BadRequest:
                    LOGGER.exception("Exception while reporting user")

        await m.reply_to_message.reply_text(
            f"{(await mention_html(m.from_user.first_name, m.from_user.id))} reported the message to the admins."
        )
        return
    return


@Alita.on_callback_query(filters.regex("^report_"))
async def report_buttons(c: Alita, q: CallbackQuery):
    splitter = q.data.replace("report_", "").split("=")
    if splitter[1] == "kick":
        try:
            await c.kick_chat_member(splitter[0], splitter[2])
            await c.unban_chat_member(splitter[0], splitter[2])
            await q.answer("✅ Succesfully kicked")
            return
        except Exception as err:
            await q.answer(f"🛑 Failed to Kick\n<b>Error:</b>\n`{err}`", show_alert=True)
    elif splitter[1] == "banned":
        try:
            await c.kick_chat_member(splitter[0], splitter[2])
            await q.answer("✅ Succesfully Banned")
            return
        except Exception as err:
            await q.answer(f"🛑 Failed to Ban\n<b>Error:</b>\n`{err}`", show_alert=True)
    elif splitter[1] == "delete":
        try:
            await c.delete_messages(splitter[0], splitter[3])
            await q.answer("✅ Message Deleted")
            return
        except Exception as err:
            await q.answer(
                f"🛑 Failed to delete message!\n<b>Error:</b>\n`{err}`", show_alert=True
            )
    return
