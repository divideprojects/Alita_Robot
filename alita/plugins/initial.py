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


from pyrogram.errors import RPCError
from pyrogram.types import Message

from alita import LOGGER
from alita.bot_class import Alita
from alita.database.blacklist_db import Blacklist as bldb
from alita.database.chats_db import Chats
from alita.database.lang_db import Langs
from alita.database.notes_db import Notes
from alita.database.rules_db import Rules
from alita.database.users_db import Users

# Initialise
langdb = Langs()
notedb = Notes()
ruledb = Rules()
userdb = Users()
chatdb = Chats()


@Alita.on_message(group=-1)
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
                # await migrate_chat(old_chat, new_chat)
                pass
            except RPCError as ef:
                LOGGER.error(ef)
                return
        else:
            await chatdb.update_chat(m.chat.id, m.chat.title, m.from_user.id)
            await userdb.update_user(
                m.from_user.id,
                m.from_user.first_name,
                m.from_user.username,
            )
            if m.reply_to_message:
                await chatdb.update_chat(
                    m.chat.id,
                    m.chat.title,
                    m.reply_to_message.from_user.id,
                )
                await userdb.update_user(
                    m.reply_to_message.from_user.id,
                    m.reply_to_message.from_user.first_name,
                    m.reply_to_message.from_user.username,
                )
            if m.forward_from:
                await chatdb.update_chat(
                    m.chat.id,
                    m.chat.title,
                    m.forward_from.from_user.id,
                )
                await userdb.update_user(
                    m.forward_from.id,
                    m.forward_from.first_name,
                    m.forward_from.username,
                )
    except AttributeError:
        pass  # Skip attribute errors!
    return


async def migrate_chat(old_chat, new_chat):
    LOGGER.info(f"Migrating from {str(old_chat)} to {str(new_chat)}")
    await userdb.migrate_chat(old_chat, new_chat)
    await langdb.migrate_chat(old_chat, new_chat)
    await ruledb.migrate_chat(old_chat, new_chat)
    await bldb.migrate_chat(old_chat, new_chat)
    await notedb.migrate_chat(old_chat, new_chat)
    LOGGER.info("Successfully migrated!")
