from alita.__main__ import Alita
from pyrogram import filters
from pyrogram.types import (
    InlineKeyboardMarkup,
    InlineKeyboardButton,
    Message,
    CallbackQuery,
)
from alita import PREFIX_HANDLER, LOGGER
from alita.utils.msg_types import Types, get_note_type
from alita.utils.string import parse_button, build_keyboard
from alita.db import notes_db as db
from alita.utils.admin_check import admin_check, owner_check


__PLUGIN__ = "Notes"
__help__ = """
Save a note, get that, even you can delete that note.
This note only avaiable for yourself only!
Also notes support inline button powered by inline query assistant bot.

**Save Note**
/save <note>
Save a note, you can get or delete that later.

**Get Note**
/get <note>
Get that note, if avaiable.

**Delete Note**
/clear <note>
Delete that note, if avaiable.

/clearall
Clears all notes in the chat!
**NOTE:** Can only be used by owner of chat!

**All Notes**
/saved or /notes
Get all your notes, if too much notes, please use this in your saved message instead!


── **Note Format** ──
-> **Button**
`[Button Text](buttonurl:skuzzers.xyz)`
-> **Bold**
`**Bold**`
-> __Italic__
`__Italic__`
-> `Code`
`Code` (grave accent)
"""

GET_FORMAT = {
    Types.TEXT.value: Alita.send_message,
    Types.DOCUMENT.value: Alita.send_document,
    Types.PHOTO.value: Alita.send_photo,
    Types.VIDEO.value: Alita.send_video,
    Types.STICKER.value: Alita.send_sticker,
    Types.AUDIO.value: Alita.send_audio,
    Types.VOICE.value: Alita.send_voice,
    Types.VIDEO_NOTE.value: Alita.send_video_note,
    Types.ANIMATION.value: Alita.send_animation,
    Types.ANIMATED_STICKER.value: Alita.send_sticker,
    Types.CONTACT: Alita.send_contact,
}


@Alita.on_message(filters.command("save", PREFIX_HANDLER) & filters.group)
async def save_note(c: Alita, m: Message):

    res = await admin_check(c, m)
    if not res:
        return

    note_name, text, data_type, content = await get_note_type(m)

    if not note_name:
        await m.reply_text(
            "```" + m.text + "```\n\nError: You must give a name for this note!"
        )
        return

    if data_type == Types.TEXT:
        teks, _ = await parse_button(text)
        if not teks:
            await m.reply_text(
                "```" + m.text + "```\n\nError: There is no text in here!"
            )
            return

    db.save_note(str(m.chat.id), note_name, text, data_type, content)
    await m.reply_text(f"Saved note `{note_name}`!")
    return


@Alita.on_message(filters.command("get", PREFIX_HANDLER) & filters.group)
async def get_note(c: Client m: Message):
    if len(m.text.split()) >= 2:
        note = m.text.split()[1]
    else:
        await m.reply_text("Give me a note tag!")

    getnotes = db.get_note(m.chat.id, note)
    if not getnotes:
        await m.reply_text("This note does not exist!")
        return

    if getnotes["type"] == Types.TEXT:
        teks, button = await parse_button(getnotes.get("value"))
        button = await build_keyboard(button)
        button = InlineKeyboardMarkup(button) if button else None
        if button:
            try:
                await m.reply_text(teks, reply_markup=button)
                return
            except Exception as ef:
                await m.reply_text("An error has accured! Cannot parse note.")
                LOGGER.error(ef)
                return
        else:
            await m.reply_text(teks)
            return
    elif getnotes["type"] in (
        Types.STICKER,
        Types.VOICE,
        Types.VIDEO_NOTE,
        Types.CONTACT,
        Types.ANIMATED_STICKER,
    ):
        await GET_FORMAT[getnotes["type"]](m.chat.id, getnotes["file"])
    else:
        if getnotes.get("value"):
            teks, button = await parse_button(getnotes.get("value"))
            button = await build_keyboard(button)
            button = InlineKeyboardMarkup(button) if button else None
        else:
            teks = None
            button = None
        if button:
            try:
                await m.reply_text(teks, reply_markup=button)
                return
            except Exception as ef:
                await m.reply_text("An error has accured! Cannot parse note.")
                LOGGER.error(ef)
                return
        else:
            await GET_FORMAT[getnotes["type"]](
                m.chat.id, getnotes["file"], caption=teks
            )
    return


@Alita.on_message(filters.command(["notes", "saved"], PREFIX_HANDLER) & filters.group)
async def local_notes(c: Client m: Message):
    getnotes = db.get_all_notes(m.chat.id)
    if not getnotes:
        await m.reply_text(f"There are no notes in <b>{m.chat.title}</b>.")
        return
    rply = f"**Notes in <b>{m.chat.title}</b>:**\n"
    for x in getnotes:
        if len(rply) >= 1800:
            await m.reply(rply)
            rply = f"**Notes in <b>{m.chat.title}</b>:**\n"
        rply += f"- `{x}`\n"

    await m.reply_text(rply)
    return


@Alita.on_message(filters.command("clear", PREFIX_HANDLER) & filters.group)
async def clear_note(c: Alita, m: Message):

    res = await admin_check(c, m)
    if not res:
        return

    if len(m.text.split()) <= 1:
        await m.reply_text("What do you want to clear?")
        return

    note = m.text.split()[1]
    getnote = db.rm_note(m.chat.id, note)
    if not getnote:
        await m.reply_text("This note does not exist!")
        return

    await m.reply_text(f"Deleted note `{note}`!")
    return


@Alita.on_message(filters.command("clearall", PREFIX_HANDLER) & filters.group)
async def clear_allnote(c: Alita, m: Message):

    res = await owner_check(c, m)
    if not res:
        return

    all_notes = db.get_all_notes(m.chat.id)
    if not all_notes:
        await m.reply_text("No notes are there in this chat")
        return

    await m.reply_text(
        "Are you sure you want to clear all notes?",
        reply_markup=InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton("⚠️ Confirm", callback_data="clear.notes"),
                    InlineKeyboardButton("❌ Cancel", callback_data="close"),
                ]
            ]
        ),
    )
    return


@Alita.on_callback_query(filters.regex("^clear.notes$"))
async def clearallnotes_callback(q: CallbackQuery):
    await q.message.edit_text("Clearing all notes...!")
    db.rm_all_note(q.message.chat.id)
    await q.message.edit_text("Cleared all notes!")
    await q.answer()
    return
