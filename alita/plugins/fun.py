from html import escape
from secrets import choice

from pyrogram.errors import MessageTooLong
from pyrogram.types import Message

from alita import LOGGER
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils import fun_strings
from alita.utils.custom_filters import command
from alita.utils.extract_user import extract_user


@Alita.on_message(command("shout"))
async def fun_shout(_, m: Message):
    if len(m.text.split()) == 1:
        await m.reply_text(
            (tlang(m, "general.check_help")),
            quote=True,
        )
        return
    try:
        text = " ".join(m.text.split(None, 1)[1])
        result = [" ".join(list(text))]
        for pos, symbol in enumerate(text[1:]):
            result.append(symbol + " " + "  " * pos + symbol)
        result = list("\n".join(result))
        result[0] = text[0]
        result = "".join(result)
        msg = "```\n" + result + "```"
        await m.reply_text(msg, parse_mode="markdown")
        LOGGER.info(f"{m.from_user.id} shouted in {m.chat.id}")
        return
    except MessageTooLong as e:
        await m.reply_text(f"Error: {e}")
        return


@Alita.on_message(command("runs"))
async def fun_run(_, m: Message):
    await m.reply_text(choice(fun_strings.RUN_STRINGS))
    LOGGER.info(f"{m.from_user.id} runed in {m.chat.id}")
    return


@Alita.on_message(command("slap"))
async def fun_slap(c: Alita, m: Message):
    me = await c.get_me()

    reply_text = m.reply_to_message.reply_text if m.reply_to_message else m.reply_text

    curr_user = escape(m.from_user.first_name)
    try:
        user_id, user_first_name, _ = await extract_user(c, m)
    except Exception:
        return

    if user_id == me.id:
        temp = choice(fun_strings.SLAP_ALITA_TEMPLATES)
    else:
        temp = choice(fun_strings.SLAP_TEMPLATES)

    if user_id:
        user1 = curr_user
        user2 = escape(user_first_name)

    else:
        user1 = me.first_name
        user2 = curr_user

    item = choice(fun_strings.ITEMS)
    hit = choice(fun_strings.HIT)
    throw = choice(fun_strings.THROW)

    reply = temp.format(user1=user1, user2=user2, item=item, hits=hit, throws=throw)
    await reply_text(reply)
    LOGGER.info(f"{m.from_user.id} slaped in {m.chat.id}")
    return


@Alita.on_message(command("roll"))
async def fun_roll(_, m: Message):
    reply_text = m.reply_to_message.reply_text if m.reply_to_message else m.reply_text
    await reply_text(choice(range(1, 7)))
    LOGGER.info(f"{m.from_user.id} roll in {m.chat.id}")
    return


@Alita.on_message(command("toss"))
async def fun_toss(_, m: Message):
    reply_text = m.reply_to_message.reply_text if m.reply_to_message else m.reply_text
    await reply_text(choice(fun_strings.TOSS))
    LOGGER.info(f"{m.from_user.id} tossed in {m.chat.id}")
    return


@Alita.on_message(command("shrug"))
async def fun_shrug(_, m: Message):
    reply_text = m.reply_to_message.reply_text if m.reply_to_message else m.reply_text
    await reply_text(r"¯\_(ツ)_/¯")
    LOGGER.info(f"{m.from_user.id} shruged in {m.chat.id}")
    return


@Alita.on_message(command("bluetext"))
async def fun_bluetext(_, m: Message):
    reply_text = m.reply_to_message.reply_text if m.reply_to_message else m.reply_text
    await reply_text(
        "/BLUE /TEXT\n/MUST /CLICK\n/I /AM /A /STUPID /ANIMAL /THAT /IS /ATTRACTED /TO /COLORS",
    )
    LOGGER.info(f"{m.from_user.id} bluetexted in {m.chat.id}")
    return


@Alita.on_message(command("decide"))
async def fun_decide(_, m: Message):
    reply_text = m.reply_to_message.reply_text if m.reply_to_message else m.reply_text
    await reply_text(choice(fun_strings.DECIDE))
    LOGGER.info(f"{m.from_user.id} decided in {m.chat.id}")
    return


@Alita.on_message(command("react"))
async def fun_table(_, m: Message):
    reply_text = m.reply_to_message.reply_text if m.reply_to_message else m.reply_text
    await reply_text(choice(fun_strings.REACTIONS))
    LOGGER.info(f"{m.from_user.id} reacted in {m.chat.id}")
    return


@Alita.on_message(command("weebify"))
async def weebify(_, m: Message):
    if len(m.text.split()) >= 2:
        args = m.text.split(None, 1)[1]
    elif m.reply_to_message and len(m.text.split()) == 1:
        args = m.reply_to_message.text
    else:
        await m.reply_text(
            "Please reply to a message or enter text after command to weebify it.",
        )
        return
    if not args:
        await m.reply_text(tlang(m, "utils.weebify.weebify_what"))
        return

    # Use split to convert to list
    # Not using list itself becuase black changes it to long format...
    normiefont = "a b c d e f g h i j k l m n o p q r s t u v w x y z".split()
    weebyfont = "卂 乃 匚 刀 乇 下 厶 卄 工 丁 长 乚 从 𠘨 口 尸 㔿 尺 丂 丅 凵 リ 山 乂 丫 乙".split()

    string = "  ".join(args).lower()
    for normiecharacter in string:
        if normiecharacter in normiefont:
            weebycharacter = weebyfont[normiefont.index(normiecharacter)]
            string = string.replace(normiecharacter, weebycharacter)

    await m.reply_text(
        (tlang(m, "utils.weebify.weebified_string").format(string=string)),
    )
    LOGGER.info(f"{m.from_user.id} weebified '{args}' in {m.chat.id}")
    return


__PLUGIN__ = "fun"

_DISABLE_CMDS_ = [
    "weebify",
    "decide",
    "react",
    "bluetext",
    "toss",
    "roll",
    "slap",
    "runs",
    "shout",
    "shrug",
]
