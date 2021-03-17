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
from alita.utils.custom_filters import admin_filter, owner_filter
from alita.utils.msg_types import Types, get_note_type
from alita.utils.parser import mention_html
from alita.utils.string import build_keyboard, parse_button

# Initialise
db = Notes()
db_settings = NotesSettings()

__PLUGIN__ = "plugins.notes.main"
__help__ = "plugins.notes.help"


async def send_cmd(client: Alita, msgtype):
    GET_FORMAT = {
        Types.TEXT.value: client.send_message,
        Types.DOCUMENT.value: client.send_document,
        Types.PHOTO.value: client.send_photo,
        Types.VIDEO.value: client.send_video,
        Types.STICKER.value: client.send_sticker,
        Types.AUDIO.value: client.send_audio,
        Types.VOICE.value: client.send_voice,
        Types.VIDEO_NOTE.value: client.send_video_note,
        Types.ANIMATION.value: client.send_animation,
        Types.ANIMATED_STICKER.value: client.send_sticker,
        Types.CONTACT: client.send_contact,
    }
    return GET_FORMAT[msgtype]


@Alita.on_message(
    filters.command("save", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def save_note(_, m: Message):

    existing_notes = [i[0] for i in db.get_all_notes(m.chat.id)]

    note_name, text, data_type, content = await get_note_type(m)

    if note_name in existing_notes:
        await m.reply_text(f"This note ({note_name}) already exists!")
        return

    if not note_name:
        await m.reply_text(
            f"<code>{m.text}</code>\n\nError: You must give a name for this note!",
        )
        return

    if data_type == Types.TEXT or text != "":
        teks, _ = await parse_button(text)
        if not teks:
            await m.reply_text(
                f"<code>{m.text}</code>\n\nError: There is no text in here!",
            )
            return

    db.save_note(m.chat.id, note_name, text, data_type, content)
    await m.reply_text(
        f"Saved note <code>{note_name}</code>!\nGet it with <code>/get {note_name}</code> or <code>#{note_name}</code>",
    )
    return


async def get_note_func(c: Alita, m: Message, note_name, priv_notes_status):
    """Get the note in normal mode, with parsing enabled."""

    if priv_notes_status:
        from alita import BOT_USERNAME

        note_hash = [i[1] for i in db.get_all_notes(m.chat.id) if i[0] == note_name][0]
        await m.reply_text(
            f"Click on the button to get the note {note_name}",
            reply_markup=InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(
                            "Note",
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
        await m.reply_text("<b>Error:</b> Cannot find a type for this note!!")
        return

    if msgtype == Types.TEXT:
        teks, button = await parse_button(getnotes["note_value"])
        button = await build_keyboard(button)
        button = InlineKeyboardMarkup(button) if button else None
        if button:
            try:
                await m.reply_text(
                    teks,
                    reply_markup=button,
                    disable_web_page_preview=True,
                )
                return
            except RPCError as ef:
                await m.reply_text("An error has occured! Cannot parse note.")
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                return
        else:
            await m.reply_text(teks)
            return
    elif msgtype in (
        Types.STICKER,
        Types.VIDEO_NOTE,
        Types.CONTACT,
        Types.ANIMATED_STICKER,
    ):
        await (await send_cmd(c, msgtype))(m.chat.id, getnotes["fileid"])
    else:
        if getnotes["note_value"]:
            teks, button = await parse_button(getnotes["note_value"])
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
                )
                return
            except RPCError as ef:
                await m.reply_text(
                    teks,
                    reply_markup=button,
                    disable_web_page_preview=True,
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
        await m.reply_text(teks, parse_mode=None)
    elif msgtype in (
        Types.STICKER,
        Types.VIDEO_NOTE,
        Types.CONTACT,
        Types.ANIMATED_STICKER,
    ):
        await (await send_cmd(c, msgtype))(
            m.chat.id,
            getnotes["fileid"],
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
    return


@Alita.on_message(filters.regex(r"^#[^\s]+") & filters.group)
async def hash_get(c: Alita, m: Message):

    # If not from user, then return
    if not m.from_user:
        return

    try:
        note = m.text[1:]
    except TypeError:
        return

    all_notes = [i[0] for i in db.get_all_notes(m.chat.id)]

    if note not in all_notes:
        # Because  - don't reply to all messages starting with #
        return

    priv_notes_status = db_settings.get_privatenotes(m.chat.id)
    await get_note_func(c, m, note, priv_notes_status)
    return


@Alita.on_message(filters.command("get", PREFIX_HANDLER) & filters.group)
async def get_note(c: Alita, m: Message):
    if len(m.text.split()) == 2:
        priv_notes_status = db_settings.get_privatenotes(m.chat.id)
        note = (m.text.split())[1]
        all_notes = [i[0] for i in db.get_all_notes(m.chat.id)]

        if note not in all_notes:
            await m.reply_text("This note does not exists!")
            return

        await get_note_func(c, m, note, priv_notes_status)
    elif len(m.text.split()) == 3 and (m.text.split())[2] in ("noformat", "raw"):
        note = (m.text.split())[1]
        await get_raw_note(c, m, note)
    else:
        await m.reply_text("Give me a note tag!")
        return

    return


@Alita.on_message(
    filters.command(["privnotes", "privatenotes"], PREFIX_HANDLER)
    & filters.group
    & admin_filter,
)
async def priv_notes(_, m: Message):

    chat_id = m.chat.id
    if len(m.text.split()) == 2:
        option = (m.text.split())[1]
        if option in ("on", "yes"):
            db_settings.set_privatenotes(chat_id, True)
            msg = "Set private notes to On"
        elif option in ("off", "no"):
            db_settings.set_privatenotes(chat_id, False)
            msg = "Set private notes to Off"
        else:
            msg = "Enter correct option"
        await m.reply_text(msg)
    elif len(m.text.split()) == 1:
        curr_pref = db_settings.get_privatenotes(m.chat.id)
        msg = msg = f"Private Notes: {curr_pref}"
        await m.reply_text(msg)
    else:
        await m.replt_text("Check help on how to use this command!")

    return


@Alita.on_message(filters.command(["notes", "saved"], PREFIX_HANDLER) & filters.group)
async def local_notes(_, m: Message):
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
                        "Notes",
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


@Alita.on_message(
    filters.command("clear", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def clear_note(_, m: Message):

    if len(m.text.split()) <= 1:
        await m.reply_text("What do you want to clear?")
        return

    note = m.text.split()[1]
    getnote = db.rm_note(m.chat.id, note)
    if not getnote:
        await m.reply_text("This note does not exist!")
        return

    await m.reply_text(f"Note '`{note}`' deleted!")
    return


@Alita.on_message(
    filters.command("clearall", PREFIX_HANDLER) & filters.group & owner_filter,
)
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
    await q.message.delete()
    await q.answer("Cleared all notes!", show_alert=True)
    return
