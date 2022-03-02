# This Dockerfile builds our base image with gosu, dumb-init and the atlantis
# user. We split this from the main Dockerfile because this base doesn't change
# and also because it kept breaking the build due to flakiness.
FROM alpine:3.15
LABEL authors="Anubhav Mishra, Luke Kysow"

# We use gosu to step down from root and run as the atlantis user so we need
# to create that user and group.
# We add the atlantis user to the root group and make its home directory
# owned by root so that OpenShift users can use /home/atlantis as their
# data dir because OpenShift runs containers as a random uid that's part of
# the root group.
RUN addgroup atlantis && \
    adduser -S -G atlantis atlantis && \
    adduser atlantis root && \
    chown atlantis:root /home/atlantis/ && \
    chmod g=u /home/atlantis/ && \
    chmod g=u /etc/passwd

# Install gosu and git-lfs.
ENV GOSU_VERSION=1.14
ENV GIT_LFS_VERSION=3.1.2

# Automatically populated with the architecture the image is being built for.
ARG TARGETPLATFORM

# Install packages needed for running Atlantis.
RUN apk add --no-cache ca-certificates curl git unzip bash openssh libcap dumb-init && \
    # Install packages needed for building dependencies.
    apk add --no-cache --virtual .build-deps gnupg openssl && \
    mkdir -p /tmp/build && \
    cd /tmp/build && \

    # git-lfs
    case ${TARGETPLATFORM} in \
        "linux/amd64") GIT_LFS_ARCH=amd64 ;; \
        "linux/arm64") GIT_LFS_ARCH=arm64 ;; \
        "linux/arm/v7") GIT_LFS_ARCH=arm ;; \
    esac && \
    curl -L -s --output git-lfs.tar.gz "https://github.com/git-lfs/git-lfs/releases/download/v${GIT_LFS_VERSION}/git-lfs-linux-${GIT_LFS_ARCH}-v${GIT_LFS_VERSION}.tar.gz" && \
    tar -xf git-lfs.tar.gz && \
    chmod +x git-lfs && \
    mv git-lfs /usr/bin/git-lfs && \
    git-lfs --version && \

    # gosu
    case ${TARGETPLATFORM} in \
        "linux/amd64") GOSU_ARCH=amd64 ;; \
        "linux/arm64") GOSU_ARCH=arm64 ;; \
        "linux/arm/v7") GOSU_ARCH=armhf ;; \
    esac && \
    curl -L -s --output gosu "https://github.com/tianon/gosu/releases/download/${GOSU_VERSION}/gosu-${GOSU_ARCH}" && \
    curl -L -s --output gosu.asc "https://github.com/tianon/gosu/releases/download/${GOSU_VERSION}/gosu-${GOSU_ARCH}.asc" && \
    for server in $(shuf -e ipv4.pool.sks-keyservers.net \
                            hkp://p80.pool.sks-keyservers.net:80 \
                            keyserver.ubuntu.com \
                            hkp://keyserver.ubuntu.com:80 \
                            pgp.mit.edu) ; do \
        gpg --keyserver "$server" --recv-keys B42F6819007F00F88E364FD4036A9C25BF357DD4 && break || : ; \
    done && \
    gpg --batch --verify gosu.asc gosu && \
    chmod +x gosu && \
    cp gosu /bin && \
    gosu --version && \

    # Cleanup
    cd /tmp && \
    rm -rf /tmp/build && \
    gpgconf --kill dirmngr && \
    gpgconf --kill gpg-agent && \
    apk del .build-deps && \
    rm -rf /root/.gnupg

# Set up nsswitch.conf for Go's "netgo" implementation
# - https://github.com/golang/go/blob/go1.9.1/src/net/conf.go#L194-L275
RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf
