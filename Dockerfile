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
RUN rm -rf /var/lib/apt/lists /var/cache/apt/archives /tmp

# Enter Workplace
WORKDIR /app/

# Copy folder
COPY . .

# Install dependencies
RUN pip3 install --upgrade pip
RUN rm -r /opt/bitnami/python/lib/python3.9/site-packages/setuptools*
RUN pip3 install --upgrade setuptools
RUN make install

# Run the bot
CMD ["make", "run"]
