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


from html import escape
from io import BytesIO
from os import remove
from time import time

from googletrans import LANGUAGES, Translator
from pyrogram import filters
from pyrogram.errors import MessageTooLong, PeerIdInvalid, RPCError
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
from alita.tr_engine import tlang
from alita.utils.aiohttp_helper import AioHttp
from alita.utils.extract_user import extract_user
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
async def ping(_, m: Message):
    start = time()
    replymsg = await m.reply_text((await tlang(m, "utils.ping.pinging")), quote=True)
    delta_ping = time() - start
    await replymsg.edit_text(f"**Pong!**\n{delta_ping * 1000:.3f} ms")
    return


@Alita.on_message(
    filters.command("lyrics", PREFIX_HANDLER) & (filters.group | filters.private),
)
async def get_lyrics(_, m: Message):
    query = m.text.split(None, 1)[1]
    song = ""
    if not query:
        await m.edit_text(await tlang(m, "utils.song.no_song_given"))
        return
    em = await m.reply_text(
        (await tlang(m, "utils.song.searching").format(song_name=query)),
    )
    song = Song.find_song(query)
    if song:
        if song.lyrics:
            reply = song.format()
        else:
            reply = await tlang(m, "utils.song.no_lyrics_found")
    else:
        reply = await tlang(m, "utils.song.song_not_found")
    try:
        await em.edit_text(reply)
    except MessageTooLong:
        with BytesIO(str.encode(reply)) as f:
            f.name = "lyrics.txt"
            await m.reply_document(
                document=f,
            )
        await em.delete()
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
                (await tlang(m, "utils.id.id_main")).format(
                    orig_sender=(await mention_html(user2.first_name, user2.id)),
                    orig_id=f"<code>{user2.id}</code>",
                    fwd_sender=(await mention_html(user1.first_name, user1.id)),
                    fwd_id=f"<code>{user1.id}</code>",
                ),
                parse_mode="HTML",
            )
        else:
            try:
                user = await c.get_users(user_id)
            except PeerIdInvalid:
                await m.reply_text(await tlang(m, "utils.no_user_db"))

            await m.reply_text(
                f"{(await mention_html(user.first_name, user.id))}'s ID is <code>{user.id}</code>.",
                parse_mode="HTML",
            )
    else:
        if m.chat.type == "private":
            await m.reply_text(
                (await tlang(m, "utils.id.my_id")).format(
                    my_id=f"<code>{m.chat.id}</code>",
                ),
            )
        else:
            await m.reply_text(
                (await tlang(m, "utils.id.group_id")).format(
                    group_id=f"<code>{m.chat.id}</code>",
                ),
            )
    return


@Alita.on_message(
    filters.command("gifid", PREFIX_HANDLER) & (filters.group | filters.private),
)
async def get_gifid(_, m: Message):
    if m.reply_to_message and m.reply_to_message.animation:
        await m.reply_text(
            f"Gif ID:\n<code>{m.reply_to_message.animation.file_id}</code>",
            parse_mode="html",
        )
    else:
        await m.reply_text(await tlang(m, "utils.gif_id.reply_gif"))
    return


