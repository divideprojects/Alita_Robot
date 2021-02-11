from alita.__main__ import Alita
from pyrogram import filters
from pyrogram.types import Message

__PLUGIN__ = "Botstaff"


@Alita.on_message(filters.command("test", DEV_PREFIX_HANDLER))
async def test_bot(m: Message):
    await m.reply_text("Test successful")
    return
