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

from alita import PREFIX_HANDLER
from alita.bot_class import Alita
from alita.database.spam_protect_db import SpamProtect
from alita.utils.custom_filters import admin_filter

# Initialise
db = SpamProtect()

__PLUGIN__ = "plugins.spam_protect.main"
__help__ = "plugins.spam_protect.help"


@Alita.on_message(
    filters.command("cas", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def cas_protect(_, m: Message):
    get_cas = db.get_cas_status(m.chat.id)

    if len(m.text.split()) == 2:
        new_s = m.text.split(None, 1)[1]
        if new_s.lower() in ("yes", "on", "true"):
            yn = True
        elif new_s.lower() in ("no", "off", "false"):
            yn = False
        else:
            await m.reply_text(
                ("Please use an option out of:\non, yes, true or no, off, false"),
            )
            return
        db.set_cas_status(m.chat.id, yn)
        await m.reply_text(f"Set CAS Status to <code>{new_s}</code>")
    else:
        await m.reply_text(f"Your current CAS Setting is: <b>{get_cas}</b>")

    return


@Alita.on_message(
    filters.command("underattack", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def underattack(_, m: Message):
    get_a = db.get_attack_status(m.chat.id)

    if len(m.text.split()) == 2:
        new_s = m.text.split(None, 1)[1]
        if new_s.lower() in ("yes", "on", "true"):
            yn = True
        elif new_s.lower() in ("no", "off", "false"):
            yn = False
        else:
            await m.reply_text(
                ("Please use an option out of:\non, yes, true or no, off, false"),
            )
            return
        db.set_attack_status(m.chat.id, yn)
        await m.reply_text(f"Set UnderAttack Status to <code>{new_s}</code>")
    else:
        await m.reply_text(f"Your current underAttack Status is: <b>{get_a}</b>")

    return