@Alita.on_message(
    filters.command("github", PREFIX_HANDLER) & (filters.group | filters.private),
)
async def github(_, m: Message):
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
    infoMsg = await m.reply_text(
        f"<code>{(await tlang(m, 'utils.user_info.getting_info'))}</code>",
    )
    user_id = (await extract_user(m))[0]
    try:
        user = await c.get_users(user_id)
    except PeerIdInvalid:
        await m.reply_text(await tlang(m, "utils.no_user_db"))
    except RPCError as ef:
        await m.reply_text(
            (await tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=f"@{SUPPORT_GROUP}",
                ef=f"<code>{ef}</code>",
            ),
        )
        return

    text = (await tlang(m, "utils.user_info.info_text.main")).format(
        user_id=user.id,
        user_name=escape(user.first_name),
    )

    if user.last_name:
        text += (await tlang(m, "utils.user_info.info_text.last_name")).format(
            user_lname=escape(user.last_name),
        )

    if user.username:
        text += (await tlang(m, "utils.user_info.info_text.username")).format(
            username=escape(user.username),
        )

    text += (await tlang(m, "utils.user_info.info_text.perma_link")).format(
        perma_link=(await mention_html("Click Here", user.id)),
    )

    if user.id == OWNER_ID:
        text += await tlang(m, "utils.user_info.support_user.owner")
    elif user.id in DEV_USERS:
        text += await tlang(m, "utils.user_info.support_user.dev")
    elif user.id in SUDO_USERS:
        text += await tlang(m, "utils.user_info.support_user.sudo")
    elif user.id in WHITELIST_USERS:
        text += await tlang(m, "utils.user_info.support_user.whitelist")

    try:
        user_member = await m.chat.get_member(user.id)
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
                text += (await tlang(m, "utils.user_info.custom_title")).format(
                    custom_title=f"<b>{custom_title}</b>",
                )
    except BaseException as ef:
        LOGGER.error(f"Error: {ef}")

    await infoMsg.edit_text(text, parse_mode="html", disable_web_page_preview=True)

    return


# Use split to convert to list
# Not using list itself becuase black changes it to long format...
normiefont = "a b c d e f g h i j k l m n o p q r s t u v w x y z".split()
weebyfont = "卂 乃 匚 刀 乇 下 厶 卄 工 丁 长 乚 从 𠘨 口 尸 㔿 尺 丂 丅 凵 リ 山 乂 丫 乙".split()


@Alita.on_message(filters.command("weebify", PREFIX_HANDLER))
async def weebify(_, m: Message):
    if len(m.text.split()) >= 2:
        args = m.text.split(None, 1)[1]
    if m.reply_to_message and len(m.text.split()) == 1:
        args = m.reply_to_message.text
    if not args:
        await m.reply_text(await tlang(m, "utils.weebify.weebify_what"))
        return
    string = "  ".join(args).lower()
    for normiecharacter in string:
        if normiecharacter in normiefont:
            weebycharacter = weebyfont[normiefont.index(normiecharacter)]
            string = string.replace(normiecharacter, weebycharacter)

    await m.reply_text(
        (await tlang(m, "utils.weebify.weebified_string").format(string=string)),
    )

    return


@Alita.on_message(filters.command("paste", PREFIX_HANDLER))
async def paste_it(_, m: Message):

    replymsg = await m.reply_text((await tlang(m, "utils.paste.pasting")), quote=True)

    if m.reply_to_message:
        if m.reply_to_message.document:
            dl_loc = await m.reply_to_message.download()
            with open(dl_loc) as f:
                txt = f.read()
            remove(dl_loc)
        else:
            txt = m.reply_to_message.text
    else:
        txt = m.text.split(None, 1)[1]

    url = (await paste(txt))[0]

    await replymsg.edit_text(
        (await tlang(m, "utils.paste.pasted")),
        reply_markup=InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        (await tlang(m, "utils.paste.nekobin_btn")),
                        url=url,
                    ),
                ],
            ],
        ),
    )

    return


@Alita.on_message(filters.command("tr", PREFIX_HANDLER))
async def translate(_, m: Message):

    translator = Translator()
    text = m.text[4:]
    lang = await get_lang(text)
    if m.reply_to_message:
        text = m.reply_to_message.text or m.reply_to_message.caption
    else:
        text = text.replace(lang, "", 1).strip() if text.startswith(lang) else text

    if text:
        sent = await m.reply_text(await tlang(m, "utils.translate.translating"))
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
            (await tlang(m, "utils.translate.translation")).format(
                from_lang=trres.src,
                to_lang=trres.dest,
                translation=res,
            ),
        )

    else:
        await m.reply_text(await tlang(m, "utils.translate.translate_usage"))

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
