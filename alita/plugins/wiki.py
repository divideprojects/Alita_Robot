import wikipedia
from wikipedia.exceptions import DisambiguationError, PageError
from alita import PREFIX_HANDLER
from alita.bot_class import Alita
from pyrogram import filters
from pyrogram.types import Message

__PLUGIN__ = "Wikipedia"

__help__ = """
Search Wikipedia on the go in your group!

**Available commands:**
 • /wiki <query>: wiki your query.
"""


@Alita.on_message(filters.command("wiki", PREFIX_HANDLER))
async def wiki(c: Alita, m: Message):
    if m.reply_to_message:
        search = m.reply_to_message.text
    else:
        search = m.text.split(None, 1)[1]
    try:
        res = wikipedia.summary(search)
    except DisambiguationError as de:
        await m.reply_text(
            f"Disambiguated pages found! Adjust your query accordingly.\n<i>{de}</i>",
            parse_mode="html",
        )
        return
    except PageError as pe:
        await m.reply_text(f"<code>{pe}</code>", parse_mode="html")
        return
    if res:
        result = f"<b>{search}</b>\n\n"
        result += f"<i>{res}</i>\n"
        result += f"""<a href="https://en.wikipedia.org/wiki/{search.replace(" ", "%20")}">Read more...</a>"""
        if len(result) > 4000:
            with open("result.txt", "rb") as f:
                await c.send_document(
                    document=f,
                    reply_to_message_id=m.message_id,
                    chat_id=m.chat.id,
                    parse_mode="html",
                )
        else:
            await m.reply_text(result, parse_mode="html", disable_web_page_preview=True)
    return
