from pyrogram import filters
from pyrogram.errors import MessageNotModified, QueryIdInvalid, UserIsBlocked
from pyrogram.types import CallbackQuery, Message

from alita import HELP_COMMANDS, LOGGER
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils.custom_filters import command
from alita.utils.kbhelpers import ikb
from alita.utils.start_utils import (
    gen_cmds_kb,
    gen_start_kb,
    get_help_msg,
    get_private_note,
    get_private_rules,
)
from alita.vars import Config


@Alita.on_message(
    command("donate") & (filters.group | filters.private),
)
async def donate(_, m: Message):
    LOGGER.info(f"{m.from_user.id} fetched donation text in {m.chat.id}")
    await m.reply_text(tlang(m, "general.donate_owner"))
    return


@Alita.on_callback_query(filters.regex("^close_admin$"))
async def close_admin_callback(_, q: CallbackQuery):
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
    await q.message.edit_text("Closed!")
    await q.answer("Closed menu!", show_alert=True)
    return


@Alita.on_message(
    command("start") & (filters.group | filters.private),
)
async def start(c: Alita, m: Message):
    if m.chat.type == "private":
        if len(m.text.split()) > 1:
            help_option = (m.text.split(None, 1)[1]).lower()

            if help_option.startswith("note") and (
                help_option not in ("note", "notes")
            ):
                await get_private_note(c, m, help_option)
                return
            if help_option.startswith("rules"):
                LOGGER.info(f"{m.from_user.id} fetched privaterules in {m.chat.id}")
                await get_private_rules(c, m, help_option)
                return

            help_msg, help_kb = await get_help_msg(m, help_option)

            if not help_msg:
                return

            await m.reply_text(
                help_msg,
                parse_mode="markdown",
                reply_markup=ikb(help_kb),
                quote=True,
                disable_web_page_preview=True,
            )
            return
        try:
            await m.reply_text(
                (tlang(m, "start.private")),
                reply_markup=(await gen_start_kb(m)),
                quote=True,
                disable_web_page_preview=True,
            )
        except UserIsBlocked:
            LOGGER.warning(f"Bot blocked by {m.from_user.id}")
    else:
        await m.reply_text(
            (tlang(m, "start.group")),
            quote=True,
        )
    return


@Alita.on_callback_query(filters.regex("^start_back$"))
async def start_back(_, q: CallbackQuery):
    try:
        await q.message.edit_text(
            (tlang(q, "start.private")),
            reply_markup=(await gen_start_kb(q.message)),
            disable_web_page_preview=True,
        )
    except MessageNotModified:
        pass
    await q.answer()
    return


@Alita.on_callback_query(filters.regex("^commands$"))
async def commands_menu(_, q: CallbackQuery):
    keyboard = ikb(
        [
            *(await gen_cmds_kb(q)),
            [(f"« {(tlang(q, 'general.back_btn'))}", "start_back")],
        ],
    )
    try:
        await q.message.edit_text(
            (tlang(q, "general.commands_available")),
            reply_markup=keyboard,
        )
    except MessageNotModified:
        pass
    except QueryIdInvalid:
        await q.message.reply_text(
            (tlang(q, "general.commands_available")),
            reply_markup=keyboard,
        )
    await q.answer()
    return


@Alita.on_message(command("help"))
async def help_menu(_, m: Message):
    if len(m.text.split()) >= 2:
        help_option = (m.text.split(None, 1)[1]).lower()
        help_msg, help_kb = await get_help_msg(m, help_option)

        if not help_msg:
            LOGGER.error(f"No help_msg found for help_option - {help_option}!!")
            return

        LOGGER.info(
            f"{m.from_user.id} fetched help for '{help_option}' text in {m.chat.id}",
        )
        if m.chat.type == "private":
            await m.reply_text(
                help_msg,
                parse_mode="markdown",
                reply_markup=ikb(help_kb),
                quote=True,
                disable_web_page_preview=True,
            )
        else:
            await m.reply_text(
                (tlang(m, "start.public_help").format(help_option=help_option)),
                reply_markup=ikb(
                    [
                        [
                            (
                                "Help",
                                f"t.me/{Config.BOT_USERNAME}?start={help_option}",
                                "url",
                            ),
                        ],
                    ],
                ),
            )
    else:
        if m.chat.type == "private":
            keyboard = ikb(
                [
                    *(await gen_cmds_kb(m)),
                    [(f"« {(tlang(m, 'general.back_btn'))}", "start_back")],
                ],
            )
            msg = tlang(m, "general.commands_available")
        else:
            keyboard = ikb(
                [[("Help", f"t.me/{Config.BOT_USERNAME}?start=help", "url")]],
            )
            msg = tlang(m, "start.pm_for_help")

        await m.reply_text(
            msg,
            reply_markup=keyboard,
        )

    return


@Alita.on_callback_query(filters.regex("^get_mod."))
async def get_module_info(_, q: CallbackQuery):
    module = q.data.split(".", 1)[1]

    help_msg = f"**{(tlang(q, str(module)))}:**\n\n" + tlang(
        q,
        HELP_COMMANDS[module]["help_msg"],
    )

    help_kb = HELP_COMMANDS[module]["buttons"] + [
        [("« " + (tlang(q, "general.back_btn")), "commands")],
    ]
    await q.message.edit_text(
        help_msg,
        parse_mode="markdown",
        reply_markup=ikb(help_kb),
    )
    await q.answer()
    return
