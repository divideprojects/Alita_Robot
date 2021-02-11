import re


def cleanhtml(raw_html):
    cleanr = re.compile("<.*?>")
    cleantext = re.sub(cleanr, "", raw_html)
    return cleantext


def escape_markdown(text):
    escape_chars = r"\*_`\["
    return re.sub(r"([%s])" % escape_chars, r"\\\1", text)


def mention_html(name, user_id):
    # name = html.escape(name)
    return u'<a href="tg://user?id={}">{}</a>'.format(user_id, name)


async def mention_markdown(name, user_id):
    return u"[{}](tg://user?id={})".format(escape_markdown(name), user_id)
