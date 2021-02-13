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


from time import time
from os import path, makedirs
from pyrogram import Client, __version__, errors
from pyrogram.raw.all import layer
from pyrogram.types import InlineKeyboardMarkup
from pyrogram.types.bots_and_keyboards.inline_keyboard_button import (
    InlineKeyboardButton,
)
from alita.plugins import ALL_PLUGINS
from alita.db import users_db as userdb
from alita.utils.redishelper import set_key, flushredis, allkeys
from alita.utils.paste import paste
from alita import (
    APP_ID,
    API_HASH,
    LOGGER,
    TOKEN,
    NO_LOAD,
    MESSAGE_DUMP,
    BOT_USERNAME,
    WORKERS,
    load_cmds,
    SUPPORT_STAFF,
    logfile,
    log_datetime,
    get_self,
)

# Check if MESSAGE_DUMP is correct
if MESSAGE_DUMP == -100 or not str(MESSAGE_DUMP).startswith("-100"):
    raise Exception(
        "Please enter a vaild Supergroup ID, A Supergroup ID starts with -100",
    )


class Alita(Client):
    """Starts the Pyrogram Client on the Bot Token when we do 'python3 -m alita'"""

    def __init__(self):
        name = self.__class__.__name__.lower()

        # Make a temporary direcory for storing session file
        SESSION_DIR = f"{name}/SESSION"
        if not path.isdir(SESSION_DIR):
            makedirs(SESSION_DIR)

        super().__init__(
            name,
            plugins=dict(root=f"{name}/plugins", exclude=NO_LOAD),
            workdir=SESSION_DIR,
            api_id=APP_ID,
            api_hash=API_HASH,
            bot_token=TOKEN,
            workers=WORKERS,
        )

    async def get_admins(self):
        LOGGER.info("Begin caching admins...")
        begin = time()

        # Flush Redis data
        try:
            await flushredis()
        except Exception as ef:
            LOGGER.error(ef)

        all_chats = userdb.get_all_chats() or []  # Get list of all chats
        LOGGER.info(f"{len(all_chats)} chats loaded.")
        ADMINDICT = {}
        for i in all_chats:
            chat_id = i.chat_id
            adminlist = []
            try:
                async for j in self.iter_chat_members(
                    chat_id=chat_id,
                    filter="administrators",
                ):
                    adminlist.append(j.user.id)

                ADMINDICT[str(i.chat_id)] = adminlist  # Remove the last space

                LOGGER.info(
                    f"Set {len(adminlist)} admins for {i.chat_id}\n- {adminlist}",
                )
            except errors.PeerIdInvalid:
                pass
            except Exception as ef:
                LOGGER.error(ef)

        try:
            await set_key("ADMINDICT", ADMINDICT)
            end = time()
            LOGGER.info(
                (
                    "Set admin list cache!"
                    f"Time Taken: {round(end - begin, 2)} seconds."
                ),
            )
        except Exception as ef:
            LOGGER.error(f"Could not set ADMINDICT!\n{ef}")

    async def start(self):
        await super().start()

        await get_self(self)  # Get bot info from pyrogram client
        LOGGER.info("Starting bot...")

        await self.send_message(MESSAGE_DUMP, "Starting Bot...")

        # Redis Content Setup!
        await self.get_admins()  # Load admins in cache
        await set_key("SUPPORT_STAFF", SUPPORT_STAFF)  # Load SUPPORT_STAFF in cache
        # Redis Content Setup!

        # Show in Log that bot has started
        LOGGER.info(
            f"Pyrogram v{__version__}\n(Layer - {layer}) started on @{BOT_USERNAME}",
        )
        cmd_list = await load_cmds(await ALL_PLUGINS)
        redis_keys = await allkeys()
        LOGGER.info(f"Plugins Loaded: {cmd_list}")
        LOGGER.info(f"Redis Keys Loaded: {redis_keys}")

        # Send a message to MESSAGE_DUMP telling that the
        # bot has started and has loaded all plugins!
        await self.send_message(
            MESSAGE_DUMP,
            (
                f"<b><i>Bot started on Pyrogram v{__version__} (Layer - {layer})</i></b>\n\n"
                "<b>Loaded Plugins:</b>\n"
                f"<i>{cmd_list}</i>\n"
                "<b>Redis Keys Loaded:</b>\n"
                f"<i>{redis_keys}</i>"
            ),
        )

        LOGGER.info("Bot Started Successfully!")

    async def stop(self):
        """Send a message to MESSAGE_DUMP telling that the bot has stopped."""
        LOGGER.info("Uploading logs before stopping...!")
        with open(logfile) as f:
            txt = f.read()
            raw = (await paste(txt))[1]
        # Send Logs to MESSAGE-DUMP
        await self.send_document(
            MESSAGE_DUMP,
            document=logfile,
            caption=f"Logs for last run.\n<code>{log_datetime}</code>",
            reply_markup=InlineKeyboardMarkup(
                [[InlineKeyboardButton("NekoBin Raw", url=raw)]],
            ),
        )
        await self.send_message(
            MESSAGE_DUMP,
            "<i><b>Bot Stopped!</b></i>",
        )
        await super().stop()
        # Flush Redis data
        try:
            await flushredis()
            LOGGER.info("Flushed Redis!")
        except Exception as ef:
            LOGGER.error(ef)
        LOGGER.info("Bot Stopped.\nkthxbye!")
