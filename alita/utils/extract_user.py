from alita.__main__ import Alita
from pyrogram.types import Message


async def extract_user(c: Alita, m: Message):
    user_id = None
    user_first_name = None

    if m.reply_to_message:
        user_id = m.reply_to_message.from_user.id
        user_first_name = m.reply_to_message.from_user.first_name

    elif not m.reply_to_message:
        if len(m.text.split()) >= 2:
            user = await c.get_users(m.command[1])
            user_id = user.id
            user_first_name = user.first_name

    else:
        user_id = m.from_user.id
        user_first_name = m.from_user.first_name

    return user_id, user_first_name
