from alita.__main__ import Alita
from pyrogram import filters, errors
from pyrogram.types import (
    Message,
    InlineKeyboardMarkup,
    InlineKeyboardButton,
    CallbackQuery,
)
from alita import PREFIX_HANDLER
from alita.utils.localization import GetLang
from alita.db import rules_db as db
from alita.utils.admin_check import admin_check

__PLUGIN__ = "Rules"

__help__ = """
Set rules for you chat so that members know what to do and \
what not to do in your group!

 × /rules: get the rules for current chat.

**Admin only:**
 × /setrules <rules>: Set the rules for this chat, also works as a reply to a message.
 × /clearrules: Clear the rules for this chat.
"""


@Alita.on_message(filters.command("rules", PREFIX_HANDLER) & filters.group)
async def get_rules(c: Alita, m: Message):
    _ = GetLang(m).strs

    chat_id = m.chat.id
    rules = db.get_rules(chat_id)

    if not rules:
        await m.reply_text(_("rules.no_rules"), reply_to_message_id=m.message_id)
        return

    try:
        await c.send_message(
            m.from_user.id,
            _("rules.get_rules").format(chat=m.chat.title, rules=rules),
        )
    except errors.UserIsBlocked:
        me = await c.get_me()
        pm_kb = InlineKeyboardMarkup(
            [[InlineKeyboardButton("PM", url=f"https://t.me/{me.username}?start")]]
        )
        await m.reply_text(
            _("rules.pm_me"), reply_to_message_id=m.message_id, reply_markup=pm_kb
        )
        return

    await m.reply_text(_("rules.sent_pm_rules"), reply_to_message_id=m.message_id)
    return


@Alita.on_message(filters.command("setrules", PREFIX_HANDLER) & filters.group)
async def set_rules(c: Alita, m: Message):

    res = await admin_check(c, m)
    if not res:
        return

    _ = GetLang(m).strs

    chat_id = m.chat.id
    if m.reply_to_message and m.reply_to_message.text:
        rules = m.reply_to_message.text
    elif (not m.reply_to_message) and len(m.text.split()) >= 2:
        rules = m.text.split(None, 1)[1]

    if len(rules) > 4000:
        rules = rules[0:3949]  # Split Rules if len > 4000 chars
        await m.reply_text("Rules truncated to 3950 characters!")

    db.set_rules(chat_id, rules)
    await m.reply_text(_("rules.set_rules"), reply_to_message_id=m.message_id)
    return


@Alita.on_message(filters.command("clearrules", PREFIX_HANDLER) & filters.group)
async def clear_rules(c: Alita, m: Message):

    res = await admin_check(c, m)
    if not res:
        return

    _ = GetLang(m).strs

    rules = db.get_rules(m.chat.id)
    if not rules:
        await m.reply_text(_("rules.no_rules"), reply_to_message_id=m.message_id)
        return

    await m.reply_text(
        "Are you sure you want to clear rules?",
        reply_markup=InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton("⚠️ Confirm", callback_data="clear.rules"),
                    InlineKeyboardButton("❌ Cancel", callback_data="close"),
                ]
            ]
        ),
    )
    return


@Alita.on_callback_query(filters.regex("^clear.rules$"))
async def clearrules_callback(c: Alita, q: CallbackQuery):
    _ = GetLang(q.message).strs
    db.clear_rules(q.message.chat.id)
    await q.message.reply_text(_("rules.clear"))
    await q.answer()
    return
