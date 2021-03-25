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


import sys
from asyncio import create_subprocess_shell, sleep, subprocess
from io import BytesIO, StringIO
from time import gmtime, strftime, time
from traceback import format_exc

from pyrogram import filters
from pyrogram.errors import (
    ChannelInvalid,
    ChannelPrivate,
    ChatAdminRequired,
    FloodWait,
    MessageTooLong,
    PeerIdInvalid,
    RPCError,
)
from pyrogram.types import InlineKeyboardButton, InlineKeyboardMarkup, Message
from speedtest import Speedtest

from alita import DEV_PREFIX_HANDLER, LOGFILE, LOGGER, MESSAGE_DUMP, UPTIME
from alita.bot_class import Alita
from alita.database.chats_db import Chats
from alita.tr_engine import tlang
from alita.utils.aiohttp_helper import AioHttp
from alita.utils.clean_file import remove_markdown_and_html
from alita.utils.custom_filters import dev_filter, sudo_filter
from alita.utils.parser import mention_markdown
from alita.utils.paste import paste

# initialise database
chatdb = Chats()


@Alita.on_message(filters.command("ping", DEV_PREFIX_HANDLER) & sudo_filter)
async def ping(_, m: Message):
    LOGGER.info(f"{m.from_user.id} used ping cmd in {m.chat.id}")
    start = time()
    replymsg = await m.reply_text((tlang(m, "utils.ping.pinging")), quote=True)
    delta_ping = time() - start
    await replymsg.edit_text(f"<b>Pong!</b>\n{delta_ping * 1000:.3f} ms")
    return


@Alita.on_message(filters.command("logs", DEV_PREFIX_HANDLER) & dev_filter)
async def send_log(c: Alita, m: Message):

    replymsg = await m.reply_text("Sending logs...!")
    await c.send_message(
        MESSAGE_DUMP,
        f"#LOGS\n\n**User:** {(await mention_markdown(m.from_user.first_name, m.from_user.id))}",
    )
    # Send logs
    with open(LOGFILE) as f:
        raw = (await paste(f.read()))[1]
    await m.reply_document(
        document=LOGFILE,
        reply_markup=InlineKeyboardMarkup(
            [[InlineKeyboardButton("Logs", url=raw)]],
        ),
        quote=True,
    )
    await replymsg.delete()
    return


@Alita.on_message(filters.command("ginfo", DEV_PREFIX_HANDLER) & sudo_filter)
async def group_info(c: Alita, m: Message):

    if not len(m.text.split()) == 2:
        await m.reply_text(
            f"It works like this: <code>{DEV_PREFIX_HANDLER} chat_id</code>",
        )
        return

    chat_id = m.text.split(None, 1)[1]

    replymsg = await m.reply_text("Fetching info about group...!")
    grp_data = await c.get_chat(chat_id)
    msg = (
        f"Information for group: {chat_id}\n\n"
        f"Group Name: {grp_data['title']}\n"
        f"Members Count: {grp_data['members_count']}\n"
        f"Type: {grp_data['type']}\n"
        f"Group ID: {grp_data['id']}"
    )
    await replymsg.edit_text(msg)
    return


@Alita.on_message(filters.command("speedtest", DEV_PREFIX_HANDLER) & dev_filter)
async def test_speed(c: Alita, m: Message):

    await c.send_message(
        MESSAGE_DUMP,
        f"#SPEEDTEST\n\n**User:** {(await mention_markdown(m.from_user.first_name, m.from_user.id))}",
    )
    sent = await m.reply_text(tlang(m, "dev.speedtest.start_speedtest"))
    s = Speedtest()
    bs = s.get_best_server()
    dl = round(s.download() / 1024 / 1024, 2)
    ul = round(s.upload() / 1024 / 1024, 2)
    await sent.edit_text(
        (tlang(m, "dev.speedtest.speedtest_txt")).format(
            host=bs["sponsor"],
            ping=int(bs["latency"]),
            download=dl,
            upload=ul,
        ),
    )
    return


