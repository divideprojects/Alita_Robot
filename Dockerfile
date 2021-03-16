FROM bitnami/python:3.9.2-prod

# Don't use cached python packages
ENV PIP_NO_CACHE_DIR 1

# Installing Required Packages
RUN apt update && \
    apt upgrade -y && \
    apt install --no-install-recommends -y \
    bash \
    python3-dev \
    python3-lxml \
    gcc \
    git \
    make \
    neofetch

# Clear apt lists
RUN rm -rf /var/lib/apt/lists/*

# Enter Workplace
WORKDIR /app/

# Copy folder
RUN git clone https://github.com/Divkix/Alita_Robot.git .

# Install dependencies
RUN make docker

# Run the bot
CMD ["make", "run"]
