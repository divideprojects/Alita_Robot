def remove_markdown_and_html(text):
    clean_html = (
        text.replace("<code>", "")
        .replace("</code>", "")
        .replace("<b>", "")
        .replace("</b>", "")
        .replace("<i>", "")
        .replace("</i>", "")
        .replace("<u>", "")
        .replace("</u>", "")
    )
    clean_text = clean_html.replace("`", "").replace("**", "").replace("__", "")
    return clean_text
