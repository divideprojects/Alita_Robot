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
from pyrogram.errors import MessageNotModified, QueryIdInvalid, RPCError, UserIsBlocked
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

import alita
from alita import HELP_COMMANDS, LOGGER, OWNER_ID, PREFIX_HANDLER, VERSION
from alita.bot_class import Alita
from alita.database.notes_db import Notes
from alita.database.rules_db import Rules
from alita.tr_engine import tlang
from alita.utils.msg_types import Types, get_note_type
from alita.utils.parser import mention_html
from alita.utils.string import build_keyboard, parse_button

# Initialize
rules_db = Rules()
notes_db = Notes()


async def gen_cmds_kb(m):
    """Generate the keyboard for languages."""
    if isinstance(m, CallbackQuery):
        m = m.message

    cmds = sorted(list(HELP_COMMANDS.keys()))
    kb = []

    while cmds:
        if cmds:
            cmd = cmds[0]
            a = [
                InlineKeyboardButton(
                    tlang(m, cmd),
                    callback_data=f"get_mod.{cmd.lower()}",
                ),
            ]
            cmds.pop(0)
        if cmds:
            cmd = cmds[0]
            a.append(
                InlineKeyboardButton(
                    tlang(m, cmd),
                    callback_data=f"get_mod.{cmd.lower()}",
                ),
            )
            cmds.pop(0)
        if cmds:
            cmd = cmds[0]
            a.append(
                InlineKeyboardButton(
                    tlang(m, cmd),
                    callback_data=f"get_mod.{cmd.lower()}",
                ),
            )
            cmds.pop(0)
        kb.append(a)
    return kb


async def gen_start_kb(q):
    """Generate keyboard with start menu options."""

    from alita import BOT_USERNAME

    keyboard = InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton(
                    f"📚 {(tlang(q, 'start.commands_btn'))}",
                    callback_data="commands",
                ),
                InlineKeyboardButton(
                    f"ℹ️ {(tlang(q, 'start.infos_btn'))}",
                    callback_data="infos",
                ),
            ],
            [
                InlineKeyboardButton(
                    f"🌐 {(tlang(q, 'start.language_btn'))}",
                    callback_data="chlang",
                ),
                InlineKeyboardButton(
                    f"➕ {(tlang(q, 'start.add_chat_btn'))}",
                    url=f"https://t.me/{BOT_USERNAME}?startgroup=new",
                ),
            ],
            [
                InlineKeyboardButton(
                    f"🗃️ {(tlang(q, 'start.source_code'))}",
                    url="https://github.com/Divkix/Alita_Robot",
                ),
            ],
        ],
    )
    return keyboard


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


async def get_private_note(c: Alita, m: Message, help_option: str):
    """Get the note in pm of user, with parsing enabled."""
    from alita import BOT_USERNAME

    help_lst = m.text.split("_")
    chat_id = help_lst[1]

    if len(help_lst) == 2:
        chat_id = help_option.replace("notes_", "")
        all_notes = notes_db.get_all_notes(int(chat_id))
        rply = f"Notes in Chat:\n\n"
        for note in all_notes:
            rply += f"- [{note[0]}](https://t.me/{BOT_USERNAME}?start=note_{chat_id}_{note[1]})\n"
        await m.reply_text(rply, disable_web_page_preview=True)
        return
    elif len(help_lst) == 3:
        note_hash = help_option.split("_")[2]
        getnotes = notes_db.get_note_by_hash(note_hash)
    else:
        return

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


async def get_private_rules(_, m: Message, help_option: str):
    chat_id = help_option.split("_")[1]
    rules = rules_db.get_rules(int(chat_id))
    await m.reply_text(
        (tlang(m, "rules.get_rules")).format(
            chat=m.chat.title,
            rules=rules,
        ),
        disable_web_page_preview=True,
    )
    return


async def get_help_msg(_, m, help_option: str):
    """Helper function for getting help_msg and it's keyboard."""

    help_msg = None
    help_kb = None

    if help_option == "help":
        help_msg = tlang(m, "general.commands_available")
        help_kb = InlineKeyboardMarkup(
            [
                *(await gen_cmds_kb(m)),
                [
                    InlineKeyboardButton(
                        f"« {(tlang(m, 'general.back_btn'))}",
                        callback_data="start_back",
                    ),
                ],
            ],
        )
    else:
        help_cmd_keys = sorted(
            [i.split(".")[1].lower() for i in list(HELP_COMMANDS.keys())],
        )
        if help_option in help_cmd_keys:
            help_option_value = HELP_COMMANDS[f"plugins.{help_option}.main"]
            help_msg = tlang(m, help_option_value)
            help_kb = InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(
                            f"« {(tlang(m, 'general.back_btn'))}",
                            callback_data="commands",
                        ),
                    ],
                ],
            )

    return help_msg, help_kb


