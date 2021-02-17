# Alita_Robot


![Forks](https://img.shields.io/github/forks/Divkix/Alita_Robot)
![Stars](https://img.shields.io/github/stars/Divkix/Alita_Robot)
![Issues](https://img.shields.io/github/issues/Divkix/Alita_Robot)
![LICENSE](https://img.shields.io/github/license/Divkix/Alita_Robot)
![Contributors](https://img.shields.io/github/contributors/Divkix/Alita_Robot)
![Repo Size](https://img.shields.io/github/repo-size/Divkix/Alita_Robot)
![Views](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https://github.com/Divkix/Alita_Robot&title=Profile%20Views)

[![DeepSource](https://static.deepsource.io/deepsource-badge-light-mini.svg)](https://deepsource.io/gh/Divkix/Alita_Robot/?ref=repository-badge)
[![Build Status](https://travis-ci.com/Divkix/Alita_Robot.svg?branch=beta)](https://travis-ci.com/Divkix/Alita_Robot)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/4ed13d169d5246c983bfcbfa813b6194)](https://app.codacy.com/gh/Divkix/Alita_Robot?utm_source=github.com&utm_medium=referral&utm_content=Divkix/Alita_Robot&utm_campaign=Badge_Grade_Settings)
[![Crowdin](https://badges.crowdin.net/alita_robot/localized.svg)](https://crowdin.com/project/alita_robot)

[![Gitpod Ready-to-Code](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/Divkix/Alita_Robot/tree/beta)

Alita is a Telegram Group managment bot made using **[Pyrogram](https://github.com/pyrogram/pyrogram) _async version_** and **[Python](https://python.org)**, which makes it modern and faster than most of the exisitng Telegram Chat Managers.

**Alita's features over other bots:**
- Modern
- Fast
- Fully asynchronous
- Fully open-source
- Frequently updated
- Multi Language Support

Can be found on Telegram as [@Alita_Robot](https://t.me/Alita_Robot).

Alita is currently available in 6 Languages as of now:
- **en-US**
- **pt-BR**
- **it-IT**
- **ru-RU**
- **hi-IN**
- **he-IL**

More languages can be managed in the _locales_ folder.

We are still working on adding new languages.

Help us bring more languages to the bot by contributing to the project in [Crowdin](https://crowdin.com/project/alitarobot)

## Requirements
- You need to have a *Postgres Database*
- You also need *Redis Cache Database*
- Linux machine (Ubuntu/Denain-based OS Preferred)


## How to setup

First Step!
- Star **⭐** the repository!!

It really motivates me to continue this project further.


### Traditional
- Install Python v3.7 or later from [Python's Website](https://python.org)
- Install virtualenv using `python3 -m pip -U install virtualenv`.
- **Fork** or Clone the project using `git clone https://github.com/Divkix/Alita_Robot.git`
- Install the requirements using `python3 -m pip install -r requirements.txt`
- Fill in all the variables in *Development* class, not *Config* class. **Sudo, Dev, Whitelist** users are optional!!
- Run the bot using `python3 -m alita`

### Docker
- Clone the repo and enter into it
- Install [Docker](https://www.docker.com/).
- Fill in the `sample.env` file and rename it to `main.env`.
- Build the docker image using: `docker build -t alita_robot:latest .`
- Run the command `docker run --env-file main.env alita_robot`


If all works well, bot should send message to the **MESSAGE_DUMP** Group!


## TO-DO
- [ ] Fix Errors (by defining them in `except Exception`)
- [ ] Fix translations (Some are still in English)
- [ ] Add Captcha (To check user when they enter group)
- [ ] Add Federations
- [ ] Add Sticker Blacklist
- [ ] Add Greetings (Welcome and Goodbye)
- [ ] Add Anti-flood
- [x] Full Asynchronous (All functions run async)
- [ ] Add backup/restore option (Chat settings can be backud up)
- [ ] Add Warnings
- [ ] Add Connections (Connect group chats to PM)
- [x] Fix Docker Configuration
- [ ] Switch to MongoDB

*Still need to add docker-compose to connect with Redis and Db using a single command!


## Contributing to the project

- Make sure your PR works and doesn't break anything.
- You must join the support group.
- Make sure it passes test using `make test`.


## Special Thanks to
- [AmanoTeam](https://github.com/AmanoTeam/) for [EduuRobot](https://github.com/AmanoTeam/EduuRobot/tree/rewrite) as that helped me make the language menu with the 4 langauges provided and some basic plugins.
- [Dan](https://github.com/delivrance) for his [Pyrogram](https://github.com/pyrogram) library
- [Paul Larsen](https://github.com/PaulSonOfLars) for his Original Marie Source Code.
- Everyone else who inspired me to make this project, more names can be seen on commits!


### Copyright & License

* Copyright (C) 2020-2021 by [Divkix](https://github.com/Divkix) ❤️️
* Licensed under the terms of the [GNU AFFERO GENERAL PUBLIC LICENSE Version 3, 29 June 2007](https://github.com/Divkix/Alita_Robot/blob/master/LICENSE)
