from html import escape
from re import sub, compile as compilere


async def cleanhtml(raw_html):
    cleanr = compilere("<.*?>")
    cleantext = sub(cleanr, "", raw_html)
    return cleantext


async def escape_markdown(text):
    escape_chars = r"\*_`\["
    return sub(r"([%s])" % escape_chars, r"\\\1", text)


async def mention_html(name, user_id):
    name = escape(name)
    return f'<a href="tg://user?id={user_id}">{name}</a>'


async def mention_markdown(name, user_id):
    return f"[{(await escape_markdown(name))}](tg://user?id={user_id})"
