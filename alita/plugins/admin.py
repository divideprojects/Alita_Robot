from asyncio import sleep
from html import escape
from os import remove
from traceback import format_exc

from pyrogram import filters
from pyrogram.errors import (
    ChatAdminInviteRequired,
    ChatAdminRequired,
    FloodWait,
    RightForbidden,
    RPCError,
    UserAdminInvalid,
)
from pyrogram.types import Message

from alita import DEV_USERS, LOGGER, OWNER_ID, SUPPORT_GROUP, SUPPORT_STAFF
from alita.bot_class import Alita
from alita.database.approve_db import Approve
from alita.database.reporting_db import Reporting
from alita.tr_engine import tlang
from alita.utils.caching import ADMIN_CACHE, TEMP_ADMIN_CACHE_BLOCK, admin_cache_reload
from alita.utils.custom_filters import (
    DEV_LEVEL,
    admin_filter,
    command,
    owner_filter,
    promote_filter,
)
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html
from alita.vars import Config


@Alita.on_message(command("adminlist"))
async def adminlist_show(_, m: Message):
    if m.chat.type != "supergroup":
        return await m.reply_text(
            "This command is made to be used in groups only!",
        )
    try:
        try:
            admin_list = ADMIN_CACHE[m.chat.id]
            note = tlang(m, "admin.adminlist.note_cached")
        except KeyError:
            admin_list = await admin_cache_reload(m, "adminlist")
            note = tlang(m, "admin.adminlist.note_updated")

        adminstr = (tlang(m, "admin.adminlist.adminstr")).format(
            chat_title=m.chat.title,
        ) + "\n\n"

        bot_admins = [i for i in admin_list if (i[1].lower()).endswith("bot")]
        user_admins = [i for i in admin_list if not (i[1].lower()).endswith("bot")]

        # format is like: (user_id, username/name,anonyamous or not)
        mention_users = [
            (
                admin[1]
                if admin[1].startswith("@")
                else (await mention_html(admin[1], admin[0]))
            )
            for admin in user_admins
            if not admin[2]  # if non-anonyamous admin
        ]
        mention_users.sort(key=lambda x: x[1])

        mention_bots = [
            (
                admin[1]
                if admin[1].startswith("@")
                else (await mention_html(admin[1], admin[0]))
            )
            for admin in bot_admins
        ]
        mention_bots.sort(key=lambda x: x[1])

        adminstr += "<b>User Admins:</b>\n"
        adminstr += "\n".join(f"- {i}" for i in mention_users)
        adminstr += "\n\n<b>Bots:</b>\n"
        adminstr += "\n".join(f"- {i}" for i in mention_bots)

        await m.reply_text(adminstr + "\n\n" + note)
        LOGGER.info(f"Adminlist cmd use in {m.chat.id} by {m.from_user.id}")

    except Exception as ef:
        if str(ef) == str(m.chat.id):
            await m.reply_text(tlang(m, "admin.adminlist.use_admin_cache"))
        else:
            ef = f"{str(ef)}{admin_list}\n"
            await m.reply_text(
                (tlang(m, "general.some_error")).format(
                    SUPPORT_GROUP=SUPPORT_GROUP,
                    ef=ef,
                ),
            )
        LOGGER.error(ef)
        LOGGER.error(format_exc())

    return


@Alita.on_message(command("zombies") & owner_filter)
async def zombie_clean(c: Alita, m: Message):

    zombie = 0

    wait = await m.reply_text("Searching ... and banning ...")
    async for member in c.iter_chat_members(m.chat.id):
        if member.user.is_deleted:
            zombie += 1
            try:
                await c.kick_chat_member(m.chat.id, member.user.id)
            except UserAdminInvalid:
                zombie -= 1
            except FloodWait as e:
                await sleep(e.x)
    if zombie == 0:
        return await wait.edit_text("Group is clean!")
    return await wait.edit_text(
        f"<b>{zombie}</b> Zombies found and has been banned!",
    )


@Alita.on_message(command("admincache"))
async def reload_admins(_, m: Message):

    if m.chat.type != "supergroup":
        return await m.reply_text(
            "This command is made to be used in groups only!",
        )

    if (
        (m.chat.id in set(TEMP_ADMIN_CACHE_BLOCK.keys()))
        and (m.from_user.id not in SUPPORT_STAFF)
        and TEMP_ADMIN_CACHE_BLOCK[m.chat.id] == "manualblock"
    ):
        await m.reply_text("Can only reload admin cache once per 10 mins!")
        return

    try:
        await admin_cache_reload(m, "admincache")
        TEMP_ADMIN_CACHE_BLOCK[m.chat.id] = "manualblock"
        await m.reply_text(tlang(m, "admin.adminlist.reloaded_admins"))
        LOGGER.info(f"Admincache cmd use in {m.chat.id} by {m.from_user.id}")
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())
    return


