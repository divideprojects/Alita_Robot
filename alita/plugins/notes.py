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


from secrets import choice
from traceback import format_exc

from pyrogram import filters
from pyrogram.errors import RPCError
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import LOGGER, PREFIX_HANDLER
from alita.bot_class import Alita
from alita.database.notes_db import Notes, NotesSettings
from alita.utils.cmd_senders import send_cmd
from alita.utils.custom_filters import admin_filter, owner_filter
from alita.utils.msg_types import Types, get_note_type
from alita.utils.parser import mention_html
from alita.utils.string import (
    build_keyboard,
    escape_mentions_using_curly_brackets,
    parse_button,
)

# Initialise
db = Notes()
db_settings = NotesSettings()


@Alita.on_message(filters.command("save", PREFIX_HANDLER) & admin_filter)
async def save_note(_, m: Message):

    existing_notes = [i[0] for i in db.get_all_notes(m.chat.id)]

    note_name, text, data_type, content = await get_note_type(m)
    note_name = note_name.lower()

    if note_name in existing_notes:
        await m.reply_text(f"This note ({note_name}) already exists!")
        return

    if not note_name:
        await m.reply_text(
            f"<code>{m.text}</code>\n\nError: You must give a name for this note!",
        )
        return

    if (
        (not m.reply_to_message)
        and (data_type == Types.TEXT)
        and (not len(m.command) > 2)
    ):
        await m.reply_text(
            f"<code>{m.text}</code>\n\nError: There is no text in here!",
        )
        return

    db.save_note(m.chat.id, note_name, text, data_type, content)
    LOGGER.info(f"{m.from_user.id} saved note ({note_name}) in {m.chat.id}")
    await m.reply_text(
        f"Saved note <code>{note_name}</code>!\nGet it with <code>/get {note_name}</code> or <code>#{note_name}</code>",
    )
    return


async def get_note_func(c: Alita, m: Message, note_name, priv_notes_status):
    """Get the note in normal mode, with parsing enabled."""

    reply_text = m.reply_to_message.reply_text if m.reply_to_message else m.reply_text
    reply_msg_id = m.reply_to_message.message_id if m.reply_to_message else m.message_id

    if priv_notes_status:
        from alita import BOT_USERNAME

        note_hash = [i[1] for i in db.get_all_notes(m.chat.id) if i[0] == note_name][0]
        await reply_text(
            f"Click on the button to get the note <code>{note_name}</code>",
            reply_markup=InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(
                            "Click Me!",
                            url=f"https://t.me/{BOT_USERNAME}?start=note_{m.chat.id}_{note_hash}",
                        ),
                    ],
                ],
            ),
        )
        return

    getnotes = db.get_note(m.chat.id, note_name)

    msgtype = getnotes["msgtype"]
    if not msgtype:
        await reply_text("<b>Error:</b> Cannot find a type for this note!!")
        return

    try:
        # support for random notes texts
        splitter = "%%%"
        note_reply = getnotes["note_value"].split(splitter)
        note_reply = choice(note_reply)
    except KeyError:
        note_reply = ""

    parse_words = [
        "first",
        "last",
        "fullname",
        "username",
        "id",
        "chatname",
        "mention",
    ]
    text = await escape_mentions_using_curly_brackets(m, note_reply, parse_words)

    if msgtype == Types.TEXT:

        teks, button = await parse_button(text)
        button = await build_keyboard(button)
        button = InlineKeyboardMarkup(button) if button else None
        if button:
            try:
                await reply_text(
                    teks,
                    reply_markup=button,
                    disable_web_page_preview=True,
                    quote=True,
                )
                return
            except RPCError as ef:
                await reply_text(
                    "An error has occured! Cannot parse note.",
                    quote=True,
                )
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                return
        else:
            await reply_text(teks, quote=True, disable_web_page_preview=True)
            return
    elif msgtype in (
        Types.STICKER,
        Types.VIDEO_NOTE,
        Types.CONTACT,
        Types.ANIMATED_STICKER,
    ):
        await (await send_cmd(c, msgtype))(
            m.chat.id,
            getnotes["fileid"],
            reply_to_message_id=reply_msg_id,
        )
    else:
        if getnotes["note_value"]:
            teks, button = await parse_button(text)
            button = await build_keyboard(button)
            button = InlineKeyboardMarkup(button) if button else None
        else:
            teks = ""
            button = None
        if button:
            try:
                await (await send_cmd(c, msgtype))(
                    m.chat.id,
                    getnotes["fileid"],
                    caption=teks,
                    reply_markup=button,
                    reply_to_message_id=reply_msg_id,
                )
                return
            except RPCError as ef:
                await m.reply_text(
                    teks,
                    reply_markup=button,
                    disable_web_page_preview=True,
                    reply_to_message_id=reply_msg_id,
                )
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                return
        else:
            await (await send_cmd(c, msgtype))(
                m.chat.id,
                getnotes["fileid"],
                caption=teks,
            )
    LOGGER.info(
        f"{m.from_user.id} fetched note {note_name} (type - {getnotes}) in {m.chat.id}",
    )
    return


