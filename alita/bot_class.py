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


from os import makedirs, path
from time import time

from pyrogram import Client, __version__, errors
from pyrogram.raw.all import layer
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup

from alita import (
    API_HASH,
    APP_ID,
    BOT_USERNAME,
    LOG_DATETIME,
    LOGFILE,
    LOGGER,
    MESSAGE_DUMP,
    NO_LOAD,
    SUPPORT_STAFF,
    TOKEN,
    WORKERS,
    get_self,
    load_cmds,
    setup_redis,
)
from alita.db import users_db as userdb
from alita.plugins import ALL_PLUGINS
from alita.utils.localization import load_langdict
from alita.utils.paste import paste
from alita.utils.redishelper import allkeys, flushredis, set_key

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
        # try:
        #     await flushredis()
        # except Exception as ef:
        #     LOGGER.error(ef)

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
                    adminlist.append(
                        (
                            j.user.id,
                            f"@{j.user.username}"
                            if j.user.username
                            else j.user.first_name,
                        ),
                    )

                ADMINDICT[str(i.chat_id)] = adminlist  # Remove the last space

                LOGGER.info(
                    f"Set {len(adminlist)} admins for {i.chat_id}\n- {adminlist}",
                )
            except errors.PeerIdInvalid:
                pass
            except Exception as ef:
                LOGGER.error(ef)
                pass

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
            LOGGER.error(f"Could not set ADMINDICT in RedisCache!\n{ef}")

    async def start(self):
        await super().start()

        meh = await get_self(self)  # Get bot info from pyrogram client
        LOGGER.info("Starting bot...")

        await self.send_message(MESSAGE_DUMP, "<i>Starting Bot...</i>")

        # Redis Content Setup!
        redis_client = await setup_redis()
        if redis_client:
            LOGGER.info(f"Connected to redis!")
            await self.get_admins()  # Load admins in cache
            await set_key("BOT_ID", meh.id)
            await set_key("BOT_USERNAME", meh.username)
            await set_key("BOT_NAME", meh.first_name)
            await set_key("SUPPORT_STAFF", SUPPORT_STAFF)  # Load SUPPORT_STAFF in cache
        else:
            LOGGER.error(f"Redis not connected!")
        # Redis Content Setup!

        # Load Languages
        lang_status = await load_langdict()
        LOGGER.info(f"Loading Languages: {lang_status}")

        # Show in Log that bot has started
        LOGGER.info(
            f"Pyrogram v{__version__}\n(Layer - {layer}) started on @{BOT_USERNAME}",
        )
        cmd_list = await load_cmds(await ALL_PLUGINS())
        LOGGER.info(f"Plugins Loaded: {cmd_list}")
        # redis_keys = await allkeys()
        # LOGGER.info(f"Redis Keys Loaded: {redis_keys}")

        # Send a message to MESSAGE_DUMP telling that the
        # bot has started and has loaded all plugins!
        await self.send_message(
            MESSAGE_DUMP,
            (
                f"<b><i>{meh.username} started on Pyrogram v{__version__} (Layer - {layer})</i></b>\n\n"
                "<b>Loaded Plugins:</b>\n"
                f"<i>{cmd_list}</i>\n"
                # "<b>Redis Keys Loaded:</b>\n"
                # f"<i>{redis_keys}</i>"
            ),
        )

        LOGGER.info("Bot Started Successfully!")

    async def stop(self):
        """Send a message to MESSAGE_DUMP telling that the bot has stopped."""
        LOGGER.info("Uploading logs before stopping...!")
        with open(LOGFILE) as f:
            txt = f.read()
            raw = (await paste(txt))[1]
        # Send Logs to MESSAGE-DUMP
        await self.send_document(
            MESSAGE_DUMP,
            document=LOGFILE,
            caption=f"Logs for last run, pasted to NekoBin.\n<code>{LOG_DATETIME}</code>",
            reply_markup=InlineKeyboardMarkup(
                [[InlineKeyboardButton("Logs", url=raw)]],
            ),
        )
        await self.send_message(
            MESSAGE_DUMP,
            "<i><b>Bot Stopped!</b></i>",
        )
        await super().stop()
        # Flush Redis data again
        # try:
        #     await flushredis()
        #     LOGGER.info("Flushed Redis!")
        # except Exception as ef:
        #     LOGGER.error(ef)
        LOGGER.info("Bot Stopped.\nkthxbye!")
