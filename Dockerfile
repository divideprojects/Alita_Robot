FROM python:3.9.1-slim-buster

# Don't use cached python packages
ENV PIP_NO_CACHE_DIR 1

# Add new fast source
RUN sed -i.bak 's/us-west-2\.ec2\.//' /etc/apt/sources.list

# Installing Required Packages
RUN apt update && apt upgrade -y && \
    apt install --no-install-recommends -y \
    bash \
    python3-dev \
    python3-lxml \
    gcc \
    make

# Clean a bit
RUN rm -rf /var/lib/apt/lists /var/cache/apt/archives /tmp

# Enter Workplace
WORKDIR /app/

# Copy folder
COPY . .

# Install dependencies
RUN make install

# Run the bots
CMD ["make", "run"]
