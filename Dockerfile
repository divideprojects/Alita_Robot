FROM python:3.9.7-slim-bullseye

# Don't use cached python packages
ENV PIP_NO_CACHE_DIR 1

# Installing Required Packages
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install --no-install-recommends -y \
    bash \
    python3-dev \
    python3-lxml \
    gcc \
    git \
    make \
    && rm -rf /var/lib/apt/lists /var/cache/apt/archives /tmp

# Enter Workplace
WORKDIR /app

# Copy folder
COPY . .

# Install dependencies
RUN pip3 install --upgrade pip

# Install Bot Deps and stuff
RUN make install

# Run the bot
ENTRYPOINT ["make", "run"]
