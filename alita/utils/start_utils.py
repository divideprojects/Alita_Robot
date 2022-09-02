from html import escape
from secrets import choice

from pyrogram.enums import ParseMode
from pyrogram.errors import RPCError
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import HELP_COMMANDS, LOGGER, SUPPORT_GROUP
from alita.bot_class import Alita
from alita.database.chats_db import Chats
from alita.database.notes_db import Notes
from alita.database.rules_db import Rules
from alita.tr_engine import tlang
from alita.utils.cmd_senders import send_cmd
from alita.utils.msg_types import Types
from alita.utils.string import (
    build_keyboard,
    escape_mentions_using_curly_brackets,
    parse_button,
)


async def gen_cmds_kb(m: Message or CallbackQuery):
    """Generate the keyboard for languages."""
    if isinstance(m, CallbackQuery):
        m = m.message

    cmds = sorted(list(HELP_COMMANDS.keys()))
    kb = [
        InlineKeyboardButton(tlang(m, cmd), callback_data=f"get_mod.{cmd.lower()}")
        for cmd in cmds
    ]

    return [kb[i: i + 3] for i in range(0, len(kb), 3)]


async def gen_start_kb(q: Message or CallbackQuery):
    """Generate keyboard with start menu options."""
    from alita.vars import Config

    return InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton(
                    f"➕ {(tlang(q, 'start.add_chat_btn'))}",
                    url=f"https://t.me/{Config.BOT_USERNAME}?startgroup=new",
                ),
                InlineKeyboardButton(
                    f"{(tlang(q, 'start.support_group'))} 👥",
                    url=f"https://t.me/{SUPPORT_GROUP}",
                ),
            ],
            [
                InlineKeyboardButton(
                    f"📚 {(tlang(q, 'start.commands_btn'))}",
                    callback_data="commands",
                ),
            ],
        ],
    )


async def get_private_note(c: Alita, m: Message, help_option: str):
    """Get the note in pm of user, with parsing enabled."""
    from alita.vars import Config
    # Initialize
    notes_db = Notes()

    help_lst = help_option.split("_")
    if len(help_lst) == 2:
        chat_id = int(help_lst[1])

        all_notes = notes_db.get_all_notes(chat_id)
        chat_title = Chats.get_chat_info(chat_id)["chat_name"]
        rply = f"Notes in {chat_title}:\n"
        note_list = [
            f"- [{note[0]}](https://t.me/{Config.BOT_USERNAME}?start=note_{chat_id}_{note[1]}_0)"
            for note in all_notes
        ]
        rply += "\n".join(note_list)
        rply += "\n\nYou can retrieve these notes by tapping on the note-name."
        await m.reply_text(rply, disable_web_page_preview=True, quote=True)
        return

    if len(help_lst) < 3:
        return

    note_hash = help_option.split("_")[2]
    raw = bool(int(help_option.split("_")[3]))
    getnotes = notes_db.get_note_by_hash(note_hash)
    if not getnotes:
        await m.reply_text("Note does not exist!", quote=True)
        return

    msgtype = getnotes["msgtype"]
    if not msgtype:
        await m.reply_text(
            "<b>Error:</b> Cannot find a type for this note!!",
            quote=True,
        )
        return

    try:
        # support for random notes texts
        splitter = "%%%"
        note_replu = getnotes["note_value"].split(splitter)
        note_replu = choice(note_replu)
    except KeyError:
        note_replu = ""

    if raw:
        if msgtype == Types.TEXT:
            await m.reply_text(note_replu, parse_mode=ParseMode.DISABLED)
        elif msgtype in (
            Types.STICKER,
            Types.VIDEO_NOTE,
            Types.CONTACT,
            Types.ANIMATED_STICKER,
        ):
            await (await send_cmd(c, msgtype))(
                m.chat.id,
                getnotes["fileid"],
            )
        else:
            await (await send_cmd(c, msgtype))(
                m.chat.id,
                getnotes["fileid"],
                caption=note_replu,
                parse_mode=ParseMode.DISABLED,
            )
            LOGGER.info(
                f"{m.from_user.id} fetched raw note {note_hash} (type - {getnotes}) in {int(help_lst[1])}",
            )
        return

    parse_words = (
        "first",
        "last",
        "fullname",
        "username",
        "mention",
        "id",
        "chatname",
    )
    texti = await escape_mentions_using_curly_brackets(
        m, note_replu, parse_words
    )
    pre = "{preview}" not in texti
    pro = "{protect}" in texti

    textt, butto = await parse_button(
        texti.replace("{preview}", "")
            .replace("{private}", "")
            .replace("{protect}", "")
    )
    buttn = await build_keyboard(butto)
    button = InlineKeyboardMarkup(buttn) if buttn else None

    try:
        if msgtype == Types.TEXT:
            await (await send_cmd(c, msgtype))(
                m.chat.id,
                textt,
                reply_markup=button,
                disable_web_page_preview=pre,
                protect_content=pro,
            )
        elif msgtype in (
            Types.STICKER,
            Types.VIDEO_NOTE,
            Types.CONTACT,
            Types.ANIMATED_STICKER,
        ):
            await (await send_cmd(c, msgtype))(
                m.chat.id,
                getnotes["fileid"],
                reply_markup=button,
                protect_content=pro,
            )
        else:
            await (await send_cmd(c, msgtype))(
                m.chat.id,
                getnotes["fileid"],
                caption=textt,
                reply_markup=button,
                protect_content=pro,
            )
        LOGGER.info(
            f"{m.from_user.id} fetched privatenote {note_hash} (type - {getnotes}) in {m.chat.id}",
        )
    except RPCError as e:
        await m.reply_text(f"Error: {e.MESSAGE}")
    return


