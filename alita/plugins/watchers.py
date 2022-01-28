from re import escape as re_escape
from time import time
from traceback import format_exc

from pyrogram import filters
from pyrogram.errors import ChatAdminRequired, RPCError, UserAdminInvalid
from pyrogram.types import ChatPermissions, Message

from alita import LOGGER, MESSAGE_DUMP, SUPPORT_STAFF
from alita.bot_class import Alita
from alita.database.antispam_db import ANTISPAM_BANNED, GBan
from alita.database.approve_db import Approve
from alita.database.blacklist_db import Blacklist
from alita.database.group_blacklist import BLACKLIST_CHATS
from alita.database.pins_db import Pins
from alita.database.warns_db import Warns, WarnSettings
from alita.tr_engine import tlang
from alita.utils.caching import ADMIN_CACHE, admin_cache_reload
from alita.utils.parser import mention_html
from alita.utils.regex_utils import regex_searcher

# Initialise
gban_db = GBan()


@Alita.on_message(filters.linked_channel)
async def antichanpin_cleanlinked(c: Alita, m: Message):
    try:
        msg_id = m.message_id
        pins_db = Pins(m.chat.id)
        curr = pins_db.get_settings()
        if curr["antichannelpin"]:
            await c.unpin_chat_message(chat_id=m.chat.id, message_id=msg_id)
            LOGGER.info(f"AntiChannelPin: msgid-{m.message_id} unpinned in {m.chat.id}")
        if curr["cleanlinked"]:
            await c.delete_messages(m.chat.id, msg_id)
            LOGGER.info(f"CleanLinked: msgid-{m.message_id} cleaned in {m.chat.id}")
    except ChatAdminRequired:
        await m.reply_text(
            "Disabled antichannelpin as I don't have enough admin rights!",
        )
        pins_db.antichannelpin_off()
        LOGGER.warning(f"Disabled antichannelpin in {m.chat.id} as i'm not an admin.")
    except Exception as ef:
        LOGGER.error(ef)
        LOGGER.error(format_exc())

    return


