from pyrogram.types import Message


async def extract_user(m: Message) -> (int, str):
    user_id = None
    user_first_name = None

    if m.reply_to_message:
        user_id = m.reply_to_message.from_user.id
        user_first_name = m.reply_to_message.from_user.first_name

    elif len(m.command) > 1:
        if len(m.entities) > 1:
            required_entity = m.entities[1]
            if required_entity.type == "text_mention":
                user_id = required_entity.user.id
                user_first_name = required_entity.user.first_name
            elif required_entity.type == "mention":
                user_id = m.text[
                    required_entity.offset : required_entity.offset
                    + required_entity.length
                ]
                user_first_name = user_id
        else:
            user_id = m.command[1]
            user_first_name = user_id

    else:
        user_id = m.from_user.id
        user_first_name = m.from_user.first_name

    return user_id, user_first_name
