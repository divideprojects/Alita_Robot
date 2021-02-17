FROM python:3.9.1-slim-buster

# Add user gitpod
RUN apt update && \
    apt upgrade -y && \
    apt-get install sudo -y
RUN useradd -l -u 33333 -G sudo -md /home/gitpod -s /bin/bash -p gitpod gitpod \
    && sed -i.bkp -e 's/%sudo\s\+ALL=(ALL\(:ALL\)\?)\s\+ALL/%sudo ALL=NOPASSWD:ALL/g' /etc/sudoers

USER gitpod

RUN sudo apt update && \
    sudo apt upgrade -y && \
    sudo apt install -y \
        make \
        bash \
        gcc \
        git \
        python3-dev \
        python3-lxml \
        neofetch
