from pyrogram import filters
from pyrogram.types import CallbackQuery, Message

from alita import LOGGER
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils.custom_filters import command
from alita.utils.kbhelpers import ikb


async def gen_formatting_kb(m):
    return ikb(
        [
            [
                ("Markdown Formatting", "formatting.md_formatting"),
                ("Fillings", "formatting.fillings"),
            ],
            [("Random Content", "formatting.random_content")],
            [(("« " + (tlang(m, "general.back_btn"))), "commands")],
        ],
    )


@Alita.on_message(
    command(["markdownhelp", "formatting"]) & filters.private,
)
async def markdownhelp(_, m: Message):
    await m.reply_text(
        tlang(m, f"plugins.{__PLUGIN__}.help"),
        quote=True,
        reply_markup=(await gen_formatting_kb(m)),
    )
    LOGGER.info(f"{m.from_user.id} used cmd '{m.command}' in {m.chat.id}")
    return


@Alita.on_callback_query(filters.regex("^formatting."))
async def get_formatting_info(_, q: CallbackQuery):
    cmd = q.data.split(".")[1]
    kb = ikb([[((tlang(q, "general.back_btn")), "back.formatting")]])

    if cmd == "md_formatting":
        await q.message.edit_text(
            tlang(q, "formatting.md_help"),
            reply_markup=kb,
            parse_mode="html",
        )
    elif cmd == "fillings":
        await q.message.edit_text(
            tlang(q, "formatting.filling_help"),
            reply_markup=kb,
            parse_mode="html",
        )
    elif cmd == "random_content":
        await q.message.edit_text(
            tlang(q, "formatting.random_help"),
            reply_markup=kb,
            parse_mode="html",
        )

    await q.answer()
    return


@Alita.on_callback_query(filters.regex("^back."))
async def send_mod_help(_, q: CallbackQuery):
    await q.message.edit_text(
        (tlang(q, "plugins.formatting.help")),
        reply_markup=(await gen_formatting_kb(q.message)),
    )
    await q.answer()
    return


__PLUGIN__ = "formatting"

__alt_name__ = ["formatting", "markdownhelp", "markdown"]
__buttons__ = [
    [
        ("Markdown Formatting", "formatting.md_formatting"),
        ("Fillings", "formatting.fillings"),
    ],
    [("Random Content", "formatting.random_content")],
]