@Alita.on_message(filters.regex(r"^(?i)@admin(s)?") & filters.group)
async def tag_admins(_, m: Message):
    db = Reporting(m.chat.id)
    if not db.get_settings():
        return

    try:
        admin_list = ADMIN_CACHE[m.chat.id]
    except KeyError:
        admin_list = await admin_cache_reload(m, "adminlist")

    user_admins = [i for i in admin_list if not (i[1].lower()).endswith("bot")]
    mention_users = [(await mention_html("\u2063", admin[0])) for admin in user_admins]
    mention_users.sort(key=lambda x: x[1])
    mention_str = "".join(mention_users)
    await m.reply_text(
        (
            f"{(await mention_html(m.from_user.first_name, m.from_user.id))}"
            f" reported the message to admins!{mention_str}"
        ),
    )


@Alita.on_message(command("fullpromote") & promote_filter)
async def fullpromote_usr(c: Alita, m: Message):

    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.promote.no_target"))
        return

    try:
        user_id, user_first_name, user_name = await extract_user(c, m)
    except Exception:
        return

    bot = await c.get_chat_member(m.chat.id, Config.BOT_ID)

    if user_id == Config.BOT_ID:
        await m.reply_text("Huh, how can I even promote myself?")
        return

    if not bot.can_promote_members:
        return await m.reply_text(
            "I don't have enough permissions!",
        )  # This should be here

    user = await c.get_chat_member(m.chat.id, m.from_user.id)
    if m.from_user.id not in [DEV_USERS, OWNER_ID] and user.status != "creator":
        return await m.reply_text("This command can only be used by chat owner.")
    # If user is alreay admin
    try:
        admin_list = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admin_list = {
            i[0] for i in (await admin_cache_reload(m, "promote_cache_update"))
        }

    if user_id in admin_list:
        await m.reply_text(
            "This user is already an admin, how am I supposed to re-promote them?",
        )
        return

    try:
        await m.chat.promote_member(
            user_id=user_id,
            can_change_info=bot.can_change_info,
            can_invite_users=bot.can_invite_users,
            can_delete_messages=bot.can_delete_messages,
            can_restrict_members=bot.can_restrict_members,
            can_pin_messages=bot.can_pin_messages,
            can_promote_members=bot.can_promote_members,
            can_manage_chat=bot.can_manage_chat,
            can_manage_voice_chats=bot.can_manage_voice_chats,
        )

        title = ""
        if len(m.text.split()) == 3 and not m.reply_to_message:
            title = m.text.split()[2]
        elif len(m.text.split()) == 2 and m.reply_to_message:
            title = m.text.split()[1]
        if title and len(title) > 16:
            title = title[:16]

        try:
            await c.set_administrator_title(m.chat.id, user_id, title)
        except RPCError as e:
            LOGGER.error(e)

        LOGGER.info(
            f"{m.from_user.id} fullpromoted {user_id} in {m.chat.id} with title '{title}'",
        )

        await m.reply_text(
            (tlang(m, "admin.promote.promoted_user")).format(
                promoter=(await mention_html(m.from_user.first_name, m.from_user.id)),
                promoted=(await mention_html(user_first_name, user_id)),
                chat_title=f"{escape(m.chat.title)} title set to {title}"
                if title
                else f"{escape(m.chat.title)} title set to Admin",
            ),
        )

        # If user is approved, disapprove them as they willbe promoted and get even more rights
        if Approve(m.chat.id).check_approve(user_id):
            Approve(m.chat.id).remove_approve(user_id)

        # ----- Add admin to temp cache -----
        try:
            inp1 = user_name or user_first_name
            admins_group = ADMIN_CACHE[m.chat.id]
            admins_group.append((user_id, inp1))
            ADMIN_CACHE[m.chat.id] = admins_group
        except KeyError:
            await admin_cache_reload(m, "promote_key_error")

    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.promote.bot_no_right"))
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
    except RPCError as e:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=e,
            ),
        )
        LOGGER.error(e)
        LOGGER.error(format_exc())
    return


