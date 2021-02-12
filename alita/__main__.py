import os
import time
from pyrogram import Client, __version__, errors
from pyrogram.raw.all import layer
from alita.plugins import ALL_PLUGINS
from alita.db import users_db as userdb
from alita import (
    APP_ID,
    API_HASH,
    LOGGER,
    TOKEN,
    NO_LOAD,
    MESSAGE_DUMP,
    HELP_COMMANDS,
    WORKERS,
    load_cmds,
    SUPPORT_STAFF,
    logfile,
    log_datetime,
)
from alita.utils.redishelper import set_key, flushredis, allkeys

# Check that MESSAGE_DUMP ID is correct
if MESSAGE_DUMP == -100 or not str(MESSAGE_DUMP).startswith("-100"):
    raise Exception(
        "Please enter a vaild Supergroup ID, A Supergroup ID starts with -100"
    )


class Alita(Client):
    """Starts the Pyrogram Client on the Bot Token when we do 'python3 -m alita'"""

    def __init__(self):
        name = self.__class__.__name__.lower()

        # Make a temporary direcory for storing session file
        SESSION_DIR = f"{name}/SESSION"
        if not os.path.isdir(SESSION_DIR):
            os.makedirs(SESSION_DIR)

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
        begin = time.time()
        c = self

        # Flush Redis data
        try:
            flushredis()
        except Exception as ef:
            LOGGER.error(ef)

        all_chats = userdb.get_all_chats() or []  # Get list of all chats
        LOGGER.info(f"{len(all_chats)} chats loaded.")
        ADMINDICT = {}
        for chat in all_chats:
            adminlist = []
            try:
                async for i in c.iter_chat_members(
                    chat_id=chat.chat_id, filter="administrators"
                ):
                    adminlist.append(i.user.id)

                ADMINDICT[str(chat.chat_id)] = adminlist  # Remove the last space

                LOGGER.info(
                    f"Set {len(adminlist)} admins for {chat.chat_id}\n{adminlist}"
                )
            except errors.PeerIdInvalid:
                pass
            except Exception as ef:
                LOGGER.error(ef)

        try:
            set_key("ADMINDICT", ADMINDICT)
            end = time.time()
            LOGGER.info(f"Set admin list cache!\nTime Taken: {round(end-begin, 2)}s")
        except Exception as ef:
            LOGGER.error(f"Could not set ADMINDICT!\n{ef}")

    async def start(self):
        await super().start()

        me = await self.get_me()  # Get bot info from pyrogram client
        LOGGER.info("Starting bot...")

        await self.send_message(MESSAGE_DUMP, "Starting Bot...")

        # Redis Content Setup!
        await self.get_admins()  # Load admins in cache
        set_key("SUPPORT_STAFF", SUPPORT_STAFF)  # Load SUPPORT_STAFF in cache
        set_key("BOT_ID", int(me.id))  # Save Bot ID in Redis!
        # Redis Content Setup!

        # Show in Log that bot has started
        LOGGER.info(
            f"Pyrogram v{__version__}\n(Layer - {layer}) started on @{me.username}"
        )
        LOGGER.info(load_cmds(ALL_PLUGINS))
        LOGGER.info(f"Redis Keys Loaded: {allkeys()}")

        # Send a message to MESSAGE_DUMP telling that the bot has started and has loaded all plugins!
        await self.send_message(
            MESSAGE_DUMP,
            (
                f"<b><i>Bot started on Pyrogram v{__version__} (Layer - {layer})</i></b>\n\n"
                "<b>Loaded Plugins:</b>\n"
                f"<i>{list(HELP_COMMANDS.keys())}</i>\n"
                "<b>Redis Keys Loaded:</b>\n"
                f"<i>{allkeys()}</i>"
            ),
        )

        LOGGER.info("Bot Started Successfully!")

    async def stop(self, *args):
        """Send a message to MESSAGE_DUMP telling that the bot has stopped!"""
        LOGGER.info("Uploading logs before stopping...!")
        # Send Logs to MESSAGE-DUMP
        await self.send_document(
            MESSAGE_DUMP,
            document=logfile,
            caption=f"Logs for last run.\n<code>{log_datetime}</code>",
        )
        await self.send_message(
            MESSAGE_DUMP,
            "<i><b>Bot Stopped!</b></i>",
        )
        await super().stop()
        # Flush Redis data
        try:
            flushredis()
            LOGGER.info("Flushed Redis!")
        except Exception as ef:
            LOGGER.error(ef)
        LOGGER.info("Bot Stopped.\nkthxbye!")


if __name__ == "__main__":
    Alita().run()
