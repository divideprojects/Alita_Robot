# Copyright (C) 2020 - 2021 Divkix. All rights reserved. Source code available under the AGPL.
#
# This file is part of Alita_Robot.
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as
# published by the Free Software Foundation, either version 3 of the
# License, or (at your option) any later version.

# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.

# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.


from datetime import datetime
from html import escape

from googletrans import LANGUAGES, Translator
from pyrogram import errors, filters
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup, Message
from tswift import Song

from alita import (
    DEV_USERS,
    LOGGER,
    OWNER_ID,
    PREFIX_HANDLER,
    SUDO_USERS,
    SUPPORT_GROUP,
    TOKEN,
    WHITELIST_USERS,
)
from alita.bot_class import Alita
from alita.utils.aiohttp_helper import AioHttp
from alita.utils.extract_user import extract_user
from alita.utils.localization import GetLang
from alita.utils.parser import mention_html
from alita.utils.paste import paste

__PLUGIN__ = "Utils"

__help__ = """
Some utils provided by bot to make your tasks easy!

 × /id: Get the current group id. If used by replying to a message, get that user's id.
 × /info: Get information about a user.
 × /ping - Get ping time of bot to telegram server.
 × /gifid: Reply to a gif to me to tell you its file ID.
 × /tr <language>: Translates the text and then replies to you with the language you have specifed, works as a reply to message.
 × /github <username>: Search for the user using github api!
 × /lyrics <song>: Get the lyrics of the song you specify!
 × /weebify <text> or a reply to message: To weebify the message.
"""


@Alita.on_message(
    filters.command("ping", PREFIX_HANDLER) & (filters.group | filters.private),
)
async def ping(_: Alita, m: Message):
    first = datetime.now()
    sent = await m.reply_text("**Ping...**")
    second = datetime.now()
    await sent.edit_text(
        f"**Pong!**\n`{round(((second-first).microseconds / 1000000), 2)}` Secs",
    )
    return


@Alita.on_message(
    filters.command("lyrics", PREFIX_HANDLER) & (filters.group | filters.private),
)
async def get_lyrics(_: Alita, m: Message):
    query = m.text.split(None, 1)[1]
    song = ""
    if not query:
        await m.edit_text("You haven't specified which song to look for!")
        return
    em = await m.reply_text(f"**Finding lyrics for:** `{query}`")
    song = Song.find_song(query)
    if song:
        if song.lyrics:
            reply = song.format()
        else:
            reply = "Couldn't find any lyrics for that song!"
    else:
        reply = "Song not found!"
    if len(reply) > 4090:
        with open("lyrics.txt", "w+") as f:
            f.write(reply)
            await m.reply_document(
                document=f,
                caption="Message length exceeded max limit!\nSent as a text file.",
            )
        await em.delete()
    else:
        await em.edit_text(reply)
    return


@Alita.on_message(
    filters.command("id", PREFIX_HANDLER) & (filters.group | filters.private),
)
async def id_info(c: Alita, m: Message):
    user_id = (await extract_user(m))[0]
    if user_id:
        if m.reply_to_message and m.reply_to_message.forward_from:
            user1 = m.reply_to_m.from_user
            user2 = m.reply_to_m.forward_from
            await m.reply_text(
                (
                    f"Original Sender - {(await mention_html(user2.first_name, user2.id))} "
                    f"(<code>{user2.id}</code>).\n"
                    f"Forwarder - {(await mention_html(user1.first_name, user1.id))} "
                    f"(<code>{user1.id}</code>)."
                ),
                parse_mode="HTML",
            )
        else:
            try:
                user = await c.get_users(user_id)
            except errors.PeerIdInvalid:
                await m.reply_text(
                    "Failed to get user\nPeer ID invalid, I haven't seen this user anywhere earlier, maybe username would help to know them!",
                )

            await m.reply_text(
                f"{(await mention_html(user.first_name, user.id))}'s ID is <code>{user.id}</code>.",
                parse_mode="HTML",
            )
    else:
        if m.chat.type == "private":
            await m.reply_text(
                f"Your ID is <code>{m.chat.id}</code>.",
                parse_mode="HTML",
            )
        else:
            await m.reply_text(
                f"This Group's ID is <code>{m.chat.id}</code>.",
                parse_mode="HTML",
            )
    return


@Alita.on_message(
    filters.command("gifid", PREFIX_HANDLER) & (filters.group | filters.private),
)
async def get_gifid(_: Alita, m: Message):
    if m.reply_to_message and m.reply_to_message.animation:
        await m.reply_text(
            f"Gif ID:\n<code>{m.reply_to_message.animation.file_id}</code>",
            parse_mode="html",
        )
    else:
        await m.reply_text("Please reply to a gif to get its ID.")
    return


@Alita.on_message(
    filters.command("github", PREFIX_HANDLER) & (filters.group | filters.private),
)
async def github(_: Alita, m: Message):
    if len(m.text.split()) == 2:
        username = m.text.split(None, 1)[1]
    else:
        await m.reply_text(
            f"Usage: `{PREFIX_HANDLER}github <username>`",
            parse_mode="md",
        )
        return

    URL = f"https://api.github.com/users/{username}"
    result, resp = await AioHttp.get_json(URL)
    if resp.status == 404:
        await m.reply_text(f"`{username} not found`", parse_mode="md")
        return

    url = result.get("html_url", None)
    name = result.get("name", None)
    company = result.get("company", None)
    bio = result.get("bio", None)
    created_at = result.get("created_at", "Not Found")

    REPLY = (
        f"**GitHub Info for** `{username}`"
        f"\n**Username:** `{name}`\n**Bio:** `{bio}`\n**URL:** {url}"
        f"\n**Company:** `{company}`\n**Created at:** `{created_at}`"
    )

    await m.reply_text(REPLY)

    return