@Alita.on_message(command("promote") & promote_filter)
async def promote_usr(c: Alita, m: Message):

    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.promote.no_target"))
        return

    try:
        user_id, user_first_name, user_name = await extract_user(c, m)
    except Exception:
        return

    bot = await c.get_chat_member(m.chat.id, Config.BOT_ID)

    if user_id == Config.BOT_ID:
        await m.reply_text("Huh, how can I even promote myself?")
        return

    if not bot.can_promote_members:
        return await m.reply_text(
            "I don't have enough permissions",
        )  # This should be here
    # If user is alreay admin
    try:
        admin_list = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admin_list = {
            i[0] for i in (await admin_cache_reload(m, "promote_cache_update"))
        }

    if user_id in admin_list:
        await m.reply_text(
            "This user is already an admin, how am I supposed to re-promote them?",
        )
        return

    try:
        await m.chat.promote_member(
            user_id=user_id,
            can_change_info=bot.can_change_info,
            can_invite_users=bot.can_invite_users,
            can_delete_messages=bot.can_delete_messages,
            can_restrict_members=bot.can_restrict_members,
            can_pin_messages=bot.can_pin_messages,
            can_manage_chat=bot.can_manage_chat,
            can_manage_voice_chats=bot.can_manage_voice_chats,
        )

        title = ""  # Deafult title
        if len(m.text.split()) == 3 and not m.reply_to_message:
            title = m.text.split()[2]
        elif len(m.text.split()) == 2 and m.reply_to_message:
            title = m.text.split()[1]
        if title and len(title) > 16:
            title = title[:16]

        try:
            await c.set_administrator_title(m.chat.id, user_id, title)
        except RPCError as e:
            LOGGER.error(e)

        LOGGER.info(
            f"{m.from_user.id} promoted {user_id} in {m.chat.id} with title '{title}'",
        )

        await m.reply_text(
            (tlang(m, "admin.promote.promoted_user")).format(
                promoter=(await mention_html(m.from_user.first_name, m.from_user.id)),
                promoted=(await mention_html(user_first_name, user_id)),
                chat_title=f"{escape(m.chat.title)} title set to {title}"
                if title
                else f"{escape(m.chat.title)} title set to Admin",
            ),
        )

        # If user is approved, disapprove them as they willbe promoted and get even more rights
        if Approve(m.chat.id).check_approve(user_id):
            Approve(m.chat.id).remove_approve(user_id)

        # ----- Add admin to temp cache -----
        try:
            inp1 = user_name or user_first_name
            admins_group = ADMIN_CACHE[m.chat.id]
            admins_group.append((user_id, inp1))
            ADMIN_CACHE[m.chat.id] = admins_group
        except KeyError:
            await admin_cache_reload(m, "promote_key_error")

    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.promote.bot_no_right"))
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
    except RPCError as e:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=e,
            ),
        )
        LOGGER.error(e)
        LOGGER.error(format_exc())
    return


@Alita.on_message(command("demote") & promote_filter)
async def demote_usr(c: Alita, m: Message):

    global ADMIN_CACHE

    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.demote.no_target"))
        return

    try:
        user_id, user_first_name, _ = await extract_user(c, m)
    except Exception:
        return

    if user_id == Config.BOT_ID:
        await m.reply_text("Get an admin to demote me!")
        return

    # If user not alreay admin
    try:
        admin_list = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admin_list = {
            i[0] for i in (await admin_cache_reload(m, "demote_cache_update"))
        }

    if user_id not in admin_list:
        await m.reply_text(
            "This user is not an admin, how am I supposed to re-demote them?",
        )
        return

    try:
        await m.chat.promote_member(
            user_id=user_id,
            can_change_info=False,
            can_invite_users=False,
            can_delete_messages=False,
            can_restrict_members=False,
            can_pin_messages=False,
            can_promote_members=False,
            can_manage_chat=False,
            can_manage_voice_chats=False,
        )
        LOGGER.info(f"{m.from_user.id} demoted {user_id} in {m.chat.id}")

        # ----- Remove admin from cache -----
        try:
            admin_list = ADMIN_CACHE[m.chat.id]
            user = next(user for user in admin_list if user[0] == user_id)
            admin_list.remove(user)
            ADMIN_CACHE[m.chat.id] = admin_list
        except (KeyError, StopIteration):
            await admin_cache_reload(m, "demote_key_stopiter_error")

        await m.reply_text(
            (tlang(m, "admin.demote.demoted_user")).format(
                demoter=(
                    await mention_html(
                        m.from_user.first_name,
                        m.from_user.id,
                    )
                ),
                demoted=(await mention_html(user_first_name, user_id)),
                chat_title=m.chat.title,
            ),
        )

    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.demote.bot_no_right"))
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())

    return


