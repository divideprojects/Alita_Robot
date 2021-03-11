# Alita_Robot

<p align='center'>
  <img src="https://img.shields.io/github/forks/Divkix/Alita_Robot?style=flat-square" alt="Forks">
  <img src="https://img.shields.io/github/stars/Divkix/Alita_Robot?style=flat-square" alt="Stars">
  <img src="https://img.shields.io/github/issues/Divkix/Alita_Robot?style=flat-square" alt="Issues">
  <img src="https://img.shields.io/github/license/Divkix/Alita_Robot?style=flat-square" alt="LICENSE">
  <img src="https://img.shields.io/github/contributors/Divkix/Alita_Robot?style=flat-square" alt="Contributors">
  <img src="https://img.shields.io/github/repo-size/Divkix/Alita_Robot?style=flat-square" alt="Repo Size">
  <img src="https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https://github.com/Divkix/Alita_Robot&amp;title=Profile%20Views" alt="Views">
  <img src="https://img.shields.io/librariesio/github/Divkix/Alita_Robot?style=flat-square" alt="Dependencies">
</p>

<p align='center'>
  <a href="https://www.python.org/" alt="made-with-python"> <img src="https://img.shields.io/badge/Made%20with-Python-1f425f.svg?style=flat-square&logo=python&color=blue" /> </a>
  <a href="https://github.com/Divkix/Alita_Robot" alt="Docker!"> <img src="https://aleen42.github.io/badges/src/docker.svg" /> </a>
  <a href="https://github.com/Divkix/Alita_Robot/graphs/commit-activity" alt="Maintenance"> <img src="https://img.shields.io/badge/Maintained%3F-yes-green.svg?style=flat-square" /> </a>
</p>

<p align='center'>
  <img src="https://img.shields.io/github/commit-activity/m/divkix/alita_robot?style=flat-square" />
</p>

<p align='center'>
  <a href="https://deepsource.io/gh/Divkix/Alita_Robot/?ref=repository-badge"><img src="https://static.deepsource.io/deepsource-badge-light-mini.svg" alt="DeepSource"></a>
  <a href="https://travis-ci.com/Divkix/Alita_Robot"><img src="https://img.shields.io/travis/com/Divkix/Alita_Robot/main?style=flat-square&logo=travis" alt="Build Status"></a>
  <a href="https://app.codacy.com/gh/Divkix/Alita_Robot?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=Divkix/Alita_Robot&amp;utm_campaign=Badge_Grade_Settings"><img src="https://api.codacy.com/project/badge/Grade/4ed13d169d5246c983bfcbfa813b6194" alt="Codacy Badge"></a>
  <a href="https://makeapullrequest.com" alt="PRs Welcome"> <img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square" /> </a>
  <a href="https://crowdin.com/project/alita_robot"><img src="https://badges.crowdin.net/alita_robot/localized.svg" alt="Crowdin"></a>
</p>

<p align='center'>
  <a href="https://gitpod.io/#https://github.com/Divkix/Alita_Robot/tree/main"><img src="https://gitpod.io/button/open-in-gitpod.svg" alt="Gitpod Ready-to-Code"></a></br></br>
  <a href="https://t.me/DivideProjects"><img src="https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&amp;logo=telegram&amp;logoColor=white" alt="Join us on Telegram"></a>
</p>


Alita is a Telegram Group managment bot made using **[Pyrogram](https://github.com/pyrogram/pyrogram) _async version_** and **[Python](https://python.org)**, which makes it modern and faster than most of the exisitng Telegram Chat Managers.

**Alita's features over other bots:**
- Modern
- Fast
- Fully asynchronous
- Fully open-source
- Frequently updated
- Multi Language Support

Can be found on Telegram as [@Alita_Robot](https://t.me/Alita_Robot)</br>
This runs the stable version, aka [production](https://github.com/Divkix/Alita_Robot/tree/production) branch.

Beta Version: [@AlitaBetaBot](https://t.me/AlitaBetaBot)</br>
This runs on beta version, which is from [main](https://github.com/Divkix/Alita_Robot/tree/main) branch.

Alita is currently available in 4 Languages as of now:
- **US English**
- **Spanish**
- **French**
- **German**

More languages can be managed in the _locales_ folder.

We are still working on adding new languages.

Help us bring more languages to the bot by contributing to the project on [Crowdin](https://crowdin.com/project/alitarobot)

## Requirements
- You need to have a *Mongo Database* (Cluster Preferred)
- Linux machine (Ubuntu/Denain-based OS Preferred)


## How to setup

First Step!
- Star **⭐** the repository!!

It really motivates me to continue this project further.

### Deploy to Heroku
- Get your `API_ID` and `API_HASH` from [here](https://my.telegram.org/)
- Get your Bot Token from [@BotFather](https://t.me/BotFather)
and click the below button!  <br />

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/Divkix/Alita_Robot)

### Traditional

- Install Python v3.7 or later from [Python's Website](https://python.org)
- Install virtualenv using `python3 -m pip -U install virtualenv`.
- **Fork** or Clone the project using `git clone https://github.com/Divkix/Alita_Robot.git`
- Install the requirements using `python3 -m pip install -r requirements.txt`
- Fill in all the variables in *Development* class, not *Config* class. **Sudo, Dev, Whitelist** users are optional!!
- Run the bot using `python3 -m alita`

### Docker

- Clone the repo and enter into it
- Install [Docker](https://www.docker.com/)
- Fill in the `sample.env` file and rename it to `main.env`.
- Build the docker image using: `docker build -t alita_robot:latest .` (The dot '.' at last is necessary!)
- Run the command `docker run --env-file main.env alita_robot`


If all works well, bot should send message to the **MESSAGE_DUMP** Group!


## Contributing to the project

- Make sure your PR works and doesn't break anything.
- You must join the support group.
- Make sure it passes test using `make test`.


## Special Thanks to
- [AmanoTeam](https://github.com/AmanoTeam/) for [EduuRobot](https://github.com/AmanoTeam/EduuRobot/tree/rewrite) as that helped me make the translation engine.
- [Dan](https://github.com/delivrance) for his [Pyrogram](https://github.com/pyrogram/pyrogram) library
- [Paul Larsen](https://github.com/PaulSonOfLars) for his Original Marie Source Code.
- Everyone else who inspired me to make this project, more names can be seen on commits!


### Copyright & License

* Copyright (C) 2020-2021 by [Divkix](https://github.com/Divkix) ❤️️
* Licensed under the terms of the [GNU AFFERO GENERAL PUBLIC LICENSE Version 3, 29 June 2007](https://github.com/Divkix/Alita_Robot/blob/master/LICENSE)
