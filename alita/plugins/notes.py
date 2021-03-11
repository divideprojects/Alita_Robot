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
from alita.database.notes_db import Notes
from alita.utils.custom_filters import admin_filter, owner_filter
from alita.utils.msg_types import Types, get_note_type
from alita.utils.parser import mention_html
from alita.utils.string import build_keyboard, parse_button

# Initialise
db = Notes()

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

    note_name, text, data_type, content = await get_note_type(m)

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
    await m.reply_text(f"Saved note <code>{note_name}</code>!")
    return


async def get_note_func(c: Alita, m: Message, note: str):

    getnotes = db.get_note(m.chat.id, note)
    all_notes = db.get_all_notes(m.chat.id)

    if note not in all_notes:
        await m.reply_text("Note does not exists!")
        return

    msgtype = getnotes["msgtype"]
    if not getnotes:
        await m.reply_text("This note does not exist!")
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
                await m.reply_text(teks, reply_markup=button)
                LOGGER.error(ef)
                return
        else:
            await (await send_cmd(c, msgtype))(
                m.chat.id,
                getnotes["fileid"],
                caption=teks,
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
    await get_note_func(c, m, note)
    return


@Alita.on_message(filters.command("get", PREFIX_HANDLER) & filters.group)
async def get_note(c: Alita, m: Message):
    if len(m.text.split()) >= 2:
        note = (m.text.split())[1]
    else:
        await m.reply_text("Give me a note tag!")
        return
    await get_note_func(c, m, note)
    return


@Alita.on_message(filters.command(["notes", "saved"], PREFIX_HANDLER) & filters.group)
async def local_notes(_, m: Message):
    getnotes = db.get_all_notes(m.chat.id)
    if not getnotes:
        await m.reply_text(f"There are no notes in <b>{m.chat.title}</b>.")
        return
    rply = f"Notes in <b>{m.chat.title}</b>:\n"
    for x in getnotes:
        if len(rply) >= 1800:
            await m.reply(rply)
            rply = f"Notes in <b>{m.chat.title}</b>:\n"
        rply += f"- <code>{x}</code>\n"

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

    all_notes = db.get_all_notes(m.chat.id)
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
