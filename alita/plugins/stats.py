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
from alita.database.antispam_db import GBan
from alita.database.approve_db import Approve
from alita.database.blacklist_db import Blacklist
from alita.database.chats_db import Chats
from alita.database.notes_db import Notes
from alita.database.rules_db import Rules
from alita.database.spam_protect_db import SpamProtect
from alita.database.users_db import Users
from alita.utils.custom_filters import dev_filter

# initialise
bldb = Blacklist()
gbandb = GBan()
notesdb = Notes()
rulesdb = Rules()
userdb = Users()
appdb = Approve()
chatdb = Chats()
spamdb = SpamProtect()


@Alita.on_message(filters.command("stats", DEV_PREFIX_HANDLER) & dev_filter)
async def get_stats(_, m: Message):
    replymsg = await m.reply_text("<b><i>Fetching Stats...</i></b>", quote=True)
    rply = (
        f"<b>Users:</b> <code>{(await userdb.count_users())}</code> in <code>{(await chatdb.count_chats())}</code> chats\n"
        f"<b>Blacklists:</b> <code>{(await bldb.count_blacklists_all())}</code> in <code>{(await bldb.count_blackists_chats())}</code> chats\n"
        f"<b>Rules:</b> Set in <code>{(await rulesdb.count_chats())}</code> chats\n"
        f"<b>Notes:</b> <code>{(await notesdb.count_all_notes())}</code> in <code>{(await notesdb.count_notes_chats())}</code>\n"
        f"<b>Globally Banned Users:</b> <code>{(await gbandb.count_gbans())}</code>\n"
        f"<b>Approved People</b>: <code>{(await appdb.count_all_approved())}</code>\n"
        "\n<b>Spam Protection:</b>\n"
        f"\t\t<b>CAS Enabled:</b> {(await spamdb.get_cas_enabled_chats_num())}\n"
        f"\t\t<b>UnderAttack Enabled:</b> {(await spamdb.get_attack_enabled_chats_num())}\n"
    )
    await replymsg.edit_text(rply, parse_mode="html")

    return
