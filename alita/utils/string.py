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
from re import compile as compilere
from time import time

from pyrogram.types import InlineKeyboardButton, Message

from alita.utils.parser import mention_html

BTN_URL_REGEX = compilere(r"(\[([^\[]+?)\]\(buttonurl:(?:/{0,2})(.+?)(:same)?\))")


async def extract_time(m, time_val):
    """Extract time from message."""
    if any(time_val.endswith(unit) for unit in ("m", "h", "d")):
        unit = time_val[-1]
        time_num = time_val[:-1]  # type: str
        if not time_num.isdigit():
            await m.reply("Unspecified amount of time.")
            return ""

        if unit == "m":
            bantime = int(time() + int(time_num) * 60)
        elif unit == "h":
            bantime = int(time() + int(time_num) * 60 * 60)
        elif unit == "s":
            bantime = int(time() + int(time_num) * 24 * 60 * 60)
        else:
            # how even...?
            return ""
        return bantime
    await m.reply(
        "Invalid time type specified. Needed m, h, or s. got: {}".format(time_val[-1]),
    )
    return ""


async def parse_button(text):
    """Parse button from text."""
    markdown_note = text
    prev = 0
    note_data = ""
    buttons = []
    for match in BTN_URL_REGEX.finditer(markdown_note):
        # Check if btnurl is escaped
        n_escapes = 0
        to_check = match.start(1) - 1
        while to_check > 0 and markdown_note[to_check] == "\\":
            n_escapes += 1
            to_check -= 1

        # if even, not escaped -> create button
        if n_escapes % 2 == 0:
            # create a thruple with button label, url, and newline status
            buttons.append((match.group(2), match.group(3), bool(match.group(4))))
            note_data += markdown_note[prev : match.start(1)]
            prev = match.end(1)
        # if odd, escaped -> move along
        else:
            note_data += markdown_note[prev:to_check]
            prev = match.start(1) - 1

    note_data += markdown_note[prev:]

    return note_data, buttons


async def build_keyboard(buttons):
    """Build keyboards from provided buttons."""
    keyb = []
    for btn in buttons:
        if btn[-1] and keyb:
            keyb[-1].append(InlineKeyboardButton(btn[0], url=btn[1]))
        else:
            keyb.append([InlineKeyboardButton(btn[0], url=btn[1])])

    return keyb


SMART_OPEN = "“"
SMART_CLOSE = "”"
START_CHAR = ("'", '"', SMART_OPEN)


async def escape_invalid_curly_brackets(text: str, valids: list) -> str:
    new_text = ""
    idx = 0
    while idx < len(text):
        if text[idx] == "{":
            if idx + 1 < len(text) and text[idx + 1] == "{":
                idx += 2
                new_text += "{{{{"
                continue
            success = False
            for v in valids:
                if text[idx:].startswith("{" + v + "}"):
                    success = True
                    break
            if success:
                new_text += text[idx : idx + len(v) + 2]
                idx += len(v) + 2
                continue
            new_text += "{{"

        elif text[idx] == "}":
            if idx + 1 < len(text) and text[idx + 1] == "}":
                idx += 2
                new_text += "}}}}"
                continue
            new_text += "}}"

        else:
            new_text += text[idx]
        idx += 1

    return new_text


async def escape_mentions_using_curly_brackets(
    m: Message,
    text: str,
    parse_words: list,
):
    teks = await escape_invalid_curly_brackets(text, parse_words)
    if teks:
        teks = teks.format(
            first=escape(m.from_user.first_name),
            last=escape(m.from_user.last_name or m.from_user.first_name),
            fullname=" ".join(
                [
                    escape(m.from_user.first_name),
                    escape(m.from_user.last_name),
                ]
                if m.from_user.last_name
                else [escape(m.from_user.first_name)],
            ),
            username="@" + escape(m.from_user.username)
            if m.from_user.username
            else (await mention_html(m.from_user.first_name, m.from_user.id)),
            mention=(await mention_html(m.from_user.first_name, m.from_user.id)),
            chatname=escape(m.chat.title)
            if m.chat.type != "private"
            else escape(m.from_user.first_name),
            id=m.from_user.id,
        )
    else:
        teks = ""

    return teks


async def split_quotes(text: str):
    """Split quotes in text."""
    if not any(text.startswith(char) for char in START_CHAR):
        return text.split(None, 1)
    counter = 1  # ignore first char -> is some kind of quote
    while counter < len(text):
        if text[counter] == "\\":
            counter += 1
        elif text[counter] == text[0] or (
            text[0] == SMART_OPEN and text[counter] == SMART_CLOSE
        ):
            break
        counter += 1
    else:
        return text.split(None, 1)

    # 1 to avoid starting quote, and counter is exclusive so avoids ending
    key = await remove_escapes(text[1:counter].strip())
    # index will be in range, or `else` would have been executed and returned
    rest = text[counter + 1 :].strip()
    if not key:
        key = text[0] + text[0]
    return list(filter(None, [key, rest]))


async def remove_escapes(text: str) -> str:
    """Remove the escaped from message."""
    counter = 0
    res = ""
    is_escaped = False
    while counter < len(text):
        if is_escaped:
            res += text[counter]
            is_escaped = False
        elif text[counter] == "\\":
            is_escaped = True
        else:
            res += text[counter]
        counter += 1
    return res
