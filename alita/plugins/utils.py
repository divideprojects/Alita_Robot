import html
import os
import aiohttp
from tswift import Song
from datetime import datetime
from alita.__main__ import Alita
from pyrogram import filters, errors
from pyrogram.types import Message
from alita import (
    PREFIX_HANDLER,
    OWNER_ID,
    DEV_USERS,
    SUDO_USERS,
    WHITELIST_USERS,
    TOKEN,
    SUPPORT_GROUP,
    LOGGER,
)
from alita.utils.aiohttp_helper import AioHttp
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html

__PLUGIN__ = "Utils"

__help__ = """
Some utils provided by bot to make your tasks easy!

 × /id: Get the current group id. If used by replying to a message, get that user's id.
 × /info: Get information about a user.
 × /ping - Get ping time of bot to telegram server.
 × /gifid: Reply to a gif to me to tell you its file ID.
 × /github <username>: Search for the user using github api!
 × /lyrics <song>: Get the lyrics of the song you specify!
 × /weebify <text> or a reply to message: To weebify the message.
"""


@Alita.on_message(
    filters.command("ping", PREFIX_HANDLER) & (filters.group | filters.private)
)
async def ping(c: Client m: Message):
    first = datetime.now()
    sent = await m.reply_text("**Ping...**")
    second = datetime.now()
    await sent.edit_text(
        f"**Pong!**\n`{round(((second-first).microseconds / 1000000), 2)}` Secs"
    )
    return


@Alita.on_message(
    filters.command("lyrics", PREFIX_HANDLER) & (filters.group | filters.private)
)
async def get_lyrics(c: Client m: Message):
    query = m.text.split()[1]
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
        with open("lyrics.txt", "w") as f:
            f.write(reply)
            f.close()
        await m.reply_document(
            document="lyrics.txt",
            caption=("Message length exceeded max limit!\nSent as a text file."),
        )
        os.remove("lyrics.txt")
        await em.delete()
    else:
        await em.edit_text(reply)
    return


@Alita.on_message(
    filters.command("id", PREFIX_HANDLER) & (filters.group | filters.private)
)
async def id_info(c: Alita, m: Message):
    user_id = extract_user(m)[0]
    if user_id:
        if m.reply_to_message and m.reply_to_message.forward_from:
            user1 = m.reply_to_m.from_user
            user2 = m.reply_to_m.forward_from
            await m.reply_text(
                (
                    f"Original Sender - {mention_html(user2.first_name, user2.id)} "
                    f"(<code>{user2.id}</code>).\n"
                    f"Forwarder - {mention_html(user1.first_name, user1.id)} "
                    f"(<code>{user1.id}</code>)."
                ),
                parse_mode="HTML",
            )
        else:
            try:
                user = await c.get_users(user_id)
            except errors.PeerIdInvalid:
                await m.reply_text(
                    "Failed to get user\nPeer ID invalid, I haven't seen this user anywhere earlier, maybe username would help to know them!"
                )

            await m.reply_text(
                f"{mention_html(user.first_name, user.id)}'s ID is <code>{user.id}</code>.",
                parse_mode="HTML",
            )
    else:
        if m.chat.type == "private":
            await m.reply_text(
                f"Your ID is <code>{m.chat.id}</code>.", parse_mode="HTML"
            )
        else:
            await m.reply_text(
                f"This Group's ID is <code>{m.chat.id}</code>.", parse_mode="HTML"
            )
    return


@Alita.on_message(
    filters.command("gifid", PREFIX_HANDLER) & (filters.group | filters.private)
)
async def get_gifid(c: Client m: Message):
    if m.reply_to_message and m.reply_to_message.animation:
        await m.reply_text(
            f"Gif ID:\n<code>{m.reply_to_message.animation.file_id}</code>",
            parse_mode="html",
        )
    else:
        await m.reply_text("Please reply to a gif to get its ID.")
    return


@Alita.on_message(
    filters.command("github", PREFIX_HANDLER) & (filters.group | filters.private)
)
async def github(c: Client m: Message):
    if len(m.text.split()) == 2:
        username = m.text.split(None, 1)[1]
    else:
        await m.reply_text(
            f"Usage: `{PREFIX_HANDLER}github <username>`", parse_mode="md"
        )
        return

    URL = f"https://api.github.com/users/{username}"
    async with aiohttp.ClientSession() as session:
        async with session.get(URL) as request:
            if request.status == 404:
                await m.reply_text(f"`{username} not found`", parse_mode="md")
                return

            result = await request.json()

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

            if not result.get("repos_url", None):
                return await m.reply_text(REPLY, parse_mode="md")
            async with session.get(result.get("repos_url", None)) as request:
                result = request.json
                if request.status == 404:
                    return await m.reply_text(REPLY, parse_mode="md")

                result = await request.json()

                REPLY += "\nRepos:\n"

                for nr in range(len(result)):
                    REPLY += f"[{result[nr].get('name', None)}]({result[nr].get('html_url', None)})\n"

                await m.reply_text(REPLY, parse_mode="md")
    return


@Alita.on_message(
    filters.command("info", PREFIX_HANDLER) & (filters.group | filters.private)
)
async def my_info(c: Alita, m: Message):
    infoMsg = await m.reply_text("<code>Getting user information...</code>")
    user_id = extract_user(m)[0]
    try:
        user = await c.get_users(user_id)
    except errors.PeerIdInvalid:
        await m.reply_text(
            "Failed to get user\nPeer ID invalid, I haven't seen this user anywhere earlier, maybe username would help to know them!"
        )
    except Exception as ef:
        await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
        return

    text = (
        f"<b>Characteristics:</b>\n"
        f"<b>ID:</b> <code>{user.id}</code>\n"
        f"<b>First Name:</b> <code>{html.escape(user.first_name)}</code>"
    )

    if user.last_name:
        text += f"\n<b>Last Name:</b></b> <code>{html.escape(user.last_name)}</code>"

    if user.username:
        text += f"\n<b>Username</b>: @{html.escape(user.username)}"

    text += f"\n<b>Permanent user link:</b> {mention_html('Click Here', user.id)}"

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
            result = AioHttp().post(
                (
                    f"https://api.telegram.org/bot{TOKEN}/"
                    f"getChatMember?chat_id={m.chat.id}&user_id={user.id}"
                )
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
async def weebify(c: Client m: Message):
    if len(m.text.split()) >= 2:
        args = m.text.split(" ", 1)[1]
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
