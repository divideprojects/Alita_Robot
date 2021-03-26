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
from pyrogram.types import Message

from alita import LOGGER
from alita.bot_class import Alita
from alita.database.antichannelpin_db import Pins
from alita.database.antiflood_db import AntiFlood
from alita.database.approve_db import Approve
from alita.database.blacklist_db import Blacklist
from alita.database.chats_db import Chats
from alita.database.filters_db import Filters
from alita.database.lang_db import Langs
from alita.database.notes_db import Notes, NotesSettings
from alita.database.reporting_db import Reporting
from alita.database.rules_db import Rules
from alita.database.users_db import Users

# Initialise
langdb = Langs()
notedb = Notes()
ruledb = Rules()
userdb = Users()
chatdb = Chats()
bldb = Blacklist()
flooddb = AntiFlood()
approvedb = Approve()
reportdb = Reporting()
notes_settings = NotesSettings()
pins_db = Pins()
fldb = Filters()


@Alita.on_message(filters.group, group=4)
async def initial_works(_, m: Message):
    try:
        if m.migrate_to_chat_id or m.migrate_from_chat_id:
            if m.migrate_to_chat_id:
                old_chat = m.chat.id
                new_chat = m.migrate_to_chat_id
            elif m.migrate_from_chat_id:
                old_chat = m.migrate_from_chat_id
                new_chat = m.chat.id
            try:
                await migrate_chat(old_chat, new_chat)
            except RPCError as ef:
                LOGGER.error(ef)
                return
        else:
            if m.reply_to_message and not m.forward_from:
                chatdb.update_chat(
                    m.chat.id,
                    m.chat.title,
                    m.reply_to_message.from_user.id,
                )
                userdb.update_user(
                    m.reply_to_message.from_user.id,
                    (
                        f"{m.reply_to_message.from_user.first_name} {m.reply_to_message.from_user.last_name}"
                        if m.reply_to_message.from_user.last_name
                        else m.reply_to_message.from_user.first_name
                    ),
                    m.reply_to_message.from_user.username,
                )
            elif m.forward_from and not m.reply_to_message:
                chatdb.update_chat(
                    m.chat.id,
                    m.chat.title,
                    m.forward_from.id,
                )
                userdb.update_user(
                    m.forward_from.id,
                    (
                        f"{m.forward_from.first_name} {m.forward_from.last_name}"
                        if m.forward_from.last_name
                        else m.forward_from.first_name
                    ),
                    m.forward_from.username,
                )
            elif m.reply_to_message and m.forward_from:
                chatdb.update_chat(
                    m.chat.id,
                    m.chat.title,
                    m.reply_to_message.forward_from.id,
                )
                userdb.update_user(
                    m.forward_from.id,
                    (
                        f"{m.reply_to_message.forward_from.first_name} {m.reply_to_message.forward_from.last_name}"
                        if m.reply_to_message.forward_from.last_name
                        else m.reply_to_message.forward_from.first_name
                    ),
                    m.forward_from.username,
                )
            else:
                chatdb.update_chat(m.chat.id, m.chat.title, m.from_user.id)
                userdb.update_user(
                    m.from_user.id,
                    (
                        f"{m.from_user.first_name} {m.from_user.last_name}"
                        if m.from_user.last_name
                        else m.from_user.first_name
                    ),
                    m.from_user.username,
                )
    except AttributeError:
        pass  # Skip attribute errors!
    return


async def migrate_chat(old_chat, new_chat):
    LOGGER.info(f"Migrating from {old_chat} to {new_chat}...")
    userdb.migrate_chat(old_chat, new_chat)
    langdb.migrate_chat(old_chat, new_chat)
    ruledb.migrate_chat(old_chat, new_chat)
    bldb.migrate_chat(old_chat, new_chat)
    notedb.migrate_chat(old_chat, new_chat)
    flooddb.migrate_chat(old_chat, new_chat)
    approvedb.migrate_chat(old_chat, new_chat)
    reportdb.migrate_chat(old_chat, new_chat)
    notes_settings.migrate_chat(old_chat, new_chat)
    pins_db.migrate_chat(old_chat, new_chat)
    fldb.migrate_chat(old_chat, new_chat)
    LOGGER.info(f"Successfully migrated from {old_chat} to {new_chat}!")
