FROM alpine:latest

RUN apk update && apk add --no-cache tzdata

RUN apk add chromium

RUN apk add --no-cache font-noto-cjk

ENV TZ=Asia/Shanghai

RUN ls /usr/share/zoneinfo/Asia/Shanghai

COPY bin/inspection-server /inspection-server

CMD ["/inspection-server"]
