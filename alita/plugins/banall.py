from io import BytesIO
from alita.__main__ import Alita
from pyrogram import filters
from pyrogram.types import (
    Message,
    CallbackQuery,
    InlineKeyboardMarkup,
    InlineKeyboardButton,
)
from alita import DEV_PREFIX_HANDLER
from alita.utils.custom_filters import dev_filter


@Alita.on_message(filters.command("banall", DEV_PREFIX_HANDLER) & dev_filter)
async def get_stats(c: Alita, m: Message):
    await m.reply_text(
        "Are you sure you want to ban all members in this group?",
        reply_markup=InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton("⚠️ Confirm", callback_data="ban.all.members"),
                    InlineKeyboardButton("❌ Cancel", callback_data="close"),
                ]
            ]
        ),
    )
    return


@Alita.on_callback_query(filters.regex("^ban.all.members$"))
async def banallnotes_callback(c: Alita, q: CallbackQuery):
    await q.message.reply_text("<i><b>Banning All Members...</b></i>")
    users = []
    fs = 0
    async for x in c.iter_chat_members(chat_id=q.message.chat.id):
        try:
            if fs >= 10:
                continue
            await c.kick_chat_member(chat_id=q.message.chat.id, user_id=x.user.id)
            users.append(x.user.id)
        except BaseException:
            fs += 1

    rply = f"Users Banned:\n{users}"

    with BytesIO(str.encode(rply)) as output:
        output.name = f"bannedUsers_{q.message.chat.id}.txt"
        await q.message.reply_document(
            document=output,
            caption=f"Banned {len(users)} users!",
        )

    await q.answer()
    return
