# syntax=docker/dockerfile:1
FROM golang:1.23-alpine

WORKDIR /app
COPY . /app