@Alita.on_message(command("invitelink"))
async def get_invitelink(c: Alita, m: Message):
    # Bypass the bot devs, sudos and owner
    if m.from_user.id not in DEV_LEVEL:
        user = await m.chat.get_member(m.from_user.id)

        if not user.can_invite_users and user.status != "creator":
            await m.reply_text(tlang(m, "admin.no_user_invite_perm"))
            return False

    try:
        link = await c.export_chat_invite_link(m.chat.id)
        await m.reply_text(
            (tlang(m, "admin.invitelink")).format(
                chat_name=m.chat.id,
                link=link,
            ),
            disable_web_page_preview=True,
        )
        LOGGER.info(f"{m.from_user.id} exported invite link in {m.chat.id}")
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except ChatAdminInviteRequired:
        await m.reply_text(tlang(m, "admin.no_invite_perm"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.no_user_invite_perm"))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())

    return


@Alita.on_message(command("setgtitle") & admin_filter)
async def setgtitle(_, m: Message):
    user = await m.chat.get_member(m.from_user.id)

    if not user.can_change_info and user.status != "creator":
        await m.reply_text(
            "You don't have enough permission to use this command!",
        )
        return False

    if len(m.command) < 1:
        return await m.reply_text("Please read /help for using it!")

    gtit = m.text.split(None, 1)[1]
    try:
        await m.chat.set_title(gtit)
    except Exception as e:
        return await m.reply_text(f"Error: {e}")
    return await m.reply_text(
        f"Successfully Changed Group Title From {m.chat.title} To {gtit}",
    )


@Alita.on_message(command("setgdes") & admin_filter)
async def setgdes(_, m: Message):

    user = await m.chat.get_member(m.from_user.id)
    if not user.can_change_info and user.status != "creator":
        await m.reply_text(
            "You don't have enough permission to use this command!",
        )
        return False

    if len(m.command) < 1:
        return await m.reply_text("Please read /help for using it!")

    desp = m.text.split(None, 1)[1]
    try:
        await m.chat.set_description(desp)
    except Exception as e:
        return await m.reply_text(f"Error: {e}")
    return await m.reply_text(
        f"Successfully Changed Group description From {m.chat.description} To {desp}",
    )


@Alita.on_message(command("title") & admin_filter)
async def set_user_title(c: Alita, m: Message):

    user = await m.chat.get_member(m.from_user.id)
    if not user.can_promote_members and user.status != "creator":
        await m.reply_text(
            "You don't have enough permission to use this command!",
        )
        return False

    if len(m.text.split()) == 1 and not m.reply_to_message:
        return await m.reply_text("To whom??")

    if m.reply_to_message:
        if len(m.text.split()) >= 2:
            reason = m.text.split(None, 1)[1]
    elif len(m.text.split()) >= 3:
        reason = m.text.split(None, 2)[2]
    try:
        user_id, _, _ = await extract_user(c, m)
    except Exception:
        return

    if not user_id:
        return await m.reply_text("Cannot find user!")

    if user_id == Config.BOT_ID:
        return await m.reply_text("Huh, why ?")

    if not reason:
        return await m.reply_text("Read /help please!")

    from_user = await c.get_users(user_id)
    title = reason
    try:
        await c.set_administrator_title(m.chat.id, from_user.id, title)
    except Exception as e:
        return await m.reply_text(f"Error: {e}")
    return await m.reply_text(
        f"Successfully Changed {from_user.mention}'s Admin Title To {title}",
    )


@Alita.on_message(command("setgpic") & admin_filter)
async def setgpic(c: Alita, m: Message):
    user = await m.chat.get_member(m.from_user.id)
    if not user.can_change_info and user.status != "creator":
        await m.reply_text(
            "You don't have enough permission to use this command!",
        )
        return False
    if not m.reply_to_message:
        return await m.reply_text("Reply to a photo to set it as chat photo")
    if not m.reply_to_message.photo and not m.reply_to_message.document:
        return await m.reply_text("Reply to a photo to set it as chat photo")
    photo = await m.reply_to_message.download()
    try:
        await m.chat.set_photo(photo)
    except Exception as e:
        remove(photo)
        return await m.reply_text(f"Error: {e}")
    await m.reply_text("Successfully Changed Group Photo!")
    remove(photo)


__PLUGIN__ = "admin"

__alt_name__ = [
    "admins",
    "promote",
    "demote",
    "adminlist",
    "setgpic",
    "title",
    "setgtitle",
    "fullpromote",
    "invitelink",
    "setgdes",
    "zombies",
]
