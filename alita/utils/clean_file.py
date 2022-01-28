async def remove_markdown_and_html(text: str) -> str:
    return await clean_markdown(await clean_html(text))


async def clean_html(text: str) -> str:
    return (
        text.replace("<code>", "")
        .replace("</code>", "")
        .replace("<b>", "")
        .replace("</b>", "")
        .replace("<i>", "")
        .replace("</i>", "")
        .replace("<u>", "")
        .replace("</u>", "")
    )


async def clean_markdown(text: str) -> str:
    return text.replace("`", "").replace("**", "").replace("__", "")
