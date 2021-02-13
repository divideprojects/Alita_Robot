# Alita_Robot


![Forks](https://img.shields.io/github/forks/Divkix/Alita_Robot)
![Stars](https://img.shields.io/github/stars/Divkix/Alita_Robot)
![LICENSE](https://img.shields.io/github/license/Divkix/Alita_Robot)
![Repo Size](https://img.shields.io/github/repo-size/Divkix/Alita_Robot)
[![DeepSource](https://static.deepsource.io/deepsource-badge-light-mini.svg)](https://deepsource.io/gh/Divkix/Alita_Robot/?ref=repository-badge)

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/4ed13d169d5246c983bfcbfa813b6194)](https://app.codacy.com/gh/Divkix/Alita_Robot?utm_source=github.com&utm_medium=referral&utm_content=Divkix/Alita_Robot&utm_campaign=Badge_Grade_Settings)
![Views](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https://github.com/Divkix/Alita_Robot&title=Profile%20Views)
[![Crowdin](https://badges.crowdin.net/alita_robot/localized.svg)](https://crowdin.com/project/alita_robot)


Alita is a Telegram Group managment bot made using **[Pyrogram](https://docs.pyrogram.org) _async version_** and **[Python](https://python.org)**, which makes it modern and faster than most of the exisitng Telegram Chat Managers.

Help us bring more languages to the bot by contributing to the project in [Crowdin](https://crowdin.com/project/alitarobot)!

**Alita's features over other bots:**
-   Modern
-   Fast
-   Fully asynchronous
-   Fully open-source
-   Frequently updated
-   Multi Language Support

Can be found on Telegram as [@AlitaRobot](https://t.me/AlitaRobot).

Alita is currently available in 5 Languages as of now: **en-US**, **pt-BR**, **it-IT**, **ru-RU**, **hi-IN**.
More languages can be managed in the _locales_ folder.

You need to have a *Postgres Database*, and *Redis Cache Database* as well!

## How  to setup:

First Step!
- Star **⭐** the repository!!

It really motivates me to continue this project further.

### Traditional:
- Install Python v3.7 or later from [Python's Website](https://python.org)
- Install virtualenv using `python3 -m pip -U install virtualenv`.
- **Fork** or Clone the project using `git clone https://github.com/Divkix/Alita_Robot.git`
- Install the requirements using `python3 -m pip install -r requirements.txt`
- Rename `sample_config.py` to `config.py` in `alita` folder and fill in all the variables in *Development* class, not *Config* class. **Sudo, Dev, Whitelist** users are optional!!
- Run the bot using `python3 -m alita`
If successful, bot should send message to the **MESSAGE_DUMP** Group!

## TO-DO
- [ ] Fix Errors, by defining them
- [ ] Proper Translations
- [ ] Add Captcha
- [ ] Add federations
- [ ] Add Sticker Blacklist
- [ ] Add Greetings (Welcome and Goodbye)
- [ ] Add anti-flood
- [ ] Add backup
- [ ] Add Logging of groups and channels
- [ ] Add warnings
- [ ] Add connections
- [x] Fix Docker Configuration (Need to enter ENV Vars Manually)

## Contributing to the project

- You must sign off on your commit.
- You must sign the commit via GPG Key.
- Make sure your PR works and doesn't break anything!

## Special Thanks to:
- [AmanoTeam](https://github.com/AmanoTeam/) for [EduuRobot](https://github.com/AmanoTeam/EduuRobot/tree/rewrite) as that helped me make the language menu with the 4 langauges provided and some basic plugins too!
- [Dan](https://github.com/delivrance) for his [Pyrogram](https://github.com/pyrogram) library
- [Paul Larsen](https://github.com/PaulSonOfLars) for his Original Marie Source Code.
- Everyone else who inspired me to make this project, more names can be seen on commits!

### Copyright & License

* Copyright (C) 2020 by [Divkix](https://github.com/Divkix) ❤️️
* Licensed under the terms of the [GNU AFFERO GENERAL PUBLIC LICENSE Version 3, 29 June 2007](https://github.com/Divkix/Alita_Robot/blob/master/LICENSE)
