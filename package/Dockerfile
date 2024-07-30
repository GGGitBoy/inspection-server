FROM alpine:latest

# 设置时区并安装必要的软件包
RUN apk update \
    && apk add --no-cache \
        tzdata \        # 安装时区数据
        chromium \      # 安装 Chromium 浏览器
        font-noto-cjk   # 安装 Noto CJK 字体

# 设置时区为 Asia/Shanghai
ENV TZ=Asia/Shanghai

# 验证时区设置是否成功
RUN ls /usr/share/zoneinfo/Asia/Shanghai

# 将应用程序和测试文件复制到镜像中
COPY bin/inspection-server /inspection-server

COPY test.pdf /opt/test.pdf

# 指定容器启动时运行的命令
CMD ["/inspection-server"]
