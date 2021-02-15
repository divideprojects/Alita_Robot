# Sorts adminlist by second tuple place in item in it
async def sort_adminlist(adminlist):
    lst_len = len(lst)
    for i in range(0, lst_len):
        for j in range(0, lst_len - i - 1):
            if lst[j][1] > lst[j + 1][1]:
                temp = lst[j]
                lst[j] = lst[j + 1]
                lst[j + 1] = temp
    return lst
