FROM python:3.9.1-slim-buster

USER root

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
