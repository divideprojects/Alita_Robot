from alita.__main__ import Alita
from pyrogram.types import Message


async def admin_check(c: Alita, m: Message) -> bool:
    chat_id = m.chat.id
    user_id = m.from_user.id

    user = await c.get_chat_member(chat_id=chat_id, user_id=user_id)
    admin_strings = ["creator", "administrator"]

    if user.status not in admin_strings:
        await m.reply_text(
            "This is an Admin Restricted command and you're not allowed to use it."
        )
        return False

    return True


async def owner_check(c: Alita, m: Message) -> bool:
    chat_id = m.chat.id
    user_id = m.from_user.id

    user = await c.get_chat_member(chat_id=chat_id, user_id=user_id)

    if user.status != "creator":
        await m.reply_text(
            "This is an Owner Restricted command and you're not allowed to use it."
        )
        return False

    return True