async def get_private_rules(_, m: Message, help_option: str):
    chat_id = int(help_option.split("_")[1])
    rules = Rules(chat_id).get_rules()
    chat_title = Chats.get_chat_info(chat_id)["chat_name"]
    if not rules:
        await m.reply_text(
            "The Admins of that group have not setup any rules, that doesn't mean you break the decorum of the chat!",
            quote=True,
        )
        return ""
    await m.reply_text(
        f"""The rules for <b>{escape(chat_title)} are</b>:\n
{rules}
""",
        quote=True,
        disable_web_page_preview=True,
    )
    return ""


async def get_help_msg(m: Message or CallbackQuery, help_option: str):
    """Helper function for getting help_msg and it's keyboard."""
    help_cmd_keys = sorted(
        k
        for j in [HELP_COMMANDS[i]["alt_cmds"] for i in list(HELP_COMMANDS.keys())]
        for k in j
    )

    if help_option in help_cmd_keys:
        help_option_name = next(
            HELP_COMMANDS[i]
            for i in HELP_COMMANDS
            if help_option in HELP_COMMANDS[i]["alt_cmds"]
        )
        help_option_value = help_option_name["help_msg"]
        help_kb = next(
            HELP_COMMANDS[i]["buttons"]
            for i in HELP_COMMANDS
            if help_option in HELP_COMMANDS[i]["alt_cmds"]
        ) + [
                      [
                          InlineKeyboardButton(
                              "« " + (tlang(m, "general.back_btn")),
                              callback_data="commands",
                          ),
                      ],
                  ]
        help_msg = (
            f"**{(tlang(m, (help_option_name['help_msg']).replace('.help', '.main')))}:**\n\n"
            + tlang(m, help_option_value)
        )
        LOGGER.info(
            f"{m.from_user.id} fetched help for {help_option} in {m.chat.id}",
        )
    else:
        help_msg = tlang(m, "general.commands_available")
        help_kb = [
            *(await gen_cmds_kb(m)),
            [
                InlineKeyboardButton(
                    f"« {(tlang(m, 'general.back_btn'))}",
                    callback_data="start_back",
                ),
            ],
        ]

    return help_msg, help_kb
