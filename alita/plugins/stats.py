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
from alita.database import db
from alita.database.antichannelpin_db import Pins
from alita.database.antispam_db import GBan
from alita.database.approve_db import Approve
from alita.database.blacklist_db import Blacklist
from alita.database.chats_db import Chats
from alita.database.filters_db import Filters
from alita.database.notes_db import Notes, NotesSettings
from alita.database.rules_db import Rules
from alita.database.users_db import Users
from alita.database.warns_db import Warns, WarnSettings
from alita.utils.custom_filters import dev_filter

# initialise
bldb = Blacklist()
gbandb = GBan()
notesdb = Notes()
rulesdb = Rules()
userdb = Users()
appdb = Approve()
chatdb = Chats()
fldb = Filters()
pinsdb = Pins()
notesettings_db = NotesSettings()
warns_db = Warns()
warns_settings_db = WarnSettings()


@Alita.on_message(filters.command("stats", DEV_PREFIX_HANDLER) & dev_filter)
async def get_stats(_, m: Message):
    replymsg = await m.reply_text("<b><i>Fetching Stats...</i></b>", quote=True)
    rply = (
        f"<b>Users:</b> <code>{(userdb.count_users())}</code> in <code>{(chatdb.count_chats())}</code> chats\n"
        f"<b>Anti Channel Pin:</b> <code>{(pinsdb.count_chats('antichannelpin'))}</code> enabled chats\n"
        f"<b>Clean Linked:</b> <code>{(pinsdb.count_chats('cleanlinked'))}</code> enabled chats\n"
        f"<b>Filters:</b> <code>{(fldb.count_filters_all())}</code> in <code>{(fldb.count_filters_chats())}</code> chats\n"
        f"    <b>Aliases:</b> <code>{(fldb.count_filter_aliases())}</code>\n"
        f"<b>Blacklists:</b> <code>{(bldb.count_blacklists_all())}</code> in <code>{(bldb.count_blackists_chats())}</code> chats\n"
        f"    <b>Action Specific:</b>\n"
        f"        <b>None:</b> <code>{(bldb.count_action_bl_all('none'))}</code> chats\n"
        f"        <b>Kick</b> <code>{(bldb.count_action_bl_all('kick'))}</code> chats\n"
        f"        <b>Warn:</b> <code>{(bldb.count_action_bl_all('warn'))}</code> chats\n"
        f"        <b>Ban</b> <code>{(bldb.count_action_bl_all('ban'))}</code> chats\n"
        f"<b>Rules:</b> Set in <code>{(rulesdb.count_chats())}</code> chats\n"
        f"    <b>Private Rules:</b> <code>{(rulesdb.count_privrules_chats())}</code> chats\n"
        f"<b>Warns:</b> <code>{(warns_db.count_warned_users())}</code> in <code>{(warns_db.count_all_chats_using_warns())}</code> chats\n"
        f"    <b>Users Warned:</b> <code>{(warns_db.count_warned_users())}</code> users\n"
        f"    <b>Action Specific:</b>\n"
        f"        <b>Kick</b>: <code>{(warns_settings_db.count_action_chats('kick'))}</code>\n"
        f"        <b>Mute</b>: <code>{(warns_settings_db.count_action_chats('mute'))}</code>\n"
        f"        <b>Ban</b>: <code>{warns_settings_db.count_action_chats('ban')}</code>\n"
        f"<b>Notes:</b> <code>{(notesdb.count_all_notes())}</code> in <code>{(notesdb.count_notes_chats())}</code> chats\n"
        f"    <b>Private Notes:</b> <code>{(notesettings_db.count_chats())}</code> chats\n"
        f"<b>GBanned Users:</b> <code>{(gbandb.count_gbans())}</code>\n"
        f"<b>Approved People</b>: <code>{(appdb.count_all_approved())}</code> in <code>{(appdb.count_approved_chats())}</code> chats\n"
        f"\n<b>Database Stats:</b>\n"
        f"<code>{(db.command('dbstats'))}</code>"
    )
    await replymsg.edit_text(rply, parse_mode="html")

    return
