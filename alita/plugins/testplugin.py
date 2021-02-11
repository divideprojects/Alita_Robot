from time import time
from alita.__main__ import Alita
from pyrogram import filters
from pyrogram.types import Message

__PLUGIN__ = "Test Plugin"


@Alita.on_message(filters.command("test", "/"))
async def test_bot(m: Message):
    start = time()
    replymsg = await m.reply_text("Calculating...")
    end = round(time() - start, 2)
    await replymsg.edit_text(f"Test complete\nTime Taken:{end} seconds")
    return