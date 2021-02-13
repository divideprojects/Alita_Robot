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


from traceback import format_exc
from io import BytesIO, StringIO
from time import strftime, gmtime, time
from speedtest import Speedtest
from asyncio import subprocess, create_subprocess_shell
from pyrogram import filters, errors
from pyrogram.types import Message
from alita import LOGGER, MESSAGE_DUMP, DEV_PREFIX_HANDLER, UPTIME, logfile
from alita.bot_class import Alita
from alita.db import users_db as userdb
from alita.utils.localization import GetLang
from alita.utils.aiohttp_helper import AioHttp
from alita.utils.custom_filters import dev_filter
from alita.utils.redishelper import get_key, allkeys
from alita.utils.parser import mention_markdown


@Alita.on_message(filters.command("logs", DEV_PREFIX_HANDLER) & dev_filter)
async def send_log(c: Alita, m: Message):
    _ = GetLang(m).strs
    rply = await m.reply_text("Sending logs...!")
    await c.send_message(
        m.chat.id,
        f"#LOGS\n\n**User:** {(await mention_markdown(m.from_user.first_name, m.from_user.id))}",
    )
    # Send logs
    await m.reply_document(document=logfile, quote=True)
    await rply.delete()
    return


@Alita.on_message(filters.command("speedtest", DEV_PREFIX_HANDLER) & dev_filter)
async def test_speed(c: Alita, m: Message):
    _ = GetLang(m).strs
    string = _("dev.speedtest")
    await c.send_message(
        MESSAGE_DUMP,
        f"#SPEEDTEST\n\n**User:** {(await mention_markdown(m.from_user.first_name, m.from_user.id))}",
    )
    sent = await m.reply_text(_("dev.start_speedtest"))
    s = Speedtest()
    bs = s.get_best_server()
    dl = round(s.download() / 1024 / 1024, 2)
    ul = round(s.upload() / 1024 / 1024, 2)
    await sent.edit_text(
        string.format(
            host=bs["sponsor"],
            ping=int(bs["latency"]),
            download=dl,
            upload=ul,
        ),
    )
    return


@Alita.on_message(filters.command("neofetch", DEV_PREFIX_HANDLER) & dev_filter)
async def neofetch_stats(c: Alita, m: Message):
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

    if len(OUTPUT) > 4090:
        with BytesIO(str.encode(OUTPUT)) as f:
            f.name = "neofetch.txt"
            await m.reply_document(document=f, caption="neofetch result")
        await m.delete()
    else:
        await m.reply_text(OUTPUT, quote=True)
    return


@Alita.on_message(filters.command(["eval", "py"], DEV_PREFIX_HANDLER) & dev_filter)
async def evaluate_code(c: Alita, m: Message):
    _ = GetLang(m).strs
    if len(m.text.split()) == 1:
        await m.reply_text(_("dev.execute_cmd_err"))
        return
    sm = await m.reply_text("`Processing...`")
    cmd = m.text.split(None, maxsplit=1)[1]

    reply_to_id = m.message_id
    if m.reply_to_message:
        reply_to_id = m.reply_to_message.message_id

    redirected_output = stdout = StringIO()
    redirected_error = stderr = StringIO()
    stdout, stderr, exc = None, None, None

    try:
        await aexec(cmd, c, m)
    except Exception as ef:
        LOGGER.error(ef)
        exc = format_exc()

    stdout = redirected_output.getvalue()
    stderr = redirected_error.getvalue()

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

    if len(final_output) > 4000:
        with BytesIO(str.encode(final_output)) as f:
            f.name = "eval.txt"
            await m.reply_document(
                document=f,
                caption=cmd,
                disable_notification=True,
                reply_to_message_id=reply_to_id,
            )
        await sm.delete()
    else:
        await sm.edit(final_output)
    return


async def aexec(code, c, m):
    exec("async def __aexec(c, m): " + "".join(f"\n {l}" for l in code.split("\n")))
    return await locals()["__aexec"](c, m)