async def get_raw_note(c: Alita, m: Message, note: str):
    """Get the note in raw format, so it can updated by just copy and pasting."""
    all_notes = [i[0] for i in db.get_all_notes(m.chat.id)]

    if note not in all_notes:
        await m.reply_text("This note does not exists!")
        return

    getnotes = db.get_note(m.chat.id, note)

    msgtype = getnotes["msgtype"]
    if not getnotes:
        await m.reply_text("<b>Error:</b> Cannot find a type for this note!!")
        return

    if msgtype == Types.TEXT:
        teks = getnotes["note_value"]
        await m.reply_text(teks, parse_mode=None, quote=True)
    elif msgtype in (
        Types.STICKER,
        Types.VIDEO_NOTE,
        Types.CONTACT,
        Types.ANIMATED_STICKER,
    ):
        await (await send_cmd(c, msgtype))(
            m.chat.id,
            getnotes["fileid"],
            reply_to_message_id=m.message_id,
        )
    else:
        if getnotes["note_value"]:
            teks = getnotes["note_value"]
        else:
            teks = ""
        await (await send_cmd(c, msgtype))(
            m.chat.id,
            getnotes["fileid"],
            caption=teks,
            parse_mode=None,
        )
    LOGGER.info(
        f"{m.from_user.id} fetched raw note {note} (type - {getnotes}) in {m.chat.id}",
    )
    return


@Alita.on_message(filters.regex(r"^#[^\s]+") & filters.group)
async def hash_get(c: Alita, m: Message):

    # If not from user, then return
    if not m.from_user:
        return

    try:
        note = (m.text[1:]).lower()
    except TypeError:
        return

    all_notes = [i[0] for i in db.get_all_notes(m.chat.id)]

    if note not in all_notes:
        # don't reply to all messages starting with #
        return

    priv_notes_status = db_settings.get_privatenotes(m.chat.id)
    await get_note_func(c, m, note, priv_notes_status)
    return


@Alita.on_message(filters.command("get", PREFIX_HANDLER) & filters.group)
async def get_note(c: Alita, m: Message):
    if len(m.text.split()) == 2:
        priv_notes_status = db_settings.get_privatenotes(m.chat.id)
        note = ((m.text.split())[1]).lower()
        all_notes = [i[0] for i in db.get_all_notes(m.chat.id)]

        if note not in all_notes:
            await m.reply_text("This note does not exists!")
            return

        await get_note_func(c, m, note, priv_notes_status)
    elif len(m.text.split()) == 3 and (m.text.split())[2] in ("noformat", "raw"):
        note = ((m.text.split())[1]).lower()
        await get_raw_note(c, m, note)
    else:
        await m.reply_text("Give me a note tag!")
        return

    return


