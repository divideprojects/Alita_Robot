import sys
from asyncio import create_subprocess_shell, sleep, subprocess
from io import BytesIO, StringIO
from time import gmtime, strftime, time
from traceback import format_exc

from pyrogram.errors import (
    ChannelInvalid,
    ChannelPrivate,
    ChatAdminRequired,
    FloodWait,
    MessageTooLong,
    PeerIdInvalid,
    RPCError,
)
from pyrogram.types import Message
from speedtest import Speedtest

from alita import LOGFILE, LOGGER, MESSAGE_DUMP, UPTIME
from alita.bot_class import Alita
from alita.database.chats_db import Chats
from alita.tr_engine import tlang
from alita.utils.clean_file import remove_markdown_and_html
from alita.utils.custom_filters import command
from alita.utils.http_helper import HTTPx
from alita.utils.kbhelpers import ikb
from alita.utils.parser import mention_markdown
from alita.vars import Config


@Alita.on_message(command("ping", sudo_cmd=True))
async def ping(_, m: Message):
    LOGGER.info(f"{m.from_user.id} used ping cmd in {m.chat.id}")
    start = time()
    replymsg = await m.reply_text((tlang(m, "utils.ping.pinging")), quote=True)
    delta_ping = time() - start
    await replymsg.edit_text(f"<b>Pong!</b>\n{delta_ping * 1000:.3f} ms")
    return


@Alita.on_message(command("logs", dev_cmd=True))
async def send_log(c: Alita, m: Message):
    replymsg = await m.reply_text("Sending logs...!")
    await c.send_message(
        MESSAGE_DUMP,
        f"#LOGS\n\n**User:** {(await mention_markdown(m.from_user.first_name, m.from_user.id))}",
    )
    # Send logs
    with open(LOGFILE) as f:
        raw = (await (f.read()))[1]
    await m.reply_document(
        document=LOGFILE,
        quote=True,
    )
    await replymsg.delete()
    return


@Alita.on_message(command("ginfo", sudo_cmd=True))
async def group_info(c: Alita, m: Message):
    if len(m.text.split()) != 2:
        await m.reply_text(
            f"It works like this: <code>{Config.PREFIX_HANDLER} chat_id</code>",
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


@Alita.on_message(command("speedtest", dev_cmd=True))
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


@Alita.on_message(command("neofetch", dev_cmd=True))
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


@Alita.on_message(command(["eval", "py"], dev_cmd=True))
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


@Alita.on_message(command(["exec", "sh"], dev_cmd=True))
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


@Alita.on_message(command("ip", dev_cmd=True))
async def public_ip(c: Alita, m: Message):
    ip = await HTTPx.get("https://api.ipify.org")
    await c.send_message(
        MESSAGE_DUMP,
        f"#IP\n\n**User:** {(await mention_markdown(m.from_user.first_name, m.from_user.id))}",
    )
    await m.reply_text(
        (tlang(m, "dev.bot_ip")).format(ip=f"<code>{ip.text}</code>"),
        quote=True,
    )
    return


@Alita.on_message(command("chatlist", dev_cmd=True))
async def chats(c: Alita, m: Message):
    exmsg = await m.reply_text(tlang(m, "dev.chatlist.exporting"))
    await c.send_message(
        MESSAGE_DUMP,
        f"#CHATLIST\n\n**User:** {(await mention_markdown(m.from_user.first_name, m.from_user.id))}",
    )
    all_chats = (Chats.list_chats_full()) or {}
    chatfile = tlang(m, "dev.chatlist.header")
    P = 1
    for chat in all_chats:
        try:
            chat_info = await c.get_chat(chat["_id"])
            chat_members = chat_info.members_count
            try:
                invitelink = chat_info.invite_link
            except KeyError:
                invitelink = "No Link!"
            chatfile += f"{P}. {chat['chat_name']} | {chat['_id']} | {chat_members} | {invitelink}\n"
            P += 1
        except ChatAdminRequired:
            pass
        except (ChannelPrivate, ChannelInvalid):
            Chats.remove_chat(chat["_id"])
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


@Alita.on_message(command("uptime", dev_cmd=True))
async def uptime(_, m: Message):
    up = strftime("%Hh %Mm %Ss", gmtime(time() - UPTIME))
    await m.reply_text((tlang(m, "dev.uptime")).format(uptime=up), quote=True)
    return


@Alita.on_message(command("leavechat", dev_cmd=True))
async def leave_chat(c: Alita, m: Message):
    if len(m.text.split()) != 2:
        await m.reply_text("Supply a chat id which I should leave!", quoet=True)
        return

    chat_id = m.text.split(None, 1)[1]

    replymsg = await m.reply_text(f"Trying to leave chat {chat_id}...", quote=True)
    try:
        await c.leave_chat(chat_id)
        await replymsg.edit_text(f"Left <code>{chat_id}</code>.")
    except PeerIdInvalid:
        await replymsg.edit_text("Haven't seen this group in this session!")
    except RPCError as ef:
        LOGGER.error(ef)
        await replymsg.edit_text(f"Failed to leave chat!\nError: <code>{ef}</code>.")
    return


@Alita.on_message(command("chatbroadcast", dev_cmd=True))
async def chat_broadcast(c: Alita, m: Message):
    if m.reply_to_message:
        msg = m.reply_to_message.text.markdown
    else:
        await m.reply_text("Reply to a message to broadcast it")
        return

    exmsg = await m.reply_text("Started broadcasting!")
    all_chats = (Chats.list_chats_by_id()) or {}
    err_str, done_broadcast = "", 0

    for chat in all_chats:
        try:
            await c.send_message(chat, msg, disable_web_page_preview=True)
            done_broadcast += 1
            await sleep(0.1)
        except RPCError as ef:
            LOGGER.error(ef)
            err_str += str(ef)
            continue

    await exmsg.edit_text(
        f"Done broadcasting ✅\nSent message to {done_broadcast} chats",
    )

    if err_str:
        with BytesIO(str.encode(await remove_markdown_and_html(err_str))) as f:
            f.name = "error_broadcast.txt"
            await m.reply_document(
                document=f,
                caption="Broadcast Error",
            )

    return


_DISABLE_CMDS_ = ["ping"]
