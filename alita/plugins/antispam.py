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


from datetime import datetime
from io import BytesIO

from pyrogram import filters
from pyrogram.errors import ChatAdminRequired, RPCError, UserAdminInvalid
from pyrogram.types import Message

from alita import (
    BOT_ID,
    LOGGER,
    MESSAGE_DUMP,
    PREFIX_HANDLER,
    SUPPORT_GROUP,
    SUPPORT_STAFF,
)
from alita.bot_class import Alita
from alita.database.antispam_db import GBan
from alita.utils.custom_filters import sudo_filter
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html
from alita.utils.redis_helper import get_key

# Initialize
db = GBan()


@Alita.on_message(filters.command(["gban", "globalban"], PREFIX_HANDLER) & sudo_filter)
async def gban(c: Alita, m: Message):

    if len(m.text.split()) == 1:
        await m.reply_text("<b>How to gban?</b>\n<b>Answer:</b> `/gban user_id reason`")
        return

    if len(m.text.split()) == 2 and not m.reply_to_message:
        await m.reply_text("Please enter a reason to gban user!")
        return

    user_id, user_first_name = await extract_user(m)

    if m.reply_to_message:
        gban_reason = m.text.split(None, 1)[1]
    else:
        gban_reason = m.text.split(None, 2)[2]

    if user_id in SUPPORT_STAFF:
        await m.reply_text("This user is part of my Support!, Can't ban our own!")
        return

    if user_id == BOT_ID:
        await m.reply_text("You can't gban me nigga!\nNice Try...!")
        return

    if await db.check_gban(user_id):
        await db.update_gban_reason(user_id, gban_reason)
        await m.reply_text(
            f"Updated Gban reason to: `{gban_reason}`.",
        )
        return

    await db.add_gban(user_id, gban_reason, m.from_user.id)
    await m.reply_text(
        (
            f"Added {user_first_name} to Global Ban List.\n"
            "They will now be banned in all groups where I'm admin!"
        ),
    )
    log_msg = (
        f"#GBAN\n"
        f"<b>Originated from:</b> {m.chat.id}\n"
        f"<b>Admin:</b> {(await mention_html(m.from_user.first_name, m.from_user.id))}\n"
        f"<b>Gbanned User:</b> {(await mention_html(user_first_name, user_id))}\n"
        f"<b>Gbanned User ID:</b> {user_id}\n"
        f"<b>Event Stamp:</b> {datetime.utcnow().strftime('%H:%M - %d-%m-%Y')}"
    )
    await c.send_message(MESSAGE_DUMP, log_msg)
    try:
        # Send message to user telling that he's gbanned
        await c.send_message(
            user_id,
            (
                "You have been added to my global ban list!\n"
                f"Reason: `{gban_reason}`\n\n"
                f"Appeal Chat: @{SUPPORT_GROUP}"
            ),
        )
    except BaseException as ef:  # TO DO: Improve Error Detection
        LOGGER.error(ef)
    return


@Alita.on_message(
    filters.command(["ungban", "unglobalban", "globalunban"], PREFIX_HANDLER)
    & sudo_filter,
)
async def ungban(c: Alita, m: Message):

    if len(m.text.split()) == 1:
        await m.reply_text("Pass a user id or username as an argument!")
        return

    user_id, user_first_name = await extract_user(m)

    if user_id in SUPPORT_STAFF:
        await m.reply_text("They can't be banned, so how am I supposed to ungban them?")
        return

    if user_id == BOT_ID:
        await m.reply_text("Nice Try...!")
        return

    if await db.check_gban(user_id):
        await db.remove_gban(user_id)
        await m.reply_text(f"Removed {user_first_name} from Global Ban List.")
        log_msg = (
            f"#UNGBAN\n"
            f"<b>Originated from:</b> {m.chat.id}\n"
            f"<b>Admin:</b> {(await mention_html(m.from_user.first_name, m.from_user.id))}\n"
            f"<b>UnGbanned User:</b> {(await mention_html(user_first_name, user_id))}\n"
            f"<b>UnGbanned User ID:</b> {user_id}\n"
            f"<b>Event Stamp:</b> {datetime.utcnow().strftime('%H:%M - %d-%m-%Y')}"
        )
        await c.send_message(MESSAGE_DUMP, log_msg)
        try:
            # Send message to user telling that he's ungbanned
            await c.send_message(
                user_id,
                "You have been removed from my global ban list!\n",
            )
        except BaseException as ef:  # TODO: Improve Error Detection
            LOGGER.error(ef)
        return

    await m.reply_text("User is not gbanned!")
    return


@Alita.on_message(
    filters.command(["numgbans", "countgbans"], PREFIX_HANDLER) & sudo_filter,
)
async def gban_count(_, m: Message):
    await m.reply_text(f"Number of people gbanned {(await db.count_collection())}")
    return


@Alita.on_message(
    filters.command(["gbanlist", "globalbanlist"], PREFIX_HANDLER) & sudo_filter,
)
async def gban_list(_, m: Message):
    banned_users = await db.list_collection()

    if not banned_users:
        await m.reply_text("There aren't any gbanned users...!")
        return

    banfile = "Here are all the globally banned geys!\n"
    for user in banned_users:
        banfile += f"[x] {user['name']} - {user['user_id']}\n"
        if user["reason"]:
            banfile += f"Reason: {user['reason']}\n"

    with BytesIO(str.encode(banfile)) as f:
        f.name = "gbanlist.txt"
        await m.reply_document(
            document=f,
            caption=banfile,
        )

        return


@Alita.on_message(filters.group, group=6)
async def gban_watcher(c: Alita, m: Message):
    try:
        try:
            _banned = await db.check_gban(m.from_user.id)
        except Exception as ef:
            LOGGER.error(ef)
            return
        if _banned:
            try:
                await m.chat.kick_member(m.from_user.id)
                await m.delete(m.message_id)  # Delete users message!
                await m.reply_text(
                    (
                        f"This user ({(await mention_html(m.from_user.first_name, m.from_user.id))}) "
                        "has been banned globally!\n\n"
                        f"To get unbanned appeal at @{SUPPORT_GROUP}"
                    ),
                )
                LOGGER.info(f"Banned user {m.from_user.id} in {m.chat.id}")
                return
            except (ChatAdminRequired, UserAdminInvalid):
                # Bot not admin in group and hence cannot ban users!
                # TO-DO - Improve Error Detection
                LOGGER.info(
                    f"User ({m.from_user.id}) is admin in group {m.chat.name} ({m.chat.id})",
                )
            except RPCError as excp:
                await c.send_message(
                    MESSAGE_DUMP,
                    f"<b>Gban Watcher Error!</b>\n<b>Chat:</b> {m.chat.id}\n<b>Error:</b> <code>{excp}</code>",
                )
    except AttributeError:
        pass  # Skip attribute errors!
    return
