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
from pyrogram.types import Message

from alita import DEV_PREFIX_HANDLER
from alita.bot_class import Alita
from alita.database.antispam_db import GBan as gbandb
from alita.database.blacklist_db import Blacklist as bldb
from alita.database.notes_db import Notes as notesdb
from alita.database.rules_db import Rules as rulesdb
from alita.database.users_db import Users as userdb
from alita.utils.custom_filters import dev_filter


@Alita.on_message(filters.command("stats", DEV_PREFIX_HANDLER) & dev_filter)
async def get_stats(_, m: Message):
    sm = await m.reply_text("**__Fetching Stats...__**")
    rply = (
        f"<b>Users:</b> <code>{userdb().num_users()}</code> in <code>{userdb().num_chats()}</code> chats\n"
        f"<b>Blacklists:</b> <code>{bldb().num_blacklist_filters()}</code> in <code>{bldb().num_blacklist_filter_chats()}</code> chats\n"
        f"<b>Rules:</b> Set in <code>{rulesdb().num_chats()}</code> chats\n"
        f"<b>Notes:</b> <code>{notesdb().num_notes_all()}</code> in <code>{notesdb().all_notes_chats()}</code>\n"
        f"<b>Globally Banned Users:</b> <code>{gbandb().num_gbanned_users()}</code>\n"
    )
    await sm.edit_text(rply, parse_mode="html")

    return
