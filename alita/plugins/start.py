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


from pyrogram import filters, errors
from pyrogram.types import (
    CallbackQuery,
    Message,
    InlineKeyboardMarkup,
    InlineKeyboardButton,
)
from alita import PREFIX_HANDLER, VERSION, HELP_COMMANDS, OWNER_ID, LOGGER
from alita.bot_class import Alita
from alita.utils.localization import GetLang


def gen_cmds_kb():
    plugins = sorted(list(HELP_COMMANDS.keys()))
    cmds = list(plugins)
    kb = []

    while cmds:
        if cmds:
            cmd = cmds[0]
            a = [
                InlineKeyboardButton(
                    f"{cmd.capitalize()}", callback_data=f"get_mod.{cmd.lower()}"
                )
            ]
            cmds.pop(0)
        if cmds:
            cmd = cmds[0]
            a.append(
                InlineKeyboardButton(
                    f"{cmd.capitalize()}", callback_data=f"get_mod.{cmd.lower()}"
                )
            )
            cmds.pop(0)
        if cmds:
            cmd = cmds[0]
            a.append(
                InlineKeyboardButton(
                    f"{cmd.capitalize()}", callback_data=f"get_mod.{cmd.lower()}"
                )
            )
            cmds.pop(0)
        kb.append(a)
    return kb


@Alita.on_message(
    filters.command("start", PREFIX_HANDLER) & (filters.group | filters.private)
)
async def start(c: Alita, m: Message):
    me = await c.get_users("self")
    _ = GetLang(m).strs
    if m.chat.type == "private":
        if errors.UserIsBlocked:
            LOGGER.warning(f"Bot blocked by {m.from_user.id}")
        keyboard = InlineKeyboardMarkup(
            inline_keyboard=[
                [
                    InlineKeyboardButton(
                        "📚 " + _("start.commands_btn"), callback_data="commands"
                    )
                ]
                + [
                    InlineKeyboardButton(
                        "ℹ️ " + _("start.infos_btn"), callback_data="infos"
                    )
                ],
                [
                    InlineKeyboardButton(
                        "🌐  " + _("start.language_btn"), callback_data="chlang"
                    )
                ]
                + [
                    InlineKeyboardButton(
                        "➕ " + _("start.add_chat_btn"),
                        url=f"https://t.me/{me.username}?startgroup=new",
                    )
                ],
            ]
        )
        await m.reply_text(
            _("start.private"), reply_markup=keyboard, reply_to_message_id=m.message_id
        )
    else:
        await m.reply_text(_("start.group"), reply_to_message_id=m.message_id)
    return


@Alita.on_callback_query(filters.regex("^start_back$"))
async def start_back(c: Alita, m: CallbackQuery):
    me = await c.get_users("self")
    _ = GetLang(m).strs
    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            [
                InlineKeyboardButton(
                    "📚 " + _("start.commands_btn"), callback_data="commands"
                )
            ]
            + [
                InlineKeyboardButton(
                    "ℹ️ " + _("start.infos_btn"), callback_data="infos"
                )
            ],
            [
                InlineKeyboardButton(
                    "🌐 " + _("start.language_btn"), callback_data="chlang"
                )
            ]
            + [
                InlineKeyboardButton(
                    "➕ " + _("start.add_chat_btn"),
                    url=f"https://t.me/{me.username}?startgroup=new",
                )
            ],
        ]
    )
    await m.message.edit_text(_("start.private"), reply_markup=keyboard)
    await m.answer()


@Alita.on_callback_query(filters.regex("^commands$"))
async def commands_menu(m: CallbackQuery):
    _ = GetLang(m).strs
    # kb = [(await gen_cmds_kb())]
    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            *gen_cmd_kb(),
            [
                InlineKeyboardButton(
                    "« " + _("general.back_btn"), callback_data="start_back"
                )
            ],
        ]
    )
    await m.message.edit_text(_("general.commands_available"), reply_markup=keyboard)
    await m.answer()
    return


@Alita.on_message(filters.command("help", PREFIX_HANDLER))
async def commands_pvt(c: Alita, m: Message):
    me = await c.get_users("self")
    _ = GetLang(m).strs
    if m.chat.type != "private":
        priv8kb = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        text="Help", url="t.me/{}?start=help".format(me.username)
                    )
                ]
            ]
        )
        await m.reply_text(
            "Contact me in PM to get the list of possible commands.",
            reply_markup=priv8kb,
            reply_to_message_id=m.message_id,
        )
        return

    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            *gen_cmds_kb(),
            [
                InlineKeyboardButton(
                    "« " + _("general.back_btn"), callback_data="start_back"
                )
            ],
        ]
    )
    await m.reply_text(_("general.commands_available"), reply_markup=keyboard)
    return


@Alita.on_callback_query(filters.regex("^get_mod."))
async def get_module_info(m: CallbackQuery):
    _ = GetLang(m).strs
    module = m.data.split(".")[1]
    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            [
                InlineKeyboardButton(
                    "« " + _("general.back_btn"), callback_data="commands"
                )
            ]
        ]
    )
    await m.message.edit_text(
        HELP_COMMANDS[module], parse_mode="markdown", reply_markup=keyboard
    )
    await m.answer()
    return


@Alita.on_callback_query(filters.regex("^infos$"))
async def infos(c: Alita, m: CallbackQuery):
    _owner = await c.get_users(OWNER_ID)
    _ = GetLang(m).strs
    res = _("start.info_page").format(
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
                    "« " + _("general.back_btn"), callback_data="start_back"
                )
            ]
        ]
    )
    await m.message.edit_text(res, reply_markup=keyboard, disable_web_page_preview=True)
    await m.answer()
    return
