# syntax=docker/dockerfile:1
# what distro is the image being built for
ARG ALPINE_TAG=3.19.0
ARG DEBIAN_TAG=12.4-slim

ARG DEFAULT_TERRAFORM_VERSION=1.5.7
ARG DEFAULT_CONFTEST_VERSION=0.46.0

# Stage 1: build artifact and download deps

FROM golang:1.21.5-alpine AS builder

ARG ATLANTIS_VERSION=dev
ENV ATLANTIS_VERSION=${ATLANTIS_VERSION}
ARG ATLANTIS_COMMIT=none
ENV ATLANTIS_COMMIT=${ATLANTIS_COMMIT}
ARG ATLANTIS_DATE=unknown
ENV ATLANTIS_DATE=${ATLANTIS_DATE}

ARG DEFAULT_TERRAFORM_VERSION
ENV DEFAULT_TERRAFORM_VERSION=${DEFAULT_TERRAFORM_VERSION}
ARG DEFAULT_CONFTEST_VERSION
ENV DEFAULT_CONFTEST_VERSION=${DEFAULT_CONFTEST_VERSION}

WORKDIR /app

# This is needed to download transitive dependencies instead of compiling them
# https://github.com/montanaflynn/golang-docker-cache
# https://github.com/golang/go/issues/27719
RUN apk add --no-cache \
        bash~=5.2
COPY go.mod go.sum ./
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod graph | awk '{if ($1 !~ "@") print $2}' | xargs go get

COPY . /app
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X 'main.version=${ATLANTIS_VERSION}' -X 'main.commit=${ATLANTIS_COMMIT}' -X 'main.date=${ATLANTIS_DATE}'" -v -o atlantis .

FROM debian:${DEBIAN_TAG} as debian-base

# Install packages needed to run Atlantis.
# We place this last as it will bust less docker layer caches when packages update
# hadolint ignore explanation
# DL3008 (pin versions using "=") - Ignored to avoid failing the build
# hadolint ignore=DL3008
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
        unzip \
        openssh-server \
        libcap2 \
        dumb-init \
        gnupg \
        openssl && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

FROM debian-base as deps

# Get the architecture the image is being built for
ARG TARGETPLATFORM
WORKDIR /tmp/build

# install conftest
# renovate: datasource=github-releases depName=open-policy-agent/conftest
ARG DEFAULT_CONFTEST_VERSION
ENV DEFAULT_CONFTEST_VERSION=${DEFAULT_CONFTEST_VERSION}
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN AVAILABLE_CONFTEST_VERSIONS=${DEFAULT_CONFTEST_VERSION} && \
    case ${TARGETPLATFORM} in \
        "linux/amd64") CONFTEST_ARCH=x86_64 ;; \
        "linux/arm64") CONFTEST_ARCH=arm64 ;; \
        # There is currently no compiled version of conftest for armv7
        "linux/arm/v7") CONFTEST_ARCH=x86_64 ;; \
    esac && \
    for VERSION in ${AVAILABLE_CONFTEST_VERSIONS}; do \
        curl -LOs "https://github.com/open-policy-agent/conftest/releases/download/v${VERSION}/conftest_${VERSION}_Linux_${CONFTEST_ARCH}.tar.gz" && \
        curl -LOs "https://github.com/open-policy-agent/conftest/releases/download/v${VERSION}/checksums.txt" && \
        sed -n "/conftest_${VERSION}_Linux_${CONFTEST_ARCH}.tar.gz/p" checksums.txt | sha256sum -c && \
        mkdir -p "/usr/local/bin/cft/versions/${VERSION}" && \
        tar -C "/usr/local/bin/cft/versions/${VERSION}" -xzf "conftest_${VERSION}_Linux_${CONFTEST_ARCH}.tar.gz" && \
        ln -s "/usr/local/bin/cft/versions/${VERSION}/conftest" /usr/local/bin/conftest && \
        rm "conftest_${VERSION}_Linux_${CONFTEST_ARCH}.tar.gz" && \
        rm checksums.txt; \
    done

# install git-lfs
# renovate: datasource=github-releases depName=git-lfs/git-lfs
ENV GIT_LFS_VERSION=3.4.1

RUN case ${TARGETPLATFORM} in \
        "linux/amd64") GIT_LFS_ARCH=amd64 ;; \
        "linux/arm64") GIT_LFS_ARCH=arm64 ;; \
        "linux/arm/v7") GIT_LFS_ARCH=arm ;; \
    esac && \
    curl -L -s --output git-lfs.tar.gz "https://github.com/git-lfs/git-lfs/releases/download/v${GIT_LFS_VERSION}/git-lfs-linux-${GIT_LFS_ARCH}-v${GIT_LFS_VERSION}.tar.gz" && \
    tar --strip-components=1 -xf git-lfs.tar.gz && \
    chmod +x git-lfs && \
    mv git-lfs /usr/bin/git-lfs && \
    git-lfs --version

