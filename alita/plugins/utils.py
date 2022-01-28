from html import escape
from io import BytesIO
from os import remove

from gpytranslate import Translator
from pyrogram import filters
from pyrogram.errors import MessageTooLong, PeerIdInvalid, RPCError
from pyrogram.types import Message
from tswift import Song
from wikipedia import summary
from wikipedia.exceptions import DisambiguationError, PageError

from alita import (
    DEV_USERS,
    LOGGER,
    OWNER_ID,
    SUDO_USERS,
    SUPPORT_GROUP,
    SUPPORT_STAFF,
    WHITELIST_USERS,
)
from alita.bot_class import Alita
from alita.database.antispam_db import GBan
from alita.database.users_db import Users
from alita.tr_engine import tlang
from alita.utils.clean_file import remove_markdown_and_html
from alita.utils.custom_filters import command
from alita.utils.extract_user import extract_user
from alita.utils.http_helper import HTTPx, http
from alita.utils.kbhelpers import ikb
from alita.utils.parser import mention_html
from alita.vars import Config

gban_db = GBan()


@Alita.on_message(command("wiki"))
async def wiki(_, m: Message):
    LOGGER.info(f"{m.from_user.id} used wiki cmd in {m.chat.id}")

    if len(m.text.split()) <= 1:
        return await m.reply_text(tlang(m, "general.check_help"))

    search = m.text.split(None, 1)[1]
    try:
        res = summary(search)
    except DisambiguationError as de:
        return await m.reply_text(
            f"Disambiguated pages found! Adjust your query accordingly.\n<i>{de}</i>",
            parse_mode="html",
        )
    except PageError as pe:
        return await m.reply_text(f"<code>{pe}</code>", parse_mode="html")
    if res:
        result = f"<b>{search}</b>\n\n"
        result += f"<i>{res}</i>\n"
        result += f"""<a href="https://en.wikipedia.org/wiki/{search.replace(" ", "%20")}">Read more...</a>"""
        try:
            return await m.reply_text(
                result,
                parse_mode="html",
                disable_web_page_preview=True,
            )
        except MessageTooLong:
            with BytesIO(str.encode(await remove_markdown_and_html(result))) as f:
                f.name = "result.txt"
                return await m.reply_document(
                    document=f,
                    quote=True,
                    parse_mode="html",
                )
    await m.stop_propagation()


@Alita.on_message(command("gdpr"))
async def gdpr_remove(_, m: Message):
    if m.from_user.id in SUPPORT_STAFF:
        await m.reply_text(
            "You're in my support staff, I cannot do that unless you are no longer a part of it!",
        )
        return

    Users(m.from_user.id).delete_user()
    await m.reply_text(
        "Your personal data has been deleted.\n"
        "Note that this will not unban you from any chats, as that is telegram data, not Alita data."
        " Flooding, warns, and gbans are also preserved, as of "
        "[this](https://ico.org.uk/for-organisations/guide-to-the-general-data-protection-regulation-gdpr/individual-rights/right-to-erasure/),"
        " which clearly states that the right to erasure does not apply 'for the performance of a task carried out in the public interest', "
        "as is the case for the aforementioned pieces of data.",
        disable_web_page_preview=True,
    )
    await m.stop_propagation()


@Alita.on_message(
    command("lyrics") & (filters.group | filters.private),
)
async def get_lyrics(_, m: Message):
    if len(m.text.split()) <= 1:
        await m.reply_text(tlang(m, "general.check_help"))
        return

    query = m.text.split(None, 1)[1]
    LOGGER.info(f"{m.from_user.id} used lyrics cmd in {m.chat.id}")
    song = ""
    if not query:
        await m.edit_text(tlang(m, "utils.song.no_song_given"))
        return

    em = await m.reply_text(
        (tlang(m, "utils.song.searching").format(song_name=query)),
    )
    song = Song.find_song(query)
    if song:
        if song.lyrics:
            reply = song.format()
        else:
            reply = tlang(m, "utils.song.no_lyrics_found")
    else:
        reply = tlang(m, "utils.song.song_not_found")
    try:
        await em.edit_text(reply)
    except MessageTooLong:
        with BytesIO(str.encode(await remove_markdown_and_html(reply))) as f:
            f.name = "lyrics.txt"
            await m.reply_document(
                document=f,
            )
        await em.delete()
    return


