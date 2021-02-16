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


from pyrogram import errors, filters
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import HELP_COMMANDS, LOGGER, OWNER_ID, PREFIX_HANDLER, VERSION
from alita.bot_class import Alita
from alita.utils.localization import GetLang
from alita.utils.redishelper import get_key


async def gen_cmds_kb():
    plugins = sorted(list(HELP_COMMANDS.keys()))
    cmds = list(plugins)
    kb = []

    while cmds:
        if cmds:
            cmd = cmds[0]
            a = [
                InlineKeyboardButton(
                    f"{cmd.capitalize()}",
                    callback_data=f"get_mod.{cmd.lower()}",
                ),
            ]
            cmds.pop(0)
        if cmds:
            cmd = cmds[0]
            a.append(
                InlineKeyboardButton(
                    f"{cmd.capitalize()}",
                    callback_data=f"get_mod.{cmd.lower()}",
                ),
            )
            cmds.pop(0)
        if cmds:
            cmd = cmds[0]
            a.append(
                InlineKeyboardButton(
                    f"{cmd.capitalize()}",
                    callback_data=f"get_mod.{cmd.lower()}",
                ),
            )
            cmds.pop(0)
        kb.append(a)
    return kb


async def back_kb(m):
    _ = GetLang(m).strs
    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            [
                InlineKeyboardButton(
                    f"« {_('general.back_btn')}",
                    callback_data="commands",
                ),
            ],
        ],
    )
    return keyboard


@Alita.on_message(
    filters.command("start", PREFIX_HANDLER) & (filters.group | filters.private),
)
async def start(c: Alita, m: Message):
    me = await get_key("BOT_USERNAME")
    _ = GetLang(m).strs
    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            [
                InlineKeyboardButton(
                    f"📚 {_('start.commands_btn')}",
                    callback_data="commands",
                ),
            ]
            + [
                InlineKeyboardButton(
                    f"ℹ️ {_('start.infos_btn')}",
                    callback_data="infos",
                ),
            ],
            [
                InlineKeyboardButton(
                    f"🌐 {_('start.language_btn')}",
                    callback_data="chlang",
                ),
            ]
            + [
                InlineKeyboardButton(
                    f"➕ {_('start.add_chat_btn')}",
                    url=f"https://t.me/{me}?startgroup=new",
                ),
            ],
            [
                InlineKeyboardButton(
                    "🗃️ Source Code",
                    url="https://github.com/Divkix/Alita_Robot",
                ),
            ],
        ],
    )
    if m.chat.type == "private":
        if errors.UserIsBlocked:
            LOGGER.warning(f"Bot blocked by {m.from_user.id}")
        await m.reply_text(
            _("start.private"),
            reply_markup=keyboard,
            reply_to_message_id=m.message_id,
        )
    else:
        await m.reply_text(_("start.group"), reply_to_message_id=m.message_id)
    return


@Alita.on_callback_query(filters.regex("^start_back$"))
async def start_back(c: Alita, m: CallbackQuery):
    me = await get_key("BOT_USERNAME")
    _ = GetLang(m).strs
    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            [
                InlineKeyboardButton(
                    f"📚 {_('start.commands_btn')}",
                    callback_data="commands",
                ),
            ]
            + [
                InlineKeyboardButton(
                    f"ℹ️ {_('start.infos_btn')}",
                    callback_data="infos",
                ),
            ],
            [
                InlineKeyboardButton(
                    f"🌐 {_('start.language_btn')}",
                    callback_data="chlang",
                ),
            ]
            + [
                InlineKeyboardButton(
                    f"➕ {_('start.add_chat_btn')}",
                    url=f"https://t.me/{me}?startgroup=new",
                ),
            ],
            [
                InlineKeyboardButton(
                    "🗃️ Source Code",
                    url="https://github.com/Divkix/Alita_Robot",
                ),
            ],
        ],
    )
    await m.message.edit_text(
        _("start.private"),
        reply_markup=keyboard,
    )
    await m.answer()
    return


@Alita.on_callback_query(filters.regex("^commands$"))
async def commands_menu(_: Alita, m: CallbackQuery):
    _ = GetLang(m).strs
    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            *(await gen_cmds_kb()),
            (await back_kb(m.message)),
        ],
    )
    await m.message.edit_text(_("general.commands_available"), reply_markup=keyboard)
    await m.answer()
    return


@Alita.on_message(filters.command("help", PREFIX_HANDLER))
async def commands_pvt(c: Alita, m: Message):
    _ = GetLang(m).strs
    me = await get_key("BOT_USERNAME")
    if m.chat.type != "private":
        priv8kb = InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        text="Help",
                        url=f"t.me/{me}?start=help",
                    ),
                ],
            ],
        )
        await m.reply_text(
            "Contact me in PM to get the list of possible commands.",
            reply_markup=priv8kb,
            reply_to_message_id=m.message_id,
        )
        return

    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            *(await gen_cmds_kb()),
            (await back_kb(m)),
        ],
    )
    await m.reply_text(_("general.commands_available"), reply_markup=keyboard)
    return


@Alita.on_callback_query(filters.regex("^get_mod."))
async def get_module_info(_: Alita, m: CallbackQuery):
    module = m.data.split(".")[1]
    await m.message.edit_text(
        HELP_COMMANDS[module],
        parse_mode="markdown",
        reply_markup=(await back_kb(m.message)),
    )
    await m.answer()
    return


@Alita.on_callback_query(filters.regex("^infos$"))
async def infos(c: Alita, m: CallbackQuery):
    _ = GetLang(m).strs
    _owner = await c.get_users(OWNER_ID)
    res = _("start.info_page").format(
        Owner=(
            f"{_owner.first_name} + {_owner.last_name}"
            if _owner.last_name
            else _owner.first_name
        ),
        ID=OWNER_ID,
        version=VERSION,
    )
    await m.message.edit_text(
        res,
        reply_markup=(await back_kb(m.message)),
        disable_web_page_preview=True,
    )
    await m.answer()
    return
