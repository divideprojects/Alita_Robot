import re
import time

from pyrogram.types import InlineKeyboardButton

BTN_URL_REGEX = re.compile(r"(\[([^\[]+?)\]\(buttonurl:(?:/{0,2})(.+?)(:same)?\))")


def replace_text(text):
    return text.replace('"', "").replace("\\r", "").replace("\\n", "").replace("\\", "")


async def extract_time(m, time_val):
    if any(time_val.endswith(unit) for unit in ("m", "h", "d")):
        unit = time_val[-1]
        time_num = time_val[:-1]  # type: str
        if not time_num.isdigit():
            await m.reply("Unspecified amount of time.")
            return ""

        if unit == "m":
            bantime = int(time.time() + int(time_num) * 60)
        elif unit == "h":
            bantime = int(time.time() + int(time_num) * 60 * 60)
        elif unit == "s":
            bantime = int(time.time() + int(time_num) * 24 * 60 * 60)
        else:
            # how even...?
            return ""
        return bantime
    await m.reply(
        "Invalid time type specified. Needed m, h, or s. got: {}".format(
            time_val[-1]
        )
    )
    return ""


async def extract_time_str(m, time_val):
    if any(time_val.endswith(unit) for unit in ("m", "h", "d")):
        unit = time_val[-1]
        time_num = time_val[:-1]  # type: str
        if not time_num.isdigit():
            await m.reply("Unspecified amount of time.")
            return ""

        if unit == "m":
            bantime = int(int(time_num) * 60)
        elif unit == "h":
            bantime = int(int(time_num) * 60 * 60)
        elif unit == "s":
            bantime = int(int(time_num) * 24 * 60 * 60)
        else:
            # how even...?
            return ""
        return bantime
    await m.reply(
        "Invalid time type specified. Needed m, h, or s. got: {}".format(
            time_val[-1]
        )
    )
    return ""


def make_time(time_val):
    if int(time_val) == 0:
        return "0"
    if int(time_val) <= 3600:
        bantime = str(int(time_val / 60)) + "m"
    elif int(time_val) >= 3600 and time_val <= 86400:
        bantime = str(int(time_val / 60 / 60)) + "h"
    elif int(time_val) >= 86400:
        bantime = str(int(time_val / 24 / 60 / 60)) + "d"
    return bantime


def id_from_reply(m):
    prev_message = m.reply_to_message
    if not prev_message:
        return None, None
    user_id = prev_message.from_user.id
    res = m.text.split(None, 1)
    if len(res) < 2:
        return user_id, ""
    return user_id, res[1]


def parse_button(text):
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


def build_keyboard(buttons):
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


def split_quotes(text: str):
    if any(text.startswith(char) for char in START_CHAR):
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
        key = remove_escapes(text[1:counter].strip())
        # index will be in range, or `else` would have been executed and returned
        rest = text[counter + 1 :].strip()
        if not key:
            key = text[0] + text[0]
        return list(filter(None, [key, rest]))
    return text.split(None, 1)


def extract_text(m):
    return m.text or m.caption or (m.sticker.emoji if m.sticker else None)


def remove_escapes(text: str) -> str:
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
