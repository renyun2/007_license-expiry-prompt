# syntax=docker/dockerfile:1

FROM node:20-alpine AS frontend
WORKDIR /src/web
COPY web/package.json ./
RUN npm install
COPY web/ ./
RUN npm run build

FROM golang:1.22-alpine AS backend
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
COPY --from=frontend /src/web/dist ./web/dist
RUN go mod tidy && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /server .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
ENV SQLITE_PATH=/app/data/certs.db
ENV INIT_SQL_PATH=/app/init.sql
COPY --from=backend /server ./server
COPY init.sql /app/init.sql
RUN mkdir -p /app/data
EXPOSE 8080
CMD ["./server"]
