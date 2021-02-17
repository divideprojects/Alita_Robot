# We're gonna use python 3.9.1 and use Debian as OS
FROM python:3.9.1-buster

# Update and Install sudo
RUN apt update && \
    apt upgrade -y && \
    apt install sudo -y

# Add gitpod usr with sudo rights
RUN useradd -l -u 33333 -G sudo -md /home/gitpod -s /bin/bash -p gitpod gitpod \
    && sed -i.bkp -e 's/%sudo\s\+ALL=(ALL\(:ALL\)\?)\s\+ALL/%sudo ALL=NOPASSWD:ALL/g' /etc/sudoers

# Change user to gitpod
USER gitpod

# Update, Upgrade and Install Dependencies
RUN sudo apt update && \
    sudo apt upgrade -y && \
    sudo apt install -y \
        build-essential \
        bash \
        git \
        python3-dev \
        python3-lxml \
        neofetch
