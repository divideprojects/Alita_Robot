from typing import Any

from pyrogram.enums import MessageEntityType
from pyrogram.errors import RPCError
from pyrogram.types import Message

from alita import LOGGER
from alita.bot_class import Alita
from alita.database.users_db import Users


async def extract_user(
    c: Alita,
    m: Message,
) -> tuple[Any | None, Any | None, Any | None] | tuple[
    int | str,
    Any | None,
    Any | None,
] | tuple[None, None, None] | tuple[
    int | None | str | Any,
    int | None | str | Any,
    str | None | Any,
]:
    """Extract the user from the provided message."""
    user_id = None
    user_first_name = None
    user_name = None

    if m.reply_to_message and m.reply_to_message.from_user:
        user_id = m.reply_to_message.from_user.id
        user_first_name = m.reply_to_message.from_user.first_name
        user_name = m.reply_to_message.from_user.username

    elif m.reply_to_message and m.reply_to_message.sender_chat:
        user_id = m.reply_to_message.sender_chat.id
        user_first_name = m.reply_to_message.sender_chat.title
        user_name = m.reply_to_message.sender_chat.username

    elif len(m.text.split()) > 1:
        if len(m.entities) > 1:
            required_entity = m.entities[1]
            if required_entity.type == MessageEntityType.TEXT_MENTION:
                user_id = required_entity.user.id
                user_first_name = required_entity.user.first_name
                user_name = required_entity.user.username
            elif required_entity.type in (
                MessageEntityType.MENTION,
                MessageEntityType.PHONE_NUMBER,
            ):
                # new long user ids are identified as phone_number
                user_found = m.text[
                    required_entity.offset : (
                        required_entity.offset + required_entity.length
                    )
                ]
                try:
                    user_found = int(user_found)
                except (AttributeError, ValueError, TypeError) as ef:
                    if "invalid literal for int() with base 10:" in str(ef):
                        user_found = str(user_found)
                    else:
                        LOGGER.error(ef)
                try:
                    user = Users.get_user_info(user_found)
                    user_id = user["_id"]
                    user_first_name = user["name"]
                    user_name = user["username"]
                except KeyError:
                    # If user not in database
                    try:
                        user = await c.get_users(user_found)
                    except (IndexError, RPCError) as ef:
                        await m.reply_text(f"User not found ! Error: {ef}")
                        return user_id, user_first_name, user_name
                    user_id = user.id
                    user_first_name = user.first_name
                    user_name = user.username
                except (AttributeError, ValueError, TypeError) as ef:
                    user_id = user_found
                    user_first_name = user_found
                    user_name = ""
                    LOGGER.error(ef)
        else:
            try:
                user_id = int(m.text.split()[1])
            except (AttributeError, ValueError, TypeError) as ef:
                if "invalid literal for int() with base 10:" in str(ef):
                    user_id = (
                        str(m.text.split()[1])
                        if (m.text.split()[1]).startswith("@")
                        else None
                    )
                else:
                    user_id = m.text.split()[1]
                    LOGGER.error(ef)
            if user_id:
                try:
                    user = Users.get_user_info(user_id)
                    user_first_name = user["name"]
                    user_name = user["username"]
                except KeyError:
                    try:
                        user = await c.get_users(user_id)
                    except (IndexError, RPCError) as ef:
                        await m.reply_text(f"User not found ! Error: {ef}")
                        return user_id, user_first_name, user_name
                    user_first_name = user.first_name
                    user_name = user.username
    else:
        await m.reply_text("User not found!")
        return user_id, user_first_name, user_name

    return user_id, user_first_name, user_name