@Alita.on_message(
    filters.command(["privnotes", "privatenotes"], PREFIX_HANDLER) & admin_filter,
)
async def priv_notes(_, m: Message):

    chat_id = m.chat.id
    if len(m.text.split()) == 2:
        option = (m.text.split())[1]
        if option in ("on", "yes"):
            db_settings.set_privatenotes(chat_id, True)
            LOGGER.info(f"{m.from_user.id} enabled privatenotes in {m.chat.id}")
            msg = "Set private notes to On"
        elif option in ("off", "no"):
            db_settings.set_privatenotes(chat_id, False)
            LOGGER.info(f"{m.from_user.id} disabled privatenotes in {m.chat.id}")
            msg = "Set private notes to Off"
        else:
            msg = "Enter correct option"
        await m.reply_text(msg)
    elif len(m.text.split()) == 1:
        curr_pref = db_settings.get_privatenotes(m.chat.id)
        msg = msg = f"Private Notes: {curr_pref}"
        LOGGER.info(f"{m.from_user.id} fetched privatenotes preference in {m.chat.id}")
        await m.reply_text(msg)
    else:
        await m.replt_text("Check help on how to use this command!")

    return


@Alita.on_message(filters.command(["notes", "saved"], PREFIX_HANDLER) & filters.group)
async def local_notes(_, m: Message):
    LOGGER.info(f"{m.from_user.id} listed all notes in {m.chat.id}")
    getnotes = db.get_all_notes(m.chat.id)
    if not getnotes:
        await m.reply_text(f"There are no notes in <b>{m.chat.title}</b>.")
        return

    curr_pref = db_settings.get_privatenotes(m.chat.id)
    if curr_pref:
        from alita import BOT_USERNAME

        pm_kb = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        "All Notes",
                        url=f"https://t.me/{BOT_USERNAME}?start=notes_{m.chat.id}",
                    ),
                ],
            ],
        )
        await m.reply_text(
            "Click on the button below to get notes!",
            quote=True,
            reply_markup=pm_kb,
        )
        return

    rply = f"Notes in <b>{m.chat.title}</b>:\n"
    for x in getnotes:
        rply += f"- <code>{x[0]}</code>\n"

    await m.reply_text(rply)
    return


@Alita.on_message(filters.command("clear", PREFIX_HANDLER) & admin_filter)
async def clear_note(_, m: Message):

    if len(m.text.split()) <= 1:
        await m.reply_text("What do you want to clear?")
        return

    note = m.text.split()[1]
    getnote = db.rm_note(m.chat.id, note)
    LOGGER.info(f"{m.from_user.id} cleared note ({note}) in {m.chat.id}")
    if not getnote:
        await m.reply_text("This note does not exist!")
        return

    await m.reply_text(f"Note '`{note}`' deleted!")
    return


@Alita.on_message(filters.command("clearall", PREFIX_HANDLER) & owner_filter)
async def clear_allnote(_, m: Message):

    all_notes = [i[0] for i in db.get_all_notes(m.chat.id)]
    if not all_notes:
        await m.reply_text("No notes are there in this chat")
        return

    await m.reply_text(
        "Are you sure you want to clear all notes?",
        reply_markup=InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        "⚠️ Confirm",
                        callback_data=f"clear_notes.{m.from_user.id}.{m.from_user.first_name}",
                    ),
                    InlineKeyboardButton("❌ Cancel", callback_data="close"),
                ],
            ],
        ),
    )
    return


@Alita.on_callback_query(filters.regex("^clear_notes."))
async def clearallnotes_callback(_, q: CallbackQuery):
    user_id = q.data.split(".")[-2]
    name = q.data.split(".")[-1]
    user_status = (await q.message.chat.get_member(user_id)).status
    if user_status != "creator":
        await q.message.edit(
            (
                f"You're an admin {await mention_html(name, user_id)}, not owner!\n"
                "Stay in your limits!"
            ),
        )
        return
    db.rm_all_notes(q.message.chat.id)
    LOGGER.info(f"{user_id} removed all notes in {q.message.chat.id}")
    await q.message.delete()
    await q.answer("Cleared all notes!", show_alert=True)
    return


__PLUGIN__ = "plugins.notes.main"
__help__ = "plugins.notes.help"
__alt_name__ = ["groupnotes", "snips"]