@Alita.on_message(filters.command("neofetch", DEV_PREFIX_HANDLER) & dev_filter)
async def neofetch_stats(_, m: Message):
    cmd = "neofetch --stdout"

    process = await create_subprocess_shell(
        cmd,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    stdout, stderr = await process.communicate()
    e = stderr.decode()
    if not e:
        e = "No Error"
    OUTPUT = stdout.decode()
    if not OUTPUT:
        OUTPUT = "No Output"

    try:
        await m.reply_text(OUTPUT, quote=True)
    except MessageTooLong:
        with BytesIO(str.encode(await remove_markdown_and_html(OUTPUT))) as f:
            f.name = "neofetch.txt"
            await m.reply_document(document=f, caption="neofetch result")
        await m.delete()
    return


@Alita.on_message(filters.command(["eval", "py"], DEV_PREFIX_HANDLER) & dev_filter)
async def evaluate_code(c: Alita, m: Message):

    if len(m.text.split()) == 1:
        await m.reply_text(tlang(m, "dev.execute_cmd_err"))
        return
    sm = await m.reply_text("`Processing...`")
    cmd = m.text.split(None, maxsplit=1)[1]

    reply_to_id = m.message_id
    if m.reply_to_message:
        reply_to_id = m.reply_to_message.message_id

    old_stderr = sys.stderr
    old_stdout = sys.stdout
    redirected_output = sys.stdout = StringIO()
    redirected_error = sys.stderr = StringIO()
    stdout, stderr, exc = None, None, None

    try:
        await aexec(cmd, c, m)
    except Exception as ef:
        LOGGER.error(ef)
        exc = format_exc()

    stdout = redirected_output.getvalue()
    stderr = redirected_error.getvalue()
    sys.stdout = old_stdout
    sys.stderr = old_stderr

    evaluation = ""
    if exc:
        evaluation = exc
    elif stderr:
        evaluation = stderr
    elif stdout:
        evaluation = stdout
    else:
        evaluation = "Success"

    final_output = f"<b>EVAL</b>: <code>{cmd}</code>\n\n<b>OUTPUT</b>:\n<code>{evaluation.strip()}</code> \n"

    try:
        await sm.edit(final_output)
    except MessageTooLong:
        with BytesIO(str.encode(await remove_markdown_and_html(final_output))) as f:
            f.name = "py.txt"
            await m.reply_document(
                document=f,
                caption=cmd,
                disable_notification=True,
                reply_to_message_id=reply_to_id,
            )
        await sm.delete()

    return


async def aexec(code, c, m):
    exec("async def __aexec(c, m): " + "".join(f"\n {l}" for l in code.split("\n")))
    return await locals()["__aexec"](c, m)


@Alita.on_message(filters.command(["exec", "sh"], DEV_PREFIX_HANDLER) & dev_filter)
async def execution(_, m: Message):

    if len(m.text.split()) == 1:
        await m.reply_text(tlang(m, "dev.execute_cmd_err"))
        return
    sm = await m.reply_text("`Processing...`")
    cmd = m.text.split(maxsplit=1)[1]
    reply_to_id = m.message_id
    if m.reply_to_message:
        reply_to_id = m.reply_to_message.message_id

    process = await create_subprocess_shell(
        cmd,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    stdout, stderr = await process.communicate()
    e = stderr.decode()
    if not e:
        e = "No Error"
    o = stdout.decode()
    if not o:
        o = "No Output"

    OUTPUT = ""
    OUTPUT += f"<b>QUERY:</b>\n<u>Command:</u>\n<code>{cmd}</code> \n"
    OUTPUT += f"<u>PID</u>: <code>{process.pid}</code>\n\n"
    OUTPUT += f"<b>stderr</b>: \n<code>{e}</code>\n\n"
    OUTPUT += f"<b>stdout</b>: \n<code>{o}</code>"

    try:
        await sm.edit_text(OUTPUT)
    except MessageTooLong:
        with BytesIO(str.encode(await remove_markdown_and_html(OUTPUT))) as f:
            f.name = "sh.txt"
            await m.reply_document(
                document=f,
                caption=cmd,
                disable_notification=True,
                reply_to_message_id=reply_to_id,
            )
        await sm.delete()
    return


@Alita.on_message(filters.command("ip", DEV_PREFIX_HANDLER) & dev_filter)
async def public_ip(c: Alita, m: Message):

    ip = (await AioHttp.get_text("https://api.ipify.org"))[0]
    await c.send_message(
        MESSAGE_DUMP,
        f"#IP\n\n**User:** {(await mention_markdown(m.from_user.first_name, m.from_user.id))}",
    )
    await m.reply_text(
        (tlang(m, "dev.bot_ip")).format(ip=f"<code>{ip}</code>"),
        quote=True,
    )
    return


@Alita.on_message(filters.command("chatlist", DEV_PREFIX_HANDLER) & dev_filter)
async def chats(c: Alita, m: Message):
    exmsg = await m.reply_text(tlang(m, "dev.chatlist.exporting"))
    await c.send_message(
        MESSAGE_DUMP,
        f"#CHATLIST\n\n**User:** {(await mention_markdown(m.from_user.first_name, m.from_user.id))}",
    )
    all_chats = (chatdb.get_all_chats()) or {}
    chatfile = tlang(m, "dev.chatlist.header")
    P = 1
    for chat, val in all_chats:
        try:
            chat_info = await c.get_chat(chat["_id"])
            chat_members = chat_info.members_count
            try:
                invitelink = chat_info.invite_link
            except KeyError:
                invitelink = "No Link!"
            chatfile += "{}. {} | {} | {} | {}\n".format(
                P,
                val["chat_name"],
                chat,
                chat_members,
                invitelink,
            )
            P += 1
        except ChatAdminRequired:
            pass
        except (ChannelPrivate, ChannelInvalid):
            chatdb.remove_chat(chat["_id"])
        except PeerIdInvalid:
            LOGGER.warning(f"Peer not found {chat['_id']}")
        except FloodWait as ef:
            LOGGER.error("FloodWait required, Sleeping for 60s")
            LOGGER.error(ef)
            sleep(60)
        except RPCError as ef:
            LOGGER.error(ef)
            await m.reply_text(f"**Error:**\n{ef}")

    with BytesIO(str.encode(await remove_markdown_and_html(chatfile))) as f:
        f.name = "chatlist.txt"
        await m.reply_document(
            document=f,
            caption=(tlang(m, "dev.chatlist.chats_in_db")),
        )
    await exmsg.delete()
    return


@Alita.on_message(filters.command("uptime", DEV_PREFIX_HANDLER) & dev_filter)
async def uptime(_, m: Message):
    up = strftime("%Hh %Mm %Ss", gmtime(time() - UPTIME))
    await m.reply_text((tlang(m, "dev.uptime")).format(uptime=up), quote=True)
    return


@Alita.on_message(filters.command("leavechat", DEV_PREFIX_HANDLER) & dev_filter)
async def leave_chat(c: Alita, m: Message):
    if len(m.text.split()) != 2:
        await m.reply_text("Supply a chat id which I should leave!", quoet=True)
        return

    chat_id = m.text.split(None, 1)[1]

    replymsg = await m.reply_text(f"Trying to leave chat {chat_id}...", quote=True)
    try:
        await c.send_message(chat_id, "Bye everyone!")
        await c.leave_chat(chat_id)
        await replymsg.edit_text(f"Left <code>{chat_id}</code>.")
    except PeerIdInvalid:
        await replymsg.edit_text("Haven't seen this group in this session")
    except RPCError as ef:
        LOGGER.error(ef)
        await replymsg.edit_text(f"Failed to leave chat!\nError: <code>{ef}</code>.")
    return


@Alita.on_message(filters.command("chatbroadcast", DEV_PREFIX_HANDLER) & dev_filter)
async def chat_broadcast(c: Alita, m: Message):
    if m.reply_to_message:
        msg = m.reply_to_message.text.markdown
    else:
        await m.reply_text("Reply to a message to broadcast it")
        return

    exmsg = await m.reply_text("Started broadcasting!")
    all_chats = (chatdb.list_chats()) or {}
    err_str = ""

    for chat in all_chats:
        try:
            await c.send_message(chat, msg)
        except RPCError as ef:
            LOGGER.error(ef)
            continue

    await exmsg.edit_text("Done broadcasting ✅")
    if err_str:
        with BytesIO(str.encode(await remove_markdown_and_html(err_str))) as f:
            f.name = "error_broadcast.txt"
            await m.reply_document(
                document=f,
                caption="Broadcast Error",
            )

    return
