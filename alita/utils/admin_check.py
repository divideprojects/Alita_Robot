from alita import OWNER_ID, DEV_USERS


async def admin_check(c, m) -> bool:
    chat_id = m.chat.id
    user_id = m.from_user.id

    if int(user_id) == int(OWNER_ID) or int(user_id) in DEV_USERS:
        return True

    user = await c.get_chat_member(chat_id=chat_id, user_id=user_id)
    admin_strings = ["creator", "administrator"]

    if user.status not in admin_strings:
        await m.reply_text(
            "This is an Admin Restricted command and you're not allowed to use it."
        )
        return False

    return True


async def owner_check(c, m) -> bool:
    chat_id = m.chat.id
    user_id = m.from_user.id

    if int(user_id) == int(OWNER_ID) or int(user_id) in DEV_USERS:
        return True

    user = await c.get_chat_member(chat_id=chat_id, user_id=user_id)

    if user.status != "creator":
        await m.reply_text(
            "This is an Owner Restricted command and you're not allowed to use it."
        )
        return False

    return True
