from platform import python_version
from threading import RLock
from time import gmtime, strftime, time

from pyrogram import Client, __version__
from pyrogram.raw.all import layer

from alita import (
    API_HASH,
    APP_ID,
    BOT_TOKEN,
    LOG_DATETIME,
    LOGFILE,
    LOGGER,
    MESSAGE_DUMP,
    NO_LOAD,
    UPTIME,
    WORKERS,
    load_cmds,
)
from alita.database import MongoDB
from alita.plugins import all_plugins
from alita.tr_engine import lang_dict
from alita.vars import Config

INITIAL_LOCK = RLock()

# Check if MESSAGE_DUMP is correct
if MESSAGE_DUMP == -100 or not str(MESSAGE_DUMP).startswith("-100"):
    raise Exception(
        "Please enter a vaild Supergroup ID, A Supergroup ID starts with -100",
    )


class Alita(Client):
    """Starts the Pyrogram Client on the Bot Token when we do 'python3 -m alita'"""

    def __init__(self):
        name = self.__class__.__name__.lower()

        super().__init__(
            "Alita_Robot",
            bot_token=BOT_TOKEN,
            plugins=dict(root=f"{name}.plugins", exclude=NO_LOAD),
            api_id=APP_ID,
            api_hash=API_HASH,
            workers=WORKERS,
        )

    async def start(self):
        """Start the bot."""
        await super().start()

        meh = await self.get_me()  # Get bot info from pyrogram client
        LOGGER.info("Starting bot...")
        Config.BOT_ID = meh.id
        Config.BOT_NAME = meh.first_name
        Config.BOT_USERNAME = meh.username

        startmsg = await self.send_message(MESSAGE_DUMP, "<i>Starting Bot...</i>")

        # Load Languages
        lang_status = len(lang_dict) >= 1
        LOGGER.info(f"Loading Languages: {lang_status}\n")

        # Show in Log that bot has started
        LOGGER.info(
            f"Pyrogram v{__version__} (Layer - {layer}) started on {meh.username}",
        )
        LOGGER.info(f"Python Version: {python_version()}\n")

        # Get cmds and keys
        cmd_list = await load_cmds(await all_plugins())

        LOGGER.info(f"Plugins Loaded: {cmd_list}")

        # Send a message to MESSAGE_DUMP telling that the
        # bot has started and has loaded all plugins!
        await startmsg.edit_text(
            (
                f"<b><i>@{meh.username} started on Pyrogram v{__version__} (Layer - {layer})</i></b>\n"
                f"\n<b>Python:</b> <u>{python_version()}</u>\n"
                "\n<b>Loaded Plugins:</b>\n"
                f"<i>{cmd_list}</i>\n"
            ),
        )

        LOGGER.info("Bot Started Successfully!\n")

    async def stop(self):
        """Stop the bot and send a message to MESSAGE_DUMP telling that the bot has stopped."""
        runtime = strftime("%Hh %Mm %Ss", gmtime(time() - UPTIME))
        LOGGER.info("Uploading logs before stopping...!\n")
        # Send Logs to MESSAGE_DUMP and LOG_CHANNEL
        await self.send_document(
            MESSAGE_DUMP,
            document=LOGFILE,
            caption=(
                "Bot Stopped!\n\n" f"Uptime: {runtime}\n" f"<code>{LOG_DATETIME}</code>"
            ),
        )
        if MESSAGE_DUMP:
            # LOG_CHANNEL is not necessary
            await self.send_document(
                MESSAGE_DUMP,
                document=LOGFILE,
                caption=f"Uptime: {runtime}",
            )
        await super().stop()
        MongoDB.close()
        LOGGER.info(
            f"""Bot Stopped.
            Logs have been uploaded to the MESSAGE_DUMP Group!
            Runtime: {runtime}s\n
        """,
        )