@Alita.on_message(
    filters.command("start", PREFIX_HANDLER) & (filters.group | filters.private),
)
async def start(c: Alita, m: Message):

    if m.chat.type == "private":
        if len(m.text.split()) > 1:
            help_option = (m.text.split(None, 1)[1]).lower()

            if help_option.startswith("note"):
                await get_private_note(c, m, help_option)
                return
            elif help_option.startswith("rules"):
                await get_private_rules(c, m, help_option)
                return

            help_msg, help_kb = await get_help_msg(m, help_option)

            if help_msg is None:
                return

            await m.reply_text(
                help_msg,
                parse_mode="markdown",
                reply_markup=help_kb,
                quote=True,
            )
            return
        try:
            await m.reply_text(
                (tlang(m, "start.private")),
                reply_markup=(await gen_start_kb(m)),
                quote=True,
            )
        except UserIsBlocked:
            LOGGER.warning(f"Bot blocked by {m.from_user.id}")
    else:
        await m.reply_text(
            (tlang(m, "start.group")),
            quote=True,
        )
    return


@Alita.on_callback_query(filters.regex("^start_back$"))
async def start_back(_, q: CallbackQuery):

    try:
        await q.message.edit_text(
            (tlang(q, "start.private")),
            reply_markup=(await gen_start_kb(q.message)),
        )
    except MessageNotModified:
        pass
    await q.answer()
    return


@Alita.on_callback_query(filters.regex("^commands$"))
async def commands_menu(_, q: CallbackQuery):

    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            *(await gen_cmds_kb(q)),
            [
                InlineKeyboardButton(
                    f"« {(tlang(q, 'general.back_btn'))}",
                    callback_data="start_back",
                ),
            ],
        ],
    )
    try:
        await q.message.edit_text(
            (tlang(q, "general.commands_available")),
            reply_markup=keyboard,
        )
    except QueryIdInvalid:
        await q.message.reply_text(
            (tlang(q, "general.commands_available")),
            reply_markup=keyboard,
        )
    await q.answer()
    return


@Alita.on_message(filters.command("help", PREFIX_HANDLER))
async def help_menu(_, m: Message):

    from alita import BOT_USERNAME

    if len(m.text.split()) >= 2:
        help_option = (m.text.split(None, 1)[1]).lower()
        help_msg, help_kb = await get_help_msg(m, help_option)
        if help_msg is None:
            return
        if m.chat.type == "private":
            await m.reply_text(
                help_msg,
                parse_mode="markdown",
                reply_markup=help_kb,
                quote=True,
            )
        else:
            await m.reply_text(
                (tlang(m, "start.public_help").format(help_option=help_option)),
                reply_markup=InlineKeyboardMarkup(
                    [
                        [
                            InlineKeyboardButton(
                                "Help",
                                url=f"t.me/{BOT_USERNAME}?start={help_option}",
                            ),
                        ],
                    ],
                ),
            )
    else:
        if m.chat.type == "private":
            keyboard = InlineKeyboardMarkup(
                inline_keyboard=[
                    *(await gen_cmds_kb(m)),
                    [
                        InlineKeyboardButton(
                            f"« {(tlang(m, 'general.back_btn'))}",
                            callback_data="start_back",
                        ),
                    ],
                ],
            )
            msg = tlang(m, "general.commands_available")
        else:
            keyboard = InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(
                            text="Help",
                            url=f"t.me/{BOT_USERNAME}?start=help",
                        ),
                    ],
                ],
            )
            msg = tlang(m, "start.pm_for_help")

        await m.reply_text(
            msg,
            reply_markup=keyboard,
        )

    return


@Alita.on_callback_query(filters.regex("^get_mod."))
async def get_module_info(_, q: CallbackQuery):

    module = q.data.split(".", 1)[1]
    keyboard = InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton(
                    "« " + (tlang(q, "general.back_btn")),
                    callback_data="commands",
                ),
            ],
        ],
    )
    help_msg = tlang(q, HELP_COMMANDS[module])
    await q.message.edit_text(
        help_msg,
        parse_mode="markdown",
        reply_markup=keyboard,
    )
    await q.answer()
    return


@Alita.on_callback_query(filters.regex("^infos$"))
async def infos(c: Alita, q: CallbackQuery):

    _owner = await c.get_users(OWNER_ID)
    res = (tlang(q, "start.info_page")).format(
        Owner=(
            f"{_owner.first_name} + {_owner.last_name}"
            if _owner.last_name
            else _owner.first_name
        ),
        ID=OWNER_ID,
        version=VERSION,
    )
    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            [
                InlineKeyboardButton(
                    f"« {(tlang(q, 'general.back_btn'))}",
                    callback_data="start_back",
                ),
            ],
        ],
    )
    await q.message.edit_text(
        res,
        reply_markup=keyboard,
        disable_web_page_preview=True,
    )
    await q.answer()
    return
