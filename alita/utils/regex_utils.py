from regex import search


async def regex_searcher(regex_string, string):
    try:
        search = search(regex_string, string, timeout=6)
    except TimeoutError:
        return False
    except BaseException:
        return False
    return search


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
