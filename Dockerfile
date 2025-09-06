FROM golang:1.24 AS builder
WORKDIR /usr/src/app

COPY ./Makefile .
COPY ./go.mod .
COPY ./go.sum .
COPY ./cmd cmd
COPY ./internal internal

RUN make build


FROM ubuntu:jammy AS base

RUN apt update -y  &&  \
    apt install net-tools curl -y && \
    apt-get clean

RUN mkdir -p /opt/cryptoswap/config
RUN useradd cryptoswap
RUN chown -R cryptoswap:cryptoswap /opt/cryptoswap

COPY --from=builder /usr/src/app/build/cryptoswap /opt/cryptoswap/cryptoswap
WORKDIR /opt/cryptoswap
USER cryptoswap
