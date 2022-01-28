from html import escape
from re import compile as compilere
from re import sub


async def cleanhtml(raw_html: str) -> str:
    """Clean html data."""
    cleanr = compilere("<.*?>")
    return sub(cleanr, "", raw_html)


async def escape_markdown(text: str) -> str:
    """Escape markdown data."""
    escape_chars = r"\*_`\["
    return sub(r"([%s])" % escape_chars, r"\\\1", text)


async def mention_html(name: str, user_id: int) -> str:
    """Mention user in html format."""
    name = escape(name)
    return f'<a href="tg://user?id={user_id}">{name}</a>'


async def mention_markdown(name: str, user_id: int) -> str:
    """Mention user in markdown format."""
    return f"[{(await escape_markdown(name))}](tg://user?id={user_id})"
