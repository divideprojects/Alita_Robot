from secrets import choice
from traceback import format_exc

from pyrogram import filters
from pyrogram.enums import ChatMemberStatus
from pyrogram.errors import RPCError
from pyrogram.types import CallbackQuery, InlineKeyboardMarkup, Message

from alita import LOGGER
from alita.bot_class import Alita
from alita.database.notes_db import Notes, NotesSettings
from alita.utils.cmd_senders import send_cmd
from alita.utils.custom_filters import admin_filter, command, owner_filter
from alita.utils.kbhelpers import ikb
from alita.utils.msg_types import Types, get_note_type
from alita.utils.string import (
    build_keyboard,
    escape_mentions_using_curly_brackets,
    parse_button,
)
from alita.vars import Config

# Initialise
db = Notes()
db_settings = NotesSettings()


@Alita.on_message(command("save") & admin_filter & ~filters.bot)
async def save_note(_, m: Message):
    existing_notes = {i[0] for i in db.get_all_notes(m.chat.id)}
    name, text, data_type, content = await get_note_type(m)
    total_notes = db.get_all_notes(m.chat.id)

    if len(total_notes) >= 1000:
        await m.reply_text(
            "Only 1000 Notes are allowed per chat!\nTo add more Notes, remove the existing ones.",
        )
        return

    if not name:
        await m.reply_text(
            f"<code>{m.text}</code>\n\nError: You must give a name for this note!",
        )
        return
    note_name = name.lower()
    if note_name in existing_notes:
        await m.reply_text(f"This note ({note_name}) already exists!")
        return

    if note_name.startswith("<") or note_name.startswith(">"):
        await m.reply_text("Cannot save a note which starts with '<' or '>'")
        return

    if not m.reply_to_message and data_type == Types.TEXT and len(m.text.split()) < 3:
        await m.reply_text(f"<code>{m.text}</code>\n\nError: There is no text in here!")
        return

    if not data_type:
        await m.reply_text(
            f"<code>{m.text}</code>\n\nError: There is no data in here!",
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
    reply_msg_id = m.reply_to_message.id if m.reply_to_message else m.id
    if m and not m.from_user:
        return

    if priv_notes_status:

        note_hash = next(i[1] for i in db.get_all_notes(m.chat.id) if i[0] == note_name)
        await reply_text(
            f"Click on the button to get the note <code>{note_name}</code>",
            reply_markup=ikb(
                [
                    [
                        (
                            "Click Me!",
                            f"https://t.me/{Config.BOT_USERNAME}?start=note_{m.chat.id}_{note_hash}",
                            "url",
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
        "id",
        "username",
        "mention",
        "chatname",
    ]
    text = await escape_mentions_using_curly_brackets(m, note_reply, parse_words)
    teks, button = await parse_button(text)
    button = await build_keyboard(button)
    button = InlineKeyboardMarkup(button) if button else None
    textt = teks

    try:
        if msgtype == Types.TEXT:
            if button:
                try:
                    await reply_text(
                        textt,
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
                await reply_text(
                    textt,
                    quote=True,
                    disable_web_page_preview=True,
                )
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
                reply_markup=button,
                reply_to_message_id=reply_msg_id,
            )
        elif button:
            try:
                await (await send_cmd(c, msgtype))(
                    m.chat.id,
                    getnotes["fileid"],
                    caption=textt,
                    reply_markup=button,
                    reply_to_message_id=reply_msg_id,
                )
                return
            except RPCError as ef:
                await m.reply_text(
                    textt,
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
                caption=textt,
                reply_markup=button,
                reply_to_message_id=reply_msg_id,
            )
        LOGGER.info(
            f"{m.from_user.id} fetched note {note_name} (type - {getnotes}) in {m.chat.id}",
        )
    except Exception as e:
        await m.reply_text(f"Error in notes: {e}")
    return


async def get_raw_note(c: Alita, m: Message, note: str):
    """Get the note in raw format, so it can updated by just copy and pasting."""
    all_notes = {i[0] for i in db.get_all_notes(m.chat.id)}
    if m and not m.from_user:
        return

    if note not in all_notes:
        await m.reply_text("This note does not exists!")
        return

    getnotes = db.get_note(m.chat.id, note)
    msg_id = m.reply_to_message.id if m.reply_to_message else m.id

    msgtype = getnotes["msgtype"]
    if not getnotes:
        await m.reply_text("<b>Error:</b> Cannot find a type for this note!!")
        return

    if msgtype == Types.TEXT:
        teks = getnotes["note_value"]
        await m.reply_text(teks, parse_mode=None, reply_to_message_id=msg_id)
    elif msgtype in (
        Types.STICKER,
        Types.VIDEO_NOTE,
        Types.CONTACT,
        Types.ANIMATED_STICKER,
    ):
        await (await send_cmd(c, msgtype))(
            m.chat.id,
            getnotes["fileid"],
            reply_to_message_id=msg_id,
        )
    else:
        teks = getnotes["note_value"] or ""
        await (await send_cmd(c, msgtype))(
            m.chat.id,
            getnotes["fileid"],
            caption=teks,
            parse_mode=None,
            reply_to_message_id=msg_id,
        )
    LOGGER.info(
        f"{m.from_user.id} fetched raw note {note} (type - {getnotes}) in {m.chat.id}",
    )
    return


@Alita.on_message(filters.regex(r"^#[^\s]+") & filters.group & ~filters.bot)
async def hash_get(c: Alita, m: Message):
    # If not from user, then return

    try:
        note = (m.text[1:]).lower()
    except TypeError:
        return

    all_notes = {i[0] for i in db.get_all_notes(m.chat.id)}

    if note not in all_notes:
        # don't reply to all messages starting with #
        return

    priv_notes_status = db_settings.get_privatenotes(m.chat.id)
    await get_note_func(c, m, note, priv_notes_status)
    return


@Alita.on_message(command("get") & filters.group & ~filters.bot)
async def get_note(c: Alita, m: Message):

    if len(m.text.split()) == 2:
        priv_notes_status = db_settings.get_privatenotes(m.chat.id)
        note = ((m.text.split())[1]).lower()
        all_notes = {i[0] for i in db.get_all_notes(m.chat.id)}

        if note not in all_notes:
            await m.reply_text("This note does not exists!")
            return

        await get_note_func(c, m, note, priv_notes_status)
    elif len(m.text.split()) == 3 and (m.text.split())[2] in ["noformat", "raw"]:
        note = ((m.text.split())[1]).lower()
        await get_raw_note(c, m, note)
    else:
        await m.reply_text("Give me a note tag!")
        return

    return


@Alita.on_message(command(["privnotes", "privatenotes"]) & admin_filter & ~filters.bot)
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
        msg = f"Private Notes: {curr_pref}"
        LOGGER.info(f"{m.from_user.id} fetched privatenotes preference in {m.chat.id}")
        await m.reply_text(msg)
    else:
        await m.reply_text("Check help on how to use this command!")

    return


@Alita.on_message(command("notes") & filters.group & ~filters.bot)
async def local_notes(_, m: Message):
    LOGGER.info(f"{m.from_user.id} listed all notes in {m.chat.id}")
    getnotes = db.get_all_notes(m.chat.id)

    if not getnotes:
        await m.reply_text(f"There are no notes in <b>{m.chat.title}</b>.")
        return

    msg_id = m.reply_to_message.id if m.reply_to_message else m.id

    if db_settings.get_privatenotes(m.chat.id):
        pm_kb = ikb(
            [
                [
                    (
                        "All Notes",
                        f"https://t.me/{Config.BOT_USERNAME}?start=notes_{m.chat.id}",
                        "url",
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
        rply += f"-> <code>#{x[0]}</code>\n"
    rply += "\nYou can get a note by #notename or <code>/get notename</code>"

    await m.reply_text(rply, reply_to_message_id=msg_id)
    return


@Alita.on_message(command("clear") & admin_filter & ~filters.bot)
async def clear_note(_, m: Message):

    if len(m.text.split()) <= 1:
        await m.reply_text("What do you want to clear?")
        return

    note = m.text.split()[1].lower()
    getnote = db.rm_note(m.chat.id, note)
    LOGGER.info(f"{m.from_user.id} cleared note ({note}) in {m.chat.id}")
    if not getnote:
        await m.reply_text("This note does not exist!")
        return

    await m.reply_text(f"Note '`{note}`' deleted!")
    return


@Alita.on_message(command("clearall") & owner_filter & ~filters.bot)
async def clear_allnote(_, m: Message):

    all_notes = {i[0] for i in db.get_all_notes(m.chat.id)}
    if not all_notes:
        await m.reply_text("No notes are there in this chat")
        return

    await m.reply_text(
        "Are you sure you want to clear all notes?",
        reply_markup=ikb(
            [[("⚠️ Confirm", "clear_notes"), ("❌ Cancel", "close_admin")]],
        ),
    )
    return


@Alita.on_callback_query(filters.regex("^clear_notes$"))
async def clearallnotes_callback(_, q: CallbackQuery):
    user_id = q.from_user.id
    user_status = (await q.message.chat.get_member(user_id)).status
    if user_status not in {ChatMemberStatus.OWNER, ChatMemberStatus.ADMINISTRATOR}:
        await q.answer(
            "You're not even an admin, don't try this explosive shit!",
            show_alert=True,
        )
        return
    if user_status != ChatMemberStatus.OWNER:
        await q.answer(
            "You're just an admin, not owner\nStay in your limits!",
            show_alert=True,
        )
        return
    db.rm_all_notes(q.message.chat.id)
    LOGGER.info(f"{user_id} removed all notes in {q.message.chat.id}")
    await q.message.edit_text("Cleared all notes!")
    return


__PLUGIN__ = "notes"

_DISABLE_CMDS_ = ["notes"]

__alt_name__ = ["groupnotes", "snips", "notes"]
