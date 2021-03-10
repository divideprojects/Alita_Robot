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


from time import gmtime, strftime, time
from traceback import format_exc

from pyrogram import filters
from pyrogram.types import Message

from alita import LOGGER, PREFIX_HANDLER
from alita.bot_class import Alita
from alita.database.afk_db import AFK
from alita.tr_engine import tlang
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html

# Initialise
db = AFK()

__PLUGIN__ = "plugins.afk.main"
__help__ = "plugins.afk.help"


@Alita.on_message(
    filters.command("afk", PREFIX_HANDLER) & filters.group,
)
async def set_afk(_, m: Message):

    afkmsg = (tlang(m, "afk.user_now_afk")).format(
        user=(await mention_html(m.from_user.first_name, m.from_user.id)),
    )

    if len(m.command) > 1:
        reason = m.text.split(None, 1)[1]
        reason_txt = (tlang(m, "afk.reason")).format(res=reason)
    else:
        reason_txt = ""

    try:
        db.add_afk(m.from_user.id, time(), reason)
        await m.reply_text(afkmsg + reason_txt)
    except Exception as ef:
        await m.reply_text(ef)
        LOGGER.error(ef)
        LOGGER.error(format_exc())

    await m.stop_propagation()


@Alita.on_message(filters.group & ~filters.bot, group=11)
async def afk_mentioned(c: Alita, m: Message):

    # if msg isn't from user, ignore nd return
    if not m.from_user:
        return

    try:
        user_id, user_first_name = await extract_user(c, m)
    except Exception as ef:
        LOGGER.error(ef)
        LOGGER.error(format_exc())
        return

    try:
        user_afk = db.check_afk(user_id)
    except Exception as ef:
        LOGGER.error(format_exc())
        await m.reply_text((tlang(m, "afk.error_check_afk")).format(ef=ef))
        return

    if not user_afk:
        return

    since = strftime("%Hh %Mm %Ss", gmtime(time() - user_afk["time"]))
    afkmsg = (tlang(m, "afk.user_afk")).format(
        first_name=user_first_name,
        since=since,
    )

    if user_afk["reason"]:
        afkmsg += (tlang(m, "afk.user_afk")).format(reason=user_afk["reason"])

    await m.reply_text(afkmsg)

    await m.stop_propagation()


@Alita.on_message(filters.group, group=12)
async def rem_afk(c: Alita, m: Message):

    if not m.from_user:
        return

    try:
        user_afk = db.check_afk(m.from_user.id)
    except Exception as ef:
        LOGGER.error(format_exc())
        await m.reply_text((tlang(m, "afk.error_check_afk")).format(ef=ef))
        return

    if not user_afk:
        return

    since = strftime("%Hh %Mm %Ss", gmtime(time() - (user_afk["time"])))
    db.remove_afk(m.from_user.id)
    user = await c.get_users(user_afk["user_id"])
    await m.reply_text(
        (tlang(m, "afk.no_longer_afk")).format(
            first_name=user.first_name,
            since=since,
        ),
    )

    return
