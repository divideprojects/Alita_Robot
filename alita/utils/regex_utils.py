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


from regex import search


async def regex_searcher(regex_string, string):
    try:
        re_search = search(regex_string, string, timeout=6)
    except TimeoutError:
        return False
    except BaseException:
        return False

    return re_search


async def infinite_loop_check(regex_string):
    loop_matches = (
        r"\((.{1,}[\+\*]){1,}\)[\+\*]."
        r"[\(\[].{1,}\{\d(,)?\}[\)\]]\{\d(,)?\}"
        r"\(.{1,}\)\{.{1,}(,)?\}\(.*\)(\+|\* |\{.*\})"
    )

    for match in loop_matches:
        match_1 = search(match, regex_string)

    if match_1:
        return True

    return False
