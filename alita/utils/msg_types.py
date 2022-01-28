from enum import IntEnum, unique

from pyrogram.types import Message


@unique
class Types(IntEnum):
    TEXT = 1
    DOCUMENT = 2
    PHOTO = 3
    VIDEO = 4
    STICKER = 5
    AUDIO = 6
    VOICE = 7
    VIDEO_NOTE = 8
    ANIMATION = 9
    ANIMATED_STICKER = 10
    CONTACT = 11


async def get_note_type(m: Message):
    """Get type of note."""
    if len(m.text.split()) <= 1:
        return None, None, None, None

    data_type = None
    content = None
    raw_text = m.text.markdown if m.text else m.caption.markdown
    args = raw_text.split(None, 2)
    note_name = args[1]

    if len(args) >= 3:
        text = args[2]
        data_type = Types.TEXT

    elif m.reply_to_message:

        if m.reply_to_message.text:
            text = m.reply_to_message.text.markdown
        elif m.reply_to_message.caption:
            text = m.reply_to_message.caption.markdown
        else:
            text = ""

        if len(args) >= 2 and m.reply_to_message.text:  # not caption, text
            data_type = Types.TEXT

        elif m.reply_to_message.sticker:
            content = m.reply_to_message.sticker.file_id
            data_type = Types.STICKER

        elif m.reply_to_message.document:
            if m.reply_to_message.document.mime_type == "application/x-bad-tgsticker":
                data_type = Types.ANIMATED_STICKER
            else:
                data_type = Types.DOCUMENT
            content = m.reply_to_message.document.file_id

        elif m.reply_to_message.photo:
            content = m.reply_to_message.photo.file_id  # last elem = best quality
            data_type = Types.PHOTO

        elif m.reply_to_message.audio:
            content = m.reply_to_message.audio.file_id
            data_type = Types.AUDIO

        elif m.reply_to_message.voice:
            content = m.reply_to_message.voice.file_id
            data_type = Types.VOICE

        elif m.reply_to_message.video:
            content = m.reply_to_message.video.file_id
            data_type = Types.VIDEO

        elif m.reply_to_message.video_note:
            content = m.reply_to_message.video_note.file_id
            data_type = Types.VIDEO_NOTE

        elif m.reply_to_message.animation:
            content = m.reply_to_message.animation.file_id
            data_type = Types.ANIMATION

    else:
        return None, None, None, None

    return note_name, text, data_type, content


async def get_filter_type(m: Message):
    """Get filter type."""
    if len(m.text.split()) <= 1:
        return None, None, None, None

    data_type = None
    content = None
    raw_text = m.text.markdown if m.text else m.caption.markdown
    args = raw_text.split(None, 2)

    if not m.reply_to_message and m.text and len(m.text.split()) >= 3:
        content = None
        text = m.text.split(None, 2)[2]
        data_type = Types.TEXT

    elif m.reply_to_message:

        if m.reply_to_message.text:
            text = m.reply_to_message.text.markdown
        elif m.reply_to_message.caption:
            text = m.reply_to_message.caption.markdown
        else:
            text = ""

        if len(args) >= 2 and m.reply_to_message.text:  # not caption, text
            data_type = Types.TEXT

        elif m.reply_to_message.sticker:
            content = m.reply_to_message.sticker.file_id
            data_type = Types.STICKER

        elif m.reply_to_message.document:
            if m.reply_to_message.document.mime_type == "application/x-bad-tgsticker":
                data_type = Types.ANIMATED_STICKER
            else:
                data_type = Types.DOCUMENT
            content = m.reply_to_message.document.file_id

        elif m.reply_to_message.photo:
            content = m.reply_to_message.photo.file_id  # last elem = best quality
            data_type = Types.PHOTO

        elif m.reply_to_message.audio:
            content = m.reply_to_message.audio.file_id
            data_type = Types.AUDIO

        elif m.reply_to_message.voice:
            content = m.reply_to_message.voice.file_id
            data_type = Types.VOICE

        elif m.reply_to_message.video:
            content = m.reply_to_message.video.file_id
            data_type = Types.VIDEO

        elif m.reply_to_message.video_note:
            content = m.reply_to_message.video_note.file_id
            data_type = Types.VIDEO_NOTE

        elif m.reply_to_message.animation:
            content = m.reply_to_message.animation.file_id
            data_type = Types.ANIMATION

    else:
        text = None
        data_type = None
        content = None

    return text, data_type, content


async def get_wlcm_type(m: Message):
    """Get wlcm type."""
    data_type = None
    content = None
    raw_text = m.text.markdown if m.text else m.caption.markdown
    args = raw_text.split(None, 1)

    if not m.reply_to_message and m.text and len(m.text.split()) >= 2:
        content = None
        text = m.text.split(None, 1)[1]
        data_type = Types.TEXT

    elif m.reply_to_message:

        if m.reply_to_message.text:
            text = m.reply_to_message.text.markdown
        elif m.reply_to_message.caption:
            text = m.reply_to_message.caption.markdown
        else:
            text = ""

        if len(args) >= 1 and m.reply_to_message.text:  # not caption, text
            data_type = Types.TEXT

        elif m.reply_to_message.sticker:
            content = m.reply_to_message.sticker.file_id
            data_type = Types.STICKER

        elif m.reply_to_message.document:
            if m.reply_to_message.document.mime_type == "application/x-bad-tgsticker":
                data_type = Types.ANIMATED_STICKER
            else:
                data_type = Types.DOCUMENT
            content = m.reply_to_message.document.file_id

        elif m.reply_to_message.photo:
            content = m.reply_to_message.photo.file_id  # last elem = best quality
            data_type = Types.PHOTO

        elif m.reply_to_message.audio:
            content = m.reply_to_message.audio.file_id
            data_type = Types.AUDIO

        elif m.reply_to_message.voice:
            content = m.reply_to_message.voice.file_id
            data_type = Types.VOICE

        elif m.reply_to_message.video:
            content = m.reply_to_message.video.file_id
            data_type = Types.VIDEO

        elif m.reply_to_message.video_note:
            content = m.reply_to_message.video_note.file_id
            data_type = Types.VIDEO_NOTE

        elif m.reply_to_message.animation:
            content = m.reply_to_message.animation.file_id
            data_type = Types.ANIMATION

    else:
        text = None
        data_type = None
        content = None

    return text, data_type, content
