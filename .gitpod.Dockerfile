FROM gitpod/workspace-full:latest

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

RUN make dev-install
