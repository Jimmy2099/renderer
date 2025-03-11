FROM ubuntu:25.04

WORKDIR /

RUN apt update
RUN apt install git curl wget
#mount your git gitconfig first
RUN --mount=type=secret,id=gitconfig,target=/root/.gitconfig cat /root/.gitconfig

RUN git clone https://github.com/Jimmy2099/renderer.git

WORKDIR /renderer

RUN git submodule update --init --recursive
