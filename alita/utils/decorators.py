from functools import wraps

from pyrogram.types import Message

from alita.bot_class import Alita


def user_admin(func):
    @wraps(func)
    async def decorator(bot: Alita, m: Message):
        if bot.is_admin(m):
            await func(bot, m)

        await m.delete()

    decorator.admin = True

    return decorator


def bot_admin(func):
    @wraps(func)
    async def decorator(bot: Alita, m: Message):
        meh = (await bot.get_me()).id

        if (await m.chat.get_member(meh)).status == "administrator":
            await func(bot, m)

        await m.delete()

    decorator.admin = True

    return decorator