# install terraform binaries
# renovate: datasource=github-releases depName=hashicorp/terraform versioning=hashicorp
ARG DEFAULT_TERRAFORM_VERSION
ENV DEFAULT_TERRAFORM_VERSION=${DEFAULT_TERRAFORM_VERSION}

# In the official Atlantis image, we only have the latest of each Terraform version.
# Each binary is about 80 MB so we limit it to the 4 latest minor releases or fewer
RUN AVAILABLE_TERRAFORM_VERSIONS="1.2.9 1.3.10 1.4.6 ${DEFAULT_TERRAFORM_VERSION}" && \
    case "${TARGETPLATFORM}" in \
        "linux/amd64") TERRAFORM_ARCH=amd64 ;; \
        "linux/arm64") TERRAFORM_ARCH=arm64 ;; \
        "linux/arm/v7") TERRAFORM_ARCH=arm ;; \
        *) echo "ERROR: 'TARGETPLATFORM' value expected: ${TARGETPLATFORM}"; exit 1 ;; \
    esac && \
    for VERSION in ${AVAILABLE_TERRAFORM_VERSIONS}; do \
        curl -LOs "https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_linux_${TERRAFORM_ARCH}.zip" && \
        curl -LOs "https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_SHA256SUMS" && \
        sed -n "/terraform_${VERSION}_linux_${TERRAFORM_ARCH}.zip/p" "terraform_${VERSION}_SHA256SUMS" | sha256sum -c && \
        mkdir -p "/usr/local/bin/tf/versions/${VERSION}" && \
        unzip "terraform_${VERSION}_linux_${TERRAFORM_ARCH}.zip" -d "/usr/local/bin/tf/versions/${VERSION}" && \
        ln -s "/usr/local/bin/tf/versions/${VERSION}/terraform" "/usr/local/bin/terraform${VERSION}" && \
        rm "terraform_${VERSION}_linux_${TERRAFORM_ARCH}.zip" && \
        rm "terraform_${VERSION}_SHA256SUMS"; \
    done && \
    ln -s "/usr/local/bin/tf/versions/${DEFAULT_TERRAFORM_VERSION}/terraform" /usr/local/bin/terraform


# Stage 2 - Alpine
# Creating the individual distro builds using targets
FROM alpine:${ALPINE_TAG} AS alpine

EXPOSE ${ATLANTIS_PORT:-4141}

HEALTHCHECK --interval=5m --timeout=3s \
  CMD curl -f http://localhost:${ATLANTIS_PORT:-4141}/healthz || exit 1

# Set up the 'atlantis' user and adjust permissions
RUN addgroup atlantis && \
    adduser -S -G atlantis atlantis && \
    chown atlantis:root /home/atlantis/ && \
    chmod u+rwx /home/atlantis/

# copy atlantis binary
COPY --from=builder /app/atlantis /usr/local/bin/atlantis
# copy terraform binaries
COPY --from=deps /usr/local/bin/terraform* /usr/local/bin/
# copy dependencies
COPY --from=deps /usr/local/bin/conftest /usr/local/bin/conftest
COPY --from=deps /usr/bin/git-lfs /usr/bin/git-lfs
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

# Install packages needed to run Atlantis.
# We place this last as it will bust less docker layer caches when packages update
RUN apk add --no-cache \
        ca-certificates~=20230506 \
        curl~=8.5 \
        git~=2.40 \
        unzip~=6.0 \
        bash~=5.2 \
        openssh~=9.3_p2 \
        libcap~=2.69 \
        dumb-init~=1.2 \
        gcompat~=1.1


# Set the entry point to the atlantis user and run the atlantis command
USER atlantis
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["server"]

# Stage 2 - Debian
FROM debian-base AS debian

EXPOSE ${ATLANTIS_PORT:-4141}

HEALTHCHECK --interval=5m --timeout=3s \
  CMD curl -f http://localhost:${ATLANTIS_PORT:-4141}/healthz || exit 1

# Set up the 'atlantis' user and adjust permissions
RUN useradd --create-home --user-group --shell /bin/bash atlantis && \
    chown atlantis:root /home/atlantis/ && \
    chmod u+rwx /home/atlantis/

# copy atlantis binary
COPY --from=builder /app/atlantis /usr/local/bin/atlantis
# copy terraform binaries
COPY --from=deps /usr/local/bin/terraform* /usr/local/bin/
# copy dependencies
COPY --from=deps /usr/local/bin/conftest /usr/local/bin/conftest
COPY --from=deps /usr/bin/git-lfs /usr/bin/git-lfs
# copy docker-entrypoint.sh
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

# Set the entry point to the atlantis user and run the atlantis command
USER atlantis
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["server"]
