# 基础镜像
FROM alpine:latest

RUN apk add --no-cache \
    curl \
    bash \
    jq \
    docker \
    bash \
    && curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" \
    && chmod +x kubectl \
    && mv kubectl /usr/local/bin/

#RUN mkdir /host
#WORKDIR /host
#PATH=$PATH:/inspection/usr/local/sbin:/inspection/usr/local/bin:/inspection/usr/sbin:/inspection/usr/bin:/inspection/sbin:/inspection/bin


#FROM docker \
#

#USER root

# 启动 Docker in Docker
#ENTRYPOINT ["dockerd-entrypoint.sh"]

# 暴露 Docker 端口
#EXPOSE 2375



# 例如，下载 crictl v1.23.0 版本
#wget https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.23.0/crictl-v1.23.0-linux-amd64.tar.gz
#tar -zxvf crictl-v1.23.0-linux-amd64.tar.gz
#mv crictl /usr/local/bin/
#
#
## 例如，下载 ctr v1.5.2 版本
#wget https://github.com/containerd/cri/releases/download/v1.5.2/ctr-1.5.2-linux-amd64.tar.gz
#tar -zxvf ctr-1.5.2-linux-amd64.tar.gz
#mv ctr /usr/local/bin/

