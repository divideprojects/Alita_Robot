# Sorts adminlist by second tuple place in item in it
async def sort_adminlist(adminlist):
    lst_len = len(adminlist)
    for i in range(0, lst_len):
        for j in range(0, lst_len - i - 1):
            if adminlist[j][1] > adminlist[j + 1][1]:
                temp = adminlist[j]
                adminlist[j] = adminlist[j + 1]
                adminlist[j + 1] = temp
    return adminlist
