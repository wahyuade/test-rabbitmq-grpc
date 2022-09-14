FROM python:alpine3.16 as db_migrator

RUN apk add --no-cache postgresql-libs python3-dev postgresql-dev gcc musl-dev g++
RUN pip install alembic
RUN pip install psycopg2
RUN pip install python-dotenv

WORKDIR /migration/
COPY alembic.ini migration.sh ./
COPY database ./database
RUN chmod +x ./migration.sh && ./migration.sh

FROM golang:1.17-alpine as builder
WORKDIR /app/
COPY . ./

RUN go mod tidy
RUN go build -o cli

FROM alpine:3.14 
WORKDIR /root/
COPY --from=builder /app/cli ./
COPY .env.docker ./.env

CMD ["./cli"]