@Alita.on_message(
    command("id") & (filters.group | filters.private),
)
async def id_info(c: Alita, m: Message):
    LOGGER.info(f"{m.from_user.id} used id cmd in {m.chat.id}")

    if m.chat.type == "supergroup" and not m.reply_to_message:
        await m.reply_text((tlang(m, "utils.id.group_id")).format(group_id=m.chat.id))
        return

    if m.chat.type == "private" and not m.reply_to_message:
        await m.reply_text((tlang(m, "utils.id.my_id")).format(my_id=m.chat.id))
        return

    user_id, _, _ = await extract_user(c, m)
    if user_id:
        if m.reply_to_message and m.reply_to_message.forward_from:
            user1 = m.reply_to_message.from_user
            user2 = m.reply_to_message.forward_from
            await m.reply_text(
                (tlang(m, "utils.id.id_main")).format(
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
                await m.reply_text(tlang(m, "utils.no_user_db"))
                return

            await m.reply_text(
                f"{(await mention_html(user.first_name, user.id))}'s ID is <code>{user.id}</code>.",
                parse_mode="HTML",
            )
    elif m.chat.type == "private":
        await m.reply_text(
            (tlang(m, "utils.id.my_id")).format(
                my_id=f"<code>{m.chat.id}</code>",
            ),
        )
    else:
        await m.reply_text(
            (tlang(m, "utils.id.group_id")).format(
                group_id=f"<code>{m.chat.id}</code>",
            ),
        )
    return


@Alita.on_message(
    command("gifid") & (filters.group | filters.private),
)
async def get_gifid(_, m: Message):
    if m.reply_to_message and m.reply_to_message.animation:
        LOGGER.info(f"{m.from_user.id} used gifid cmd in {m.chat.id}")
        await m.reply_text(
            f"Gif ID:\n<code>{m.reply_to_message.animation.file_id}</code>",
            parse_mode="html",
        )
    else:
        await m.reply_text(tlang(m, "utils.gif_id.reply_gif"))
    return


@Alita.on_message(
    command("github") & (filters.group | filters.private),
)
async def github(_, m: Message):
    if len(m.text.split()) == 2:
        username = m.text.split(None, 1)[1]
        LOGGER.info(f"{m.from_user.id} used github cmd in {m.chat.id}")
    else:
        await m.reply_text(
            f"Usage: <code>{Config.PREFIX_HANDLER}github username</code>",
        )
        return

    URL = f"https://api.github.com/users/{username}"
    r = await HTTPx.get(URL)
    if r.status_code == 404:
        await m.reply_text(f"<code>{username}</code> not found", quote=True)
        return

    r_json = r.json()
    url = r_json.get("html_url", None)
    name = r_json.get("name", None)
    company = r_json.get("company", None)
    followers = r_json.get("followers", 0)
    following = r_json.get("following", 0)
    public_repos = r_json.get("public_repos", 0)
    bio = r_json.get("bio", None)
    created_at = r_json.get("created_at", "Not Found")

    REPLY = (
        f"<b>GitHub Info for @{username}:</b>"
        f"\n<b>Name:</b> <code>{name}</code>\n"
        f"<b>Bio:</b> <code>{bio}</code>\n"
        f"<b>URL:</b> {url}\n"
        f"<b>Public Repos:</b> {public_repos}\n"
        f"<b>Followers:</b> {followers}\n"
        f"<b>Following:</b> {following}\n"
        f"<b>Company:</b> <code>{company}</code>\n"
        f"<b>Created at:</b> <code>{created_at}</code>"
    )

    await m.reply_text(REPLY, quote=True, disable_web_page_preview=True)
    return


@Alita.on_message(
    command("info") & (filters.group | filters.private),
)
async def my_info(c: Alita, m: Message):
    try:
        user_id, name, user_name = await extract_user(c, m)
    except PeerIdInvalid:
        await m.reply_text(tlang(m, "utils.user_info.peer_id_error"))
        return
    except ValueError as ef:
        if "Peer id invalid" in str(ef):
            await m.reply_text(tlang(m, "utils.user_info.id_not_found"))
        return
    try:
        user = Users.get_user_info(int(user_id))
        name = user["name"]
        user_name = user["username"]
        user_id = user["_id"]
    except KeyError:
        LOGGER.warning(f"Calling api to fetch info about user {user_id}")
        user = await c.get_users(user_id)
        name = (
            escape(user["first_name"] + " " + user["last_name"])
            if user["last_name"]
            else user["first_name"]
        )
        user_name = user["username"]
        user_id = user["id"]
    except PeerIdInvalid:
        await m.reply_text(tlang(m, "utils.no_user_db"))
        return
    except (RPCError, Exception) as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        return

    gbanned, reason_gban = gban_db.get_gban(user_id)
    LOGGER.info(f"{m.from_user.id} used info cmd for {user_id} in {m.chat.id}")

    text = (tlang(m, "utils.user_info.info_text.main")).format(
        user_id=user_id,
        user_name=name,
    )

    if user_name:
        text += (tlang(m, "utils.user_info.info_text.username")).format(
            username=user_name,
        )

    text += (tlang(m, "utils.user_info.info_text.perma_link")).format(
        perma_link=(await mention_html("Click Here", user_id)),
    )

    if gbanned:
        text += f"\nThis user is Globally banned beacuse: {reason_gban}\n"

    if user_id == OWNER_ID:
        text += tlang(m, "utils.user_info.support_user.owner")
    elif user_id in DEV_USERS:
        text += tlang(m, "utils.user_info.support_user.dev")
    elif user_id in SUDO_USERS:
        text += tlang(m, "utils.user_info.support_user.sudo")
    elif user_id in WHITELIST_USERS:
        text += tlang(m, "utils.user_info.support_user.whitelist")

    await m.reply_text(text, parse_mode="html", disable_web_page_preview=True)
    return


@Alita.on_message(command("paste"))
async def paste_it(_, m: Message):
    replymsg = await m.reply_text((tlang(m, "utils.paste.pasting")), quote=True)
    try:
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
        ur = "https://hastebin.com/documents"
        r = await http.post(ur, json={"content": txt})
        url = f"https://hastebin.com/{r.json().get('key')}"
        await replymsg.edit_text(
            (tlang(m, "utils.paste.pasted_nekobin")),
            reply_markup=ikb([[((tlang(m, "utils.paste.nekobin_btn")), url, "url")]]),
        )
        LOGGER.info(f"{m.from_user.id} used paste cmd in {m.chat.id}")
    except Exception as e:
        await replymsg.edit_text(f"Error: {e}")
        return
    return


@Alita.on_message(command("tr"))
async def translate(_, m: Message):
    trl = Translator()
    if m.reply_to_message and (m.reply_to_message.text or m.reply_to_message.caption):
        if len(m.text.split()) == 1:
            target_lang = "en"
        else:
            target_lang = m.text.split()[1]
        if m.reply_to_message.text:
            text = m.reply_to_message.text
        else:
            text = m.reply_to_message.caption
    else:
        if len(m.text.split()) <= 2:
            await m.reply_text(
                "Provide lang code.\n[Available options](https://telegra.ph/Lang-Codes-11-08).\n<b>Usage:</b> <code>/tr en</code>",
            )
            return
        target_lang = m.text.split(None, 2)[1]
        text = m.text.split(None, 2)[2]
    detectlang = await trl.detect(text)
    try:
        tekstr = await trl(text, targetlang=target_lang)
    except ValueError as err:
        await m.reply_text(f"Error: <code>{str(err)}</code>")
        return
    LOGGER.info(f"{m.from_user.id} used translate cmd in {m.chat.id}")
    return await m.reply_text(
        f"<b>Translated:</b> from {detectlang} to {target_lang} \n<code>``{tekstr.text}``</code>",
    )


__PLUGIN__ = "utils"
_DISABLE_CMDS_ = [
    "paste",
    "wiki",
    "id",
    "gifid",
    "lyrics",
    "tr",
    "github",
    "git",
    "info",
]
__alt_name__ = ["util", "misc", "tools"]