@Alita.on_message(filters.text & filters.group, group=5)
async def bl_watcher(_, m: Message):
    if m and not m.from_user:
        return

    bl_db = Blacklist(m.chat.id)

    async def perform_action_blacklist(m: Message, action: str, trigger: str):
        if action == "kick":
            await m.chat.kick_member(m.from_user.id, int(time() + 45))
            await m.reply_text(
                tlang(m, "blacklist.bl_watcher.action_kick").format(
                    user=m.from_user.username or f"<b>{m.from_user.first_name}</b>",
                ),
            )

        elif action == "ban":
            (
                await m.chat.kick_member(
                    m.from_user.id,
                )
            )
            await m.reply_text(
                tlang(m, "blacklist.bl_watcher.action_ban").format(
                    user=m.from_user.username or f"<b>{m.from_user.first_name}</b>",
                ),
            )

        elif action == "mute":
            await m.chat.restrict_member(
                m.from_user.id,
                ChatPermissions(),
            )

            await m.reply_text(
                tlang(m, "blacklist.bl_watcher.action_mute").format(
                    user=m.from_user.username or f"<b>{m.from_user.first_name}</b>",
                ),
            )

        elif action == "warn":
            warns_settings_db = WarnSettings(m.chat.id)
            warns_db = Warns(m.chat.id)
            warn_settings = warns_settings_db.get_warnings_settings()
            warn_reason = bl_db.get_reason()
            _, num = warns_db.warn_user(m.from_user.id, warn_reason)
            if num >= warn_settings["warn_limit"]:
                if warn_settings["warn_mode"] == "kick":
                    await m.chat.ban_member(
                        m.from_user.id,
                        until_date=int(time() + 45),
                    )
                    action = "kicked"
                elif warn_settings["warn_mode"] == "ban":
                    await m.chat.ban_member(m.from_user.id)
                    action = "banned"
                elif warn_settings["warn_mode"] == "mute":
                    await m.chat.restrict_member(m.from_user.id, ChatPermissions())
                    action = "muted"
                await m.reply_text(
                    (
                        f"Warnings {num}/{warn_settings['warn_limit']}\n"
                        f"{(await mention_html(m.from_user.first_name, m.from_user.id))} has been <b>{action}!</b>"
                    ),
                )
                return
            await m.reply_text(
                (
                    f"{(await mention_html(m.from_user.first_name, m.from_user.id))} warned {num}/{warn_settings['warn_limit']}\n"
                    # f"Last warn was for:\n<i>{warn_reason}</i>"
                    f"Last warn was for:\n<i>{warn_reason.format(trigger)}</i>"
                ),
            )
        return

    if m.from_user.id in SUPPORT_STAFF:
        # Don't work on Support Staff!
        return

    # If no blacklists, then return
    chat_blacklists = bl_db.get_blacklists()
    if not chat_blacklists:
        return

    # Get admins from admin_cache, reduces api calls
    try:
        admin_ids = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admin_ids = await admin_cache_reload(m, "blacklist_watcher")

    if m.from_user.id in admin_ids:
        return

    # Get approved user from cache/database
    app_users = Approve(m.chat.id).list_approved()
    if m.from_user.id in {i[0] for i in app_users}:
        return

    # Get action for blacklist
    action = bl_db.get_action()
    for trigger in chat_blacklists:
        pattern = r"( |^|[^\w])" + re_escape(trigger) + r"( |$|[^\w])"
        match = await regex_searcher(pattern, m.text.lower())
        if not match:
            continue
        if match:
            try:
                await perform_action_blacklist(m, action, trigger)
                LOGGER.info(
                    f"{m.from_user.id} {action}ed for using blacklisted word {trigger} in {m.chat.id}",
                )
                await m.delete()
            except RPCError as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
            break
    return


@Alita.on_message(filters.user(list(ANTISPAM_BANNED)) & filters.group)
async def gban_watcher(c: Alita, m: Message):
    from alita import SUPPORT_GROUP

    if m and not m.from_user:
        return

    try:
        _banned = gban_db.check_gban(m.from_user.id)
    except Exception as ef:
        LOGGER.error(ef)
        LOGGER.error(format_exc())
        return

    if _banned:
        try:
            await m.chat.ban_member(m.from_user.id)
            await m.delete(m.message_id)  # Delete users message!
            await m.reply_text(
                (tlang(m, "antispam.watcher_banned")).format(
                    user_gbanned=(
                        await mention_html(m.from_user.first_name, m.from_user.id)
                    ),
                    SUPPORT_GROUP=SUPPORT_GROUP,
                ),
            )
            LOGGER.info(f"Banned user {m.from_user.id} in {m.chat.id} due to antispam")
            return
        except (ChatAdminRequired, UserAdminInvalid):
            # Bot not admin in group and hence cannot ban users!
            # TO-DO - Improve Error Detection
            LOGGER.info(
                f"User ({m.from_user.id}) is admin in group {m.chat.title} ({m.chat.id})",
            )
        except RPCError as ef:
            await c.send_message(
                MESSAGE_DUMP,
                tlang(m, "antispam.gban.gban_error_log").format(
                    chat_id=m.chat.id,
                    ef=ef,
                ),
            )
    return


@Alita.on_message(filters.chat(BLACKLIST_CHATS))
async def bl_chats_watcher(c: Alita, m: Message):
    from alita import SUPPORT_GROUP

    await c.send_message(
        m.chat.id,
        (
            "This is a blacklisted group!\n"
            f"For Support, Join @{SUPPORT_GROUP}\n"
            "Now, I'm outta here!"
        ),
    )
    await c.leave_chat(m.chat.id)
    LOGGER.info(f"Joined and Left blacklisted chat {m.chat.id}")
    return
