from html import escape

from pyrogram import filters
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import HELP_COMMANDS, LOGGER
from alita.bot_class import Alita
from alita.database.disable_db import Disabling
from alita.utils.custom_filters import (
    admin_filter,
    can_change_filter,
    command,
    owner_filter,
)


@Alita.on_message(command("disable") & can_change_filter)
async def disableit(_, m: Message):
    if len(m.text.split()) < 2:
        return await m.reply_text("What to disable?")
    disable_cmd_keys = sorted(
        k
        for j in [HELP_COMMANDS[i]["disablable"] for i in list(HELP_COMMANDS.keys())]
        for k in j
    )

    db = Disabling(m.chat.id)
    disable_list = db.get_disabled()
    LOGGER.info(f"{m.from_user.id} used disabled cmd in {m.chat.id}")

    if str(m.text.split(None, 1)[1]) in disable_list:
        return await m.reply_text("It's already disabled!")

    if str((m.text.split(None, 1)[1]).lower()) in disable_cmd_keys:
        db.add_disable((str(m.text.split(None, 1)[1])).lower())
        return await m.reply_text(f"Disabled {m.text.split(None, 1)[1]}!")
    if str(m.text.split(None, 1)[1]) not in disable_cmd_keys:
        return await m.reply_text("Can't do it sorry !")


@Alita.on_message(command("disabledel") & can_change_filter)
async def set_dsbl_action(_, m: Message):
    db = Disabling(m.chat.id)

    status = db.get_action()
    if status == "none":
        cur = False
    else:
        cur = True
    args = m.text.split(" ", 1)

    LOGGER.info(f"{m.from_user.id} disabledel used in {m.chat.id}")

    if len(args) >= 2:
        if args[1].lower() == "on":
            db.set_action("del")
            await m.reply_text("Oke done!")
        elif args[1].lower() == "off":
            db.set_action("none")
            await m.reply_text("Oke i will not delete!")
        else:
            await m.reply_text("what are you trying to do ??")
    else:
        await m.reply_text(f"Current settings:- {cur}")
    return


@Alita.on_message(command("enable") & can_change_filter)
async def enableit(_, m: Message):
    if len(m.text.split()) < 2:
        return await m.reply_text("What to enable?")
    db = Disabling(m.chat.id)
    disable_list = db.get_disabled()
    if str(m.text.split(None, 1)[1]) not in disable_list:
        return await m.reply_text("It's not disabled!")
    db.remove_disabled((str(m.text.split(None, 1)[1])).lower())
    LOGGER.info(f"{m.from_user.id} enabled something in {m.chat.id}")
    return await m.reply_text(f"Enabled {m.text.split(None, 1)[1]}!")


@Alita.on_message(command("disableable") & can_change_filter)
async def disabling(_, m: Message):
    disable_cmd_keys = sorted(
        k
        for j in [HELP_COMMANDS[i]["disablable"] for i in list(HELP_COMMANDS.keys())]
        for k in j
    )
    tes = "List of commnds that can be disabled:\n"
    tes += "\n".join(f" • <code>{escape(i)}</code>" for i in disable_cmd_keys)
    LOGGER.info(f"{m.from_user.id} checked disableable {m.chat.id}")
    return await m.reply_text(tes)


@Alita.on_message(command("disabled") & admin_filter)
async def disabled(_, m: Message):
    db = Disabling(m.chat.id)
    disable_list = db.get_disabled()
    if not disable_list:
        await m.reply_text("No disabled items!")
        return
    tex = "Disabled commands:\n"
    tex += "\n".join(f" • <code>{escape(i)}</code>" for i in disable_list)
    LOGGER.info(f"{m.from_user.id} checked disabled {m.chat.id}")
    return await m.reply_text(tex)


@Alita.on_message(command("enableall") & owner_filter)
async def rm_alldisbl(_, m: Message):
    db = Disabling(m.chat.id)
    all_dsbl = db.get_disabled()
    if not all_dsbl:
        await m.reply_text("No disabled commands in this chat")
        return
    await m.reply_text(
        "Are you sure you want to enable all?",
        reply_markup=InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        "Confirm",
                        callback_data="enableallcmds",
                    ),
                    InlineKeyboardButton("Cancel", callback_data="close_admin"),
                ],
            ],
        ),
    )
    return


@Alita.on_callback_query(filters.regex("^enableallcmds$"))
async def enablealll(_, q: CallbackQuery):
    user_id = q.from_user.id
    user_status = (await q.message.chat.get_member(user_id)).status
    if user_status not in {"creator", "administrator"}:
        await q.answer(
            "You're not even an admin, don't try this explosive shit!",
            show_alert=True,
        )
        return
    if user_status != "creator":
        await q.answer(
            "You're just an admin, not owner\nStay in your limits!",
            show_alert=True,
        )
        return
    db = Disabling(q.message.chat.id)
    db.rm_all_disabled()
    LOGGER.info(f"{user_id} enabled all in {q.message.chat.id}")
    await q.message.edit_text("Enabled all!", show_alert=True)
    return
