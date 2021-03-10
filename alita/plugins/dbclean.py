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


from asyncio import sleep
from traceback import format_exc

from pyrogram import filters
from pyrogram.errors import BadRequest, ChatWriteForbidden, RPCError, Unauthorized
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import DEV_PREFIX_HANDLER, LOGGER
from alita.bot_class import Alita
from alita.database.antispam_db import GBan
from alita.database.chats_db import Chats
from alita.utils.custom_filters import dev_filter

gban_db = GBan()
user_db = Chats()

__PLUGIN__ = "Database Cleaning"


async def get_invalid_chats(c: Alita, m: Message, remove: bool = False):
    chats = user_db.get_all_chats()
    kicked_chats, progress = 0, 0
    chat_list = []
    progress_message = m

    for chat in chats:
        if ((100 * chats.index(chat)) / len(chats)) > progress:
            progress_bar = f"{progress}% completed in getting invalid chats."
            if progress_message:
                try:
                    await m.edit_text(progress_bar)
                except RPCError:
                    pass
            else:
                progress_message = await m.reply_text(progress_bar)
            progress += 5

        cid = chat.chat_id
        await sleep(0.1)
        try:
            await c.get_chat(cid)
        except (BadRequest, Unauthorized):
            kicked_chats += 1
            chat_list.append(cid)
        except RPCError:
            pass

    try:
        await progress_message.delete()
    except RPCError:
        pass

    if not remove:
        return kicked_chats
    for muted_chat in chat_list:
        try:
            await sleep(0.1)
            user_db.rem_chat(muted_chat)
        except RPCError:
            pass
    return kicked_chats


async def get_invalid_gban(c: Alita, _, remove: bool = False):
    banned = gban_db.get_gban_list()
    ungbanned_users = 0
    ungban_list = []

    for user in banned:
        user_id = user["user_id"]
        await sleep(0.1)
        try:
            await c.get_users(user_id)
        except BadRequest:
            ungbanned_users += 1
            ungban_list.append(user_id)
        except RPCError:
            pass

    if remove:
        for user_id in ungban_list:
            try:
                await sleep(0.1)
                gban_db.ungban_user(user_id)
            except RPCError:
                pass

    return ungbanned_users


async def get_muted_chats(c: Alita, m: Message, leave: bool = False):
    chat_id = m.chat.id
    chats = user_db.get_all_chats()
    muted_chats, progress = 0, 0
    chat_list = []
    progress_message = m

    for chat in chats:

        if ((100 * chats.index(chat)) / len(chats)) > progress:
            progress_bar = f"{progress}% completed in getting muted chats."
            if progress_message:
                try:
                    await m.edit_text(progress_bar, chat_id)
                except RPCError:
                    pass
            else:
                progress_message = await m.edit_text(progress_bar)
            progress += 5

        cid = chat.chat_id
        await sleep(0.1)

        try:
            await c.send_chat_action(cid, "typing")
        except (BadRequest, Unauthorized, ChatWriteForbidden):
            muted_chats += 1
            chat_list.append(cid)
        except RPCError:
            pass

    try:
        await progress_message.delete()
    except RPCError:
        pass

    if not leave:
        return muted_chats
    for muted_chat in chat_list:
        await sleep(0.1)
        try:
            await c.leave_chat(muted_chat)
            user_db.rem_chat(muted_chat)
        except RPCError:
            pass
    return muted_chats


@Alita.on_message(filters.command("dbclean", DEV_PREFIX_HANDLER) & dev_filter)
async def dbcleanxyz(_, m: Message):
    buttons = [
        [InlineKeyboardButton("Invalid Chats", callback_data="dbclean.invalidchats")],
    ]
    buttons += [
        [InlineKeyboardButton("Muted Chats", callback_data="dbclean.mutedchats")],
    ]
    buttons += [[InlineKeyboardButton("Invalid Gbans", callback_data="dbclean.gbans")]]
    await m.reply_text(
        "What do you want to clean?",
        reply_markup=InlineKeyboardMarkup(buttons),
    )
    return


@Alita.on_callback_query(filters.regex("^dbclean_"))
async def dbclean_callback(c: Alita, q: CallbackQuery):
    args = q.data.split(".")
    # Invalid Chats
    if args[1] == "invalidchats":
        await q.message.edit_text("Getting Invalid Chat Count ...")
        invalid_chat_count = await get_invalid_chats(c, q.message)

        if not invalid_chat_count > 0:
            await q.message.edit_text("No Invalid Chats.")
            return

        await q.message.reply_text(
            f"Total invalid chats - {invalid_chat_count}",
            reply_markup=InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(
                            "Remove Invalid Chats",
                            callback_data="remove.inavlid_chats",
                        ),
                    ],
                ],
            ),
        )
        await q.message.delete()
        return

    # Muted Chats
    if args[1] == "mutedchats":
        await q.message.edit_text("Getting Muted Chat Count...")
        muted_chat_count = await get_muted_chats(c, q.message)

        if not muted_chat_count > 0:
            await q.message.delete()
            await q.message.edit_text("I'm not muted in any Chats.")
            return

        await q.message.reply_text(
            f"Muted Chats - {muted_chat_count}",
            reply_markup=InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(
                            "Leave Muted Chats",
                            callback_data="remove.muted_chats",
                        ),
                    ],
                ],
            ),
        )
        await q.message.delete()
        return

    # Invalid Gbans
    if args[1] == "gbans":
        await q.message.edit_text("Getting Invalid Gban Count ...")
        invalid_gban_count = await get_invalid_gban(c, q.message)

        if not invalid_gban_count > 0:
            await q.message.edit_text("No Invalid Gbans")
            return

        await q.message.reply_text(
            f"Invalid Gbans - {invalid_gban_count}",
            reply_markup=InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(
                            "Remove Invalid Gbans",
                            callback_data="remove.invalid_gbans",
                        ),
                    ],
                ],
            ),
        )
        await q.message.delete()
        return
    return


@Alita.on_callback_query(filters.regex("^remove."))
async def db_clean_callbackAction(c: Alita, q: CallbackQuery):
    try:
        args = q.data.split(".")[1]

        if args == "muted_chats":
            await q.message.edit_text("Leaving chats ...")
            chat_count = await get_muted_chats(c, q.message, True)
            await q.message.edit_text(f"Left {chat_count} chats.")

        elif args == "inavlid_chats":
            await q.message.edit_text("Cleaning up Db...")
            invalid_chat_count = await get_invalid_chats(c, q.message, True)
            await q.message.edit_text(f"Cleaned up {invalid_chat_count} chats from Db.")

        elif args == "invalid_gbans":
            await q.message.edit_text("Removing Invalid Gbans from Db...")
            invalid_gban_count = await get_invalid_gban(c, q.message, True)
            await q.message.edit_text(
                f"Cleaned up {invalid_gban_count} gbanned users from Db",
            )
    except Exception as ef:
        LOGGER.error(f"Error while cleaning db:\n{ef}")
        LOGGER.error(format_exc())
    return
