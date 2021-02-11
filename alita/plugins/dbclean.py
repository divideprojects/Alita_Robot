from time import sleep
from alita.__main__ import Alita
from pyrogram import filters
from pyrogram.types import (
    Message,
    InlineKeyboardMarkup,
    InlineKeyboardButton,
    CallbackQuery,
)
from pyrogram.errors import BadRequest, Unauthorized
from alita import DEV_PREFIX_HANDLER, LOGGER
from alita.db import users_db as user_db, antispam_db as gban_db
from alita.utils.custom_filters import dev_filter

__PLUGIN__ = "Database Cleaning"


async def get_invalid_chats(c: Alita, m: Message, remove: bool = False):
    chats = user_db.get_all_chats()
    kicked_chats, progress = 0, 0
    chat_list = []
    progress_message = m

    for chat in chats:
        if ((100 * chats.index(chat)) / len(chats)) > progress:
            progress_bar = f"{progress}% completed in getting invalid chats."
            if progress_message:
                try:
                    await m.edit_text(progress_bar)
                except BaseException:
                    pass
            else:
                progress_message = await m.reply_text(progress_bar)
            progress += 5

        cid = chat.chat_id
        sleep(0.1)
        try:
            await c.get_chat(cid)
        except (BadRequest, Unauthorized):
            kicked_chats += 1
            chat_list.append(cid)
        except BaseException:
            pass

    try:
        await progress_message.delete()
    except BaseException:
        pass

    if not remove:
        return kicked_chats
    for muted_chat in chat_list:
        sleep(0.1)
        user_db.rem_chat(muted_chat)
    return kicked_chats


async def get_invalid_gban(c: Alita, m: Message, remove: bool = False):
    banned = gban_db.get_gban_list()
    ungbanned_users = 0
    ungban_list = []

    for user in banned:
        user_id = user["user_id"]
        sleep(0.1)
        try:
            await c.get_users(user_id)
        except BadRequest:
            ungbanned_users += 1
            ungban_list.append(user_id)
        except BaseException:
            pass

    if remove:
        for user_id in ungban_list:
            sleep(0.1)
            gban_db.ungban_user(user_id)
        return ungbanned_users
    return ungbanned_users


async def get_muted_chats(c: Alita, m: Message, leave: bool = False):
    chat_id = m.chat.id
    chats = user_db.get_all_chats()
    muted_chats, progress = 0, 0
    chat_list = []
    progress_message = m

    for chat in chats:

        if ((100 * chats.index(chat)) / len(chats)) > progress:
            progress_bar = f"{progress}% completed in getting muted chats."
            if progress_message:
                try:
                    await m.edit_text(
                        progress_bar, chat_id
                    )
                except BaseException:
                    pass
            else:
                progress_message = await m.edit_text(progress_bar)
            progress += 5

        cid = chat.chat_id
        sleep(0.1)

        try:
            await c.send_chat_action(cid, "typing")
        except (BadRequest, Unauthorized):
            muted_chats += 1
            chat_list.append(cid)
        except BaseException:
            pass

    try:
        await progress_message.delete()
    except BaseException:
        pass

    if not leave:
        return muted_chats
    for muted_chat in chat_list:
        sleep(0.1)
        try:
            await c.leave_chat(muted_chat)
        except BaseException:
            pass
        user_db.rem_chat(muted_chat)
    return muted_chats


@Alita.on_message(filters.command("dbclean", DEV_PREFIX_HANDLER) & dev_filter)
async def dbcleanxyz(m: Message):
    buttons = [
        [InlineKeyboardButton("Invalid Chats", callback_data="dbclean_invalidchats")]
    ]
    buttons += [
        [InlineKeyboardButton("Muted Chats", callback_data="dbclean_mutedchats")]
    ]
    buttons += [[InlineKeyboardButton("Invalid Gbans", callback_data="dbclean_gbans")]]
    await m.reply_text(
        "What do you want to clean?", reply_markup=InlineKeyboardMarkup(buttons)
    )
    return


@Alita.on_callback_query(filters.regex("^dbclean_"))
async def dbclean_callback(c: Alita, q: CallbackQuery):
    # Invalid Chats
    if q.data.split("_")[1] == "invalidchats":
        await q.message.edit_text("Getting Invalid Chat Count ...")
        invalid_chat_count = await get_invalid_chats(c, q.message)

        if not invalid_chat_count > 0:
            await q.message.edit_text("No Invalid Chats.")
            return

        await q.message.reply_text(
            f"Total invalid chats - {invalid_chat_count}",
            reply_markup=InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(
                            "Remove Invalid Chats",
                            callback_data="db_clean_inavlid_chats",
                        )
                    ]
                ]
            ),
        )
        await q.message.delete()
        return

    # Muted Chats
    if q.data.split("_")[1] == "mutedchats":
        await q.message.edit_text("Getting Muted Chat Count...")
        muted_chat_count = await get_muted_chats(c, q.message)

        if not muted_chat_count > 0:
            await q.message.delete()
            await q.message.edit_text("I'm not muted in any Chats.")
            return

        await q.message.reply_text(
            f"Muted Chats - {muted_chat_count}",
            reply_markup=InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(
                            "Leave Muted Chats",
                            callback_data="db_clean_muted_chats",
                        )
                    ]
                ]
            ),
        )
        await q.message.delete()
        return

    # Invalid Gbans
    if q.data.split("_")[1] == "gbans":
        await q.message.edit_text("Getting Invalid Gban Count ...")
        invalid_gban_count = await get_invalid_gban(c, q.message)

        if not invalid_gban_count > 0:
            await q.message.edit_text("No Invalid Gbans")
            return

        await q.message.reply_text(
            f"Invalid Gbans - {invalid_gban_count}",
            reply_markup=InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(
                            "Remove Invalid Gbans",
                            callback_data="db_clean_invalid_gbans",
                        )
                    ]
                ]
            ),
        )
        await q.message.delete()
        return
    return


@Alita.on_callback_query(filters.regex("^db_clean_"))
async def db_clean_callbackAction(c: Alita, q: CallbackQuery):
    try:
        query_type = q.data.split("_")[1]

        if query_type == "muted_chats":
            await q.message.edit_text("Leaving chats ...")
            chat_count = await get_muted_chats(c, q.message, True)
            await q.message.edit_text(f"Left {chat_count} chats.")

        elif query_type == "inavlid_chats":
            await q.message.edit_text("Cleaning up Db...")
            invalid_chat_count = await get_invalid_chats(c, q.message, True)
            await q.message.edit_text(f"Cleaned up {invalid_chat_count} chats from Db.")

        elif query_type == "invalid_gbans":
            await q.message.edit_text("Removing Invalid Gbans from Db...")
            invalid_gban_count = await get_invalid_gban(c, q.message, True)
            await q.message.edit_text(
                f"Cleaned up {invalid_gban_count} gbanned users from Db"
            )
    except Exception as ef:
        LOGGER.error(f"Error while cleaning db:\n{ef}")
    return
