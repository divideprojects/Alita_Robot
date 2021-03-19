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


from re import escape as re_escape
from time import time
from traceback import format_exc

from pyrogram import filters
from pyrogram.errors import ChatAdminRequired, RPCError, UserAdminInvalid
from pyrogram.types import ChatPermissions, Message

from alita import LOGGER, MESSAGE_DUMP
from alita.bot_class import Alita
from alita.database.antichannelpin_db import AntiChannelPin
from alita.database.antispam_db import ANTISPAM_BANNED, GBan
from alita.database.approve_db import Approve
from alita.database.blacklist_db import Blacklist
from alita.database.group_blacklist import BLACKLIST_CHATS
from alita.tr_engine import tlang
from alita.utils.admin_cache import (
    ADMIN_CACHE,
    TEMP_ADMIN_CACHE_BLOCK,
    admin_cache_reload,
)
from alita.utils.parser import mention_html
from alita.utils.regex_utils import regex_searcher

# Initialise
bl_db = Blacklist()
app_db = Approve()
gban_db = GBan()
antichannel_db = AntiChannelPin()


@Alita.on_message(filters.linked_channel)
async def antichanpin(c: Alita, m: Message):
    try:
        msg_id = m.message_id
        antipin_status = antichannel_db.check_antipin(m.chat.id)
        if antipin_status:
            await c.unpin_chat_message(chat_id=m.chat.id, message_id=msg_id)
    except Exception as ef:
        LOGGER.error(ef)
        LOGGER.error(format_exc())

    return


@Alita.on_message(filters.text & filters.group, group=5)
async def bl_watcher(_, m: Message):
    global TEMP_ADMIN_CACHE_BLOCK

    # TODO - Add warn option when Warn db is added!!
    async def perform_action_blacklist(m: Message, action: str):
        if action == "kick":
            (await m.chat.kick_member(m.from_user.id, int(time() + 45)))
            await m.reply_text(
                tlang(m, "blacklist.bl_watcher.action_kick").format(
                    user=(
                        m.from_user.username
                        if m.from_user.username
                        else ("<b>" + m.from_user.first_name + "</b>")
                    ),
                ),
            )
        elif action == "ban":
            (
                await m.chat.kick_member(
                    m.from_user.id,
                )
            )
            await m.reply_text(
                tlang(m, "blacklist.bl_watcher.action_ban").format(
                    user=(
                        m.from_user.username
                        if m.from_user.username
                        else ("<b>" + m.from_user.first_name + "</b>")
                    ),
                ),
            )
        elif action == "mute":
            (
                await m.chat.restrict_member(
                    m.from_user.id,
                    ChatPermissions(
                        can_send_messages=False,
                        can_send_media_messages=False,
                        can_send_stickers=False,
                        can_send_animations=False,
                        can_send_games=False,
                        can_use_inline_bots=False,
                        can_add_web_page_previews=False,
                        can_send_polls=False,
                        can_change_info=False,
                        can_invite_users=True,
                        can_pin_messages=False,
                    ),
                )
            )
            await m.reply_text(
                tlang(m, "blacklist.bl_watcher.action_mute").format(
                    user=(
                        m.from_user.username
                        if m.from_user.username
                        else ("<b>" + m.from_user.first_name + "</b>")
                    ),
                ),
            )
        elif action == "none":
            return
        return

    # If no blacklists, then return
    chat_blacklists = bl_db.get_blacklists(m.chat.id)
    if not chat_blacklists:
        return

    # Get admins from admin_cache, reduces api calls
    try:
        admin_ids = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        TEMP_ADMIN_CACHE_BLOCK[m.chat.id] = "blacklistwatcher"
        admin_ids = await admin_cache_reload(m)

    if m.from_user.id in admin_ids:
        return

    # Get approved user from cache/database
    app_users = app_db.list_approved(m.chat.id)
    if m.from_user.id in {i[0] for i in app_users}:
        return

    # Get action for blacklist
    action = bl_db.get_action(m.chat.id)
    for trigger in chat_blacklists:
        pattern = r"( |^|[^\w])" + re_escape(trigger) + r"( |$|[^\w])"
        match = await regex_searcher(pattern, m.text.lower())
        if not match:
            continue
        if match:
            try:
                await perform_action_blacklist(m, action)
                await m.delete()
            except RPCError as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
            break
    return


@Alita.on_message(filters.user(list(ANTISPAM_BANNED)) & filters.group)
async def gban_watcher(c: Alita, m: Message):
    from alita import SUPPORT_GROUP

    try:
        _banned = gban_db.check_gban(m.from_user.id)
    except Exception as ef:
        LOGGER.error(ef)
        LOGGER.error(format_exc())
        return

    if _banned:
        try:
            await m.chat.kick_member(m.from_user.id)
            await m.delete(m.message_id)  # Delete users message!
            await m.reply_text(
                (tlang(m, "antispam.watcher_banned")).format(
                    user_gbanned=(
                        await mention_html(m.from_user.first_name, m.from_user.id)
                    ),
                    SUPPORT_GROUP=SUPPORT_GROUP,
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
        except RPCError as ef:
            await c.send_message(
                MESSAGE_DUMP,
                tlang(m, "antispam.gban.gban_error_log").format(
                    chat_id=m.chat.id,
                    ef=ef,
                ),
            )
    return


@Alita.on_message(filters.chat(BLACKLIST_CHATS))
async def bl_chats_watcher(c: Alita, m: Message):
    from alita import SUPPORT_GROUP

    await c.send_message(
        m.chat.id,
        (
            "This is a blacklisted group!\n"
            f"For Support, Join @{SUPPORT_GROUP}\n"
            "Now, I'm outta here!"
        ),
    )
    await c.leave_chat(m.chat.id)
    return
