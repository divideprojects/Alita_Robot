from traceback import format_exc

from pyrogram.types import CallbackQuery, Message

from alita import DEV_USERS, LOGGER, OWNER_ID, SUDO_USERS

SUDO_LEVEL = SUDO_USERS + DEV_USERS + [int(OWNER_ID)]
DEV_LEVEL = DEV_USERS + [int(OWNER_ID)]


async def admin_check(m: Message or CallbackQuery) -> bool:
    """Checks if user is admin or not."""
    if isinstance(m, Message):
        user_id = m.from_user.id
    if isinstance(m, CallbackQuery):
        user_id = m.message.from_user.id

    try:
        if user_id in SUDO_LEVEL:
            return True
    except Exception as ef:
        LOGGER.error(format_exc())

    user = await m.chat.get_member(user_id)
    admin_strings = ("creator", "administrator")

    if user.status not in admin_strings:
        reply = "Nigga, you're not admin, don't try this explosive shit."
        try:
            await m.edit_text(reply)
        except Exception as ef:
            await m.reply_text(reply)
            LOGGER.error(ef)
            LOGGER.error(format_exc())
        return False

    return True


async def check_rights(m: Message or CallbackQuery, rights) -> bool:
    """Check Admin Rights"""
    if isinstance(m, Message):
        user_id = m.from_user.id
        chat_id = m.chat.id
        app = m._client
    if isinstance(m, CallbackQuery):
        user_id = m.message.from_user.id
        chat_id = m.message.chat.id
        app = m.message._client

    user = await app.get_chat_member(chat_id, user_id)
    if user.status == "member":
        return False
    admin_strings = ("creator", "administrator")
    if user.status in admin_strings:
        return bool(getattr(user, rights, None))
    return False


async def owner_check(m: Message or CallbackQuery) -> bool:
    """Checks if user is owner or not."""
    if isinstance(m, Message):
        user_id = m.from_user.id
    if isinstance(m, CallbackQuery):
        user_id = m.message.from_user.id
        m = m.message

    try:
        if user_id in SUDO_LEVEL:
            return True
    except Exception as ef:
        LOGGER.info(ef, m)
        LOGGER.error(format_exc())

    user = await m.chat.get_member(user_id)

    if user.status != "creator":
        if user.status == "administrator":
            reply = "Stay in your limits, or lose adminship too."
        else:
            reply = "You ain't even admin, what are you trying to do?"
        try:
            await m.edit_text(reply)
        except Exception as ef:
            await m.reply_text(reply)
            LOGGER.error(ef)
            LOGGER.error(format_exc())

        return False

    return True
