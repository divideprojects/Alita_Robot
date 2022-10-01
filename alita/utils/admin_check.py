from pyrogram.enums import ChatMemberStatus
from pyrogram.types import CallbackQuery, Message

from alita import DEV_USERS, OWNER_ID, SUDO_USERS

SUDO_LEVEL = SUDO_USERS + DEV_USERS + [OWNER_ID]
DEV_LEVEL = DEV_USERS + [OWNER_ID]
admin_strings = (ChatMemberStatus.OWNER, ChatMemberStatus.ADMINISTRATOR)


async def admin_check(m: Message or CallbackQuery) -> bool:
    """Checks if user is admin or not."""
    user_id = 0
    if isinstance(m, Message):
        user_id = m.from_user.id
    if isinstance(m, CallbackQuery):
        user_id = m.message.from_user.id

    if user_id in SUDO_LEVEL:
        return True
    user = await m.chat.get_member(user_id)

    if user.status not in admin_strings:
        reply = "Nigga, you're not admin, don't try this explosive shit."
        await m.edit_text(reply)
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
    if user.status == ChatMemberStatus.MEMBER:
        return False
    if user.status in admin_strings:
        return bool(getattr(user, rights, None))
    return False


async def owner_check(m: Message or CallbackQuery) -> bool:
    """Checks if user is owner or not."""
    user_id = 0
    if isinstance(m, Message):
        user_id = m.from_user.id
    if isinstance(m, CallbackQuery):
        user_id = m.message.from_user.id
        m = m.message

    if user_id in SUDO_LEVEL:
        return True

    user = await m.chat.get_member(user_id)

    if user.status != ChatMemberStatus.OWNER:
        if user.status == ChatMemberStatus.ADMINISTRATOR:
            reply = "Stay in your limits or lose admin ship too."
        else:
            reply = "You ain't even admin, what are you trying to do?"
        await m.edit_text(reply)
        return False

    return True
