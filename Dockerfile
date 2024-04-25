# syntax=docker/dockerfile:1@sha256:dbbd5e059e8a07ff7ea6233b213b36aa516b4c53c645f1817a4dd18b83cbea56
# what distro is the image being built for
ARG ALPINE_TAG=3.19.1@sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b
ARG DEBIAN_TAG=12.5-slim@sha256:155280b00ee0133250f7159b567a07d7cd03b1645714c3a7458b2287b0ca83cb
ARG GOLANG_TAG=1.22.1-alpine

# renovate: datasource=github-releases depName=hashicorp/terraform versioning=hashicorp
ARG DEFAULT_TERRAFORM_VERSION=1.8.0
# renovate: datasource=github-releases depName=hashicorp/terraform versioning=hashicorp
ARG DEFAULT_OPENTOFU_VERSION=1.6.2
# renovate: datasource=github-releases depName=open-policy-agent/conftest
ARG DEFAULT_CONFTEST_VERSION=0.49.1

# Stage 1: build artifact and download deps

FROM golang:${GOLANG_TAG} AS builder

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
ENV GIT_LFS_VERSION=3.5.1

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
ARG DEFAULT_TERRAFORM_VERSION
ENV DEFAULT_TERRAFORM_VERSION=${DEFAULT_TERRAFORM_VERSION}
ARG DEFAULT_OPENTOFU_VERSION
ENV DEFAULT_OPENTOFU_VERSION=${DEFAULT_OPENTOFU_VERSION}

# COPY scripts/download-release.sh .
COPY --from=builder /app/scripts/download-release.sh download-release.sh

# In the official Atlantis image, we only have the latest of each Terraform version.
# Each binary is about 80 MB so we limit it to the 4 latest minor releases or fewer
RUN ./download-release.sh \
        "terraform" \
        "${TARGETPLATFORM}" \
        "${DEFAULT_TERRAFORM_VERSION}" \
        "1.5.7 1.6.6 1.7.5 ${DEFAULT_TERRAFORM_VERSION}" \
    && ./download-release.sh \
        "tofu" \
        "${TARGETPLATFORM}" \
        "${DEFAULT_OPENTOFU_VERSION}" \
        "${DEFAULT_OPENTOFU_VERSION}"

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
COPY --from=deps /usr/local/bin/terraform/terraform* /usr/local/bin/
COPY --from=deps /usr/local/bin/tofu/tofu* /usr/local/bin/
# copy dependencies
COPY --from=deps /usr/local/bin/conftest /usr/local/bin/conftest
COPY --from=deps /usr/bin/git-lfs /usr/bin/git-lfs
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

# Install packages needed to run Atlantis.
# We place this last as it will bust less docker layer caches when packages update
RUN apk add --no-cache \
        ca-certificates~=20240226-r0 \
        curl~=8 \
        git~=2 \
        unzip~=6 \
        bash~=5 \
        openssh~=9 \
        dumb-init~=1 \
        gcompat~=1

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
COPY --from=deps /usr/local/bin/terraform/terraform* /usr/local/bin/
COPY --from=deps /usr/local/bin/tofu/tofu* /usr/local/bin/
# copy dependencies
COPY --from=deps /usr/local/bin/conftest /usr/local/bin/conftest
COPY --from=deps /usr/bin/git-lfs /usr/bin/git-lfs
# copy docker-entrypoint.sh
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

# Set the entry point to the atlantis user and run the atlantis command
USER atlantis
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["server"]