@Alita.on_message(
    filters.command("info", PREFIX_HANDLER) & (filters.group | filters.private),
)
async def my_info(c: Alita, m: Message):
    infoMsg = await m.reply_text("<code>Getting user information...</code>")
    user_id = (await extract_user(m))[0]
    try:
        user = await c.get_users(user_id)
    except errors.PeerIdInvalid:
        await m.reply_text(
            "Failed to get user\nPeer ID invalid, I haven't seen this user anywhere earlier, maybe username would help to know them!",
        )
    except Exception as ef:
        await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
        return

    text = (
        f"<b>Characteristics:</b>\n"
        f"<b>ID:</b> <code>{user.id}</code>\n"
        f"<b>First Name:</b> <code>{escape(user.first_name)}</code>"
    )

    if user.last_name:
        text += f"\n<b>Last Name:</b></b> <code>{escape(user.last_name)}</code>"

    if user.username:
        text += f"\n<b>Username</b>: @{escape(user.username)}"

    text += (
        f"\n<b>Permanent user link:</b> {(await mention_html('Click Here', user.id))}"
    )

    if user.id == OWNER_ID:
        text += "\n\nThis person is my Owner, I would never do anything against them!"
    elif user.id in DEV_USERS:
        text += "\n\nThis member is one of my Developers ⚡️"
    elif user.id in SUDO_USERS:
        text += "\n\nThe Power level of this person is 'Sudo'"
    elif user.id in WHITELIST_USERS:
        text += "\n\nThis person is 'Whitelist User', they cannot be banned!"

    try:
        user_member = await c.get_users(user.id)
        if user_member.status == "administrator":
            result = await AioHttp.post(
                (
                    f"https://api.telegram.org/bot{TOKEN}/"
                    f"getChatMember?chat_id={m.chat.id}&user_id={user.id}"
                ),
            )
            result = result.json()["result"]
            if "custom_title" in result.keys():
                custom_title = result["custom_title"]
                text += f"\n\nThis user holds the title <b>{custom_title}</b> here."
    except BaseException:
        LOGGER.error("BaseException")

    await infoMsg.edit_text(text, parse_mode="html", disable_web_page_preview=True)

    return


# Use split to convert to list
# Not using list itself becuase black changes it to long format...
normiefont = "a b c d e f g h i j k l m n o p q r s t u v w x y z".split()
weebyfont = "卂 乃 匚 刀 乇 下 厶 卄 工 丁 长 乚 从 𠘨 口 尸 㔿 尺 丂 丅 凵 リ 山 乂 丫 乙".split()


@Alita.on_message(filters.command("weebify", PREFIX_HANDLER))
async def weebify(_: Alita, m: Message):
    if len(m.text.split()) >= 2:
        args = m.text.split(None, 1)[1]
    if m.reply_to_message and len(m.text.split()) == 1:
        args = m.reply_to_message.text
    if not args:
        await m.reply_text("`What am I supposed to Weebify?`")
        return
    string = "  ".join(args).lower()
    for normiecharacter in string:
        if normiecharacter in normiefont:
            weebycharacter = weebyfont[normiefont.index(normiecharacter)]
            string = string.replace(normiecharacter, weebycharacter)

    await m.reply_text(f"**Weebified String:**\n`{string}`")

    return


@Alita.on_message(filters.command("paste", PREFIX_HANDLER))
async def paste_it(_: Alita, m: Message):

    replymsg = await m.reply_text("Pasting...", quote=True)

    if m.reply_to_message:
        txt = m.reply_to_message.text
    else:
        txt = m.text.split(None, 1)[1]

    url = (await paste(txt))[0]

    await replymsg.edit_text(
        "Pasted to NekoBin!",
        reply_markup=InlineKeyboardMarkup([[InlineKeyboardButton("NekoBin", url=url)]]),
    )

    return


@Alita.on_message(filters.command("tr", PREFIX_HANDLER))
async def translate(_: Alita, m: Message):
    _ = GetLang(m).strs
    translator = Translator()
    text = m.text[4:]
    lang = await get_lang(text)
    if m.reply_to_message:
        text = m.reply_to_message.text or m.reply_to_message.caption
    else:
        text = text.replace(lang, "", 1).strip() if text.startswith(lang) else text

    if text:
        sent = await m.reply_text(
            _("translate.translating"),
            reply_to_message_id=m.message_id,
        )
        langs = {}

        if len(lang.split("-")) > 1:
            langs["src"] = lang.split("-")[0]
            langs["dest"] = lang.split("-")[1]
        else:
            langs["dest"] = lang

        trres = translator.translate(text, **langs)
        text = trres.text

        res = escape(text)
        await sent.edit_text(
            _("translate.translation").format(
                from_lang=trres.src,
                to_lang=trres.dest,
                translation=res,
            ),
            parse_mode="HTML",
        )

    else:
        await m.reply_text(
            _("translate.translate_usage"),
            reply_to_message_id=m.message_id,
            parse_mode="markdown",
        )

    return


async def get_lang(text):
    if len(text.split()) > 0:
        lang = text.split()[0]
        if lang.split("-")[0] not in LANGUAGES:
            lang = "en"
        if len(lang.split("-")) > 1 and lang.split("-")[1] not in LANGUAGES:
            lang = "en"
    else:
        lang = "en"
    return lang