@Alita.on_message(filters.command(["exec", "sh"], DEV_PREFIX_HANDLER) & dev_filter)
async def execution(c: Alita, m: Message):
    _ = GetLang(m).strs
    if len(m.text.split()) == 1:
        await m.reply_text(_("dev.execute_cmd_err"))
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

    if len(OUTPUT) > 4000:
        with BytesIO(str.encode(OUTPUT)) as f:
            f.name = "exec.txt"
            await m.reply_document(
                document=f,
                caption=cmd,
                disable_notification=True,
                reply_to_message_id=reply_to_id,
            )
        await sm.delete()
    else:
        await sm.edit_text(OUTPUT)
    return


@Alita.on_message(filters.command("ip", DEV_PREFIX_HANDLER) & dev_filter)
async def public_ip(c: Alita, m: Message):
    _ = GetLang(m).strs
    ip = (await AioHttp.get_text("https://api.ipify.org"))[0]
    await c.send_message(
        MESSAGE_DUMP,
        f"#IP\n\n**User:** {(await mention_markdown(m.from_user.first_name, m.from_user.id))}",
    )
    await m.reply_text(_("dev.bot_ip").format(ip=ip))
    return


@Alita.on_message(filters.command("chatlist", DEV_PREFIX_HANDLER) & dev_filter)
async def chats(c: Alita, m: Message):
    exmsg = await m.reply_text("`Exporting Chatlist...`")
    await c.send_message(
        MESSAGE_DUMP,
        f"#CHATLIST\n\n**User:** {(await mention_markdown(m.from_user.first_name, m.from_user.id))}",
    )
    all_chats = userdb.get_all_chats() or []
    chatfile = "List of chats.\n\nChat name | Chat ID | Members count\n"
    P = 1
    for chat in all_chats:
        try:
            chat_info = await c.get_chat(chat.chat_id)
            chat_members = chat_info.members_count
            try:
                invitelink = chat_info.invite_link
            except KeyError:
                invitelink = "No Link!"
            chatfile += "{}. {} | {} | {} | {}\n".format(
                P,
                chat.chat_name,
                chat.chat_id,
                chat_members,
                invitelink,
            )
            P += 1
        except errors.ChatAdminRequired:
            pass
        except errors.ChannelPrivate:
            userdb.rem_chat(chat.chat_id)
        except Exception as ef:
            await m.reply_text(f"**Error:**\n{ef}")

    with BytesIO(str.encode(chatfile)) as output:
        output.name = "chatlist.txt"
        await m.reply_document(
            document=output,
            caption="Here is the list of chats in my Database.",
        )
    await exmsg.delete()
    return


@Alita.on_message(filters.command("uptime", DEV_PREFIX_HANDLER) & dev_filter)
async def uptime(c: Alita, m: Message):
    up = strftime("%Hh %Mm %Ss", gmtime(time() - UPTIME))
    await m.reply_text(f"<b>Uptime:</b> `{up}`")
    return


@Alita.on_message(filters.command("loadmembers", DEV_PREFIX_HANDLER) & dev_filter)
async def store_members(c: Alita, m: Message):
    sm = await m.reply_text("Updating Members...")

    lv = 0  # lv = local variable

    try:
        async for member in m.chat.iter_members():
            try:
                userdb.update_user(
                    member.user.id,
                    member.user.username,
                    m.chat.id,
                    m.chat.title,
                )
                lv += 1
            except BaseException:
                pass
    except BaseException:
        await c.send_message(chat_id=MESSAGE_DUMP, text="Error while storing members!")
        return
    await sm.edit_text(f"Stored {lv} members")
    return


@Alita.on_message(filters.command("alladmins", DEV_PREFIX_HANDLER) & dev_filter)
async def list_all_admins(c: Alita, m: Message):

    admindict = await get_key("ADMINDICT")

    if len(str(admindict)) > 4000:
        with BytesIO(str.encode(admindict)) as output:
            output.name = "alladmins.txt"
            await m.reply_document(
                document=output,
                caption="Here is the list of all admins in my Cache.",
            )
    else:
        await m.reply_text(admindict)

    return


@Alita.on_message(filters.command("rediskeys", DEV_PREFIX_HANDLER) & dev_filter)
async def show_redis_keys(c: Alita, m: Message):
    keys = await allkeys()
    await m.reply_text(keys)
    return
