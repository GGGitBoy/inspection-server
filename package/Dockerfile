FROM ubuntu:20.04

# 更新包列表并安装必要的软件包
RUN apt-get update \
    && apt-get install -y \
        tzdata \
        chromium-browser \
        fonts-noto-cjk \
    && ln -sf /usr/share/zoneinfo/$TZ /etc/localtime \
    && echo $TZ > /etc/timezone \
    && apt-get clean

# 设置时区和环境变量
ENV TZ=Asia/Shanghai

# 验证时区设置是否成功
RUN ls /usr/share/zoneinfo/Asia/Shanghai

# 将应用程序和测试文件复制到镜像中
COPY bin/inspection-server /inspection-server
COPY package/test.pdf /opt/test.pdf
COPY agent/yaml /opt/yaml

# 指定容器启动时运行的命令
CMD ["/inspection-server"]
