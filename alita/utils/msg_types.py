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


from enum import IntEnum, unique


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


async def get_message_type(m):
    """Get type of message."""
    if m.text or m.caption:
        content = None
        message_type = Types.TEXT
    elif m.sticker:
        content = m.sticker.file_id
        message_type = Types.STICKER

    elif m.document:
        if m.document.mime_type == "application/x-bad-tgsticker":
            message_type = Types.ANIMATED_STICKER
        else:
            message_type = Types.DOCUMENT
        content = m.document.file_id

    elif m.photo:
        content = m.photo.file_id  # last elem = best quality
        message_type = Types.PHOTO

    elif m.audio:
        content = m.audio.file_id
        message_type = Types.AUDIO

    elif m.voice:
        content = m.voice.file_id
        message_type = Types.VOICE

    elif m.video:
        content = m.video.file_id
        message_type = Types.VIDEO

    elif m.video_note:
        content = m.video_note.file_id
        message_type = Types.VIDEO_NOTE

    elif m.animation:
        content = m.animation.file_id
        message_type = Types.ANIMATION

    # TODO
    # elif m.contact:
    # 	content = m.contact.phone_number
    # 	# text = None
    # 	message_type = Types.CONTACT

    # TODO
    # elif m.animated_sticker:
    # 	content = m.animation.file_id
    # 	text = None
    # 	message_type = Types.ANIMATED_STICKER

    else:
        return None, None

    return content, message_type


async def get_note_type(m):
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


async def get_welcome_type(m):
    """Get type of welcome."""
    data_type = None
    content = None

    if m.reply_to_message:
        if m.reply_to_message.text:
            text = m.reply_to_message.text.markdown
        elif m.reply_to_message.caption:
            text = m.reply_to_message.caption.markdown
        else:
            text = None
    else:
        text = m.text.split(None, 1)

    if m.reply_to_message:
        if m.reply_to_message.text:
            text = m.reply_to_message.text.markdown
            data_type = Types.TEXT

        elif m.reply_to_message.sticker:
            if m.reply_to_message.document.mime_type == "application/x-tgsticker":
                data_type = Types.ANIMATED_STICKER
            else:
                data_type = Types.STICKER
            content = m.reply_to_message.sticker.file_id
            text = None

        elif m.reply_to_message.document:
            if m.reply_to_message.document.mime_type == "application/x-bad-tgsticker":
                data_type = Types.ANIMATED_STICKER
            else:
                data_type = Types.DOCUMENT
            content = m.reply_to_message.document.file_id
        # text = m.reply_to_message.caption

        elif m.reply_to_message.photo:
            content = m.reply_to_message.photo[-1].file_id
            # text = m.reply_to_message.caption
            data_type = Types.PHOTO

        elif m.reply_to_message.audio:
            content = m.reply_to_message.audio.file_id
            # text = m.reply_to_message.caption
            data_type = Types.AUDIO

        elif m.reply_to_message.voice:
            content = m.reply_to_message.voice.file_id
            text = None
            data_type = Types.VOICE

        elif m.reply_to_message.video:
            content = m.reply_to_message.video.file_id
            # text = m.reply_to_message.caption
            data_type = Types.VIDEO

        elif m.reply_to_message.video_note:
            content = m.reply_to_message.video_note.file_id
            text = None
            data_type = Types.VIDEO_NOTE

        elif m.reply_to_message.animation:
            content = m.reply_to_message.animation.file_id
            # text = None
            data_type = Types.ANIMATION

    else:
        if m.caption:
            text = m.caption.split(None, 1)
            if len(text) >= 2:
                text = m.caption.markdown.split(None, 1)[1]
        elif m.text:
            text = m.text.split(None, 1)
            if len(text) >= 2:
                text = m.text.markdown.split(None, 1)[1]
        data_type = Types.TEXT

    return text, data_type, content
