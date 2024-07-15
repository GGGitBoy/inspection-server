# 基础镜像
FROM docker

# 安装 kubectl
RUN apk add --no-cache \
    curl \
    bash \
    jq \
    && curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" \
    && chmod +x kubectl \
    && mv kubectl /usr/local/bin/

USER root

# 启动 Docker in Docker
ENTRYPOINT ["dockerd-entrypoint.sh"]

# 暴露 Docker 端口
EXPOSE 2375