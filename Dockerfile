# syntax=docker/dockerfile:1@sha256:b6afd42430b15f2d2a4c5a02b919e98a525b785b1aaff16747d2f623364e39b6
# what distro is the image being built for
ARG ALPINE_TAG=3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659
ARG DEBIAN_TAG=12.13-slim@sha256:74d56e3931e0d5a1dd51f8c8a2466d21de84a271cd3b5a733b803aa91abf4421
# renovate: datasource=docker depName=golang versioning=docker
ARG GOLANG_TAG=1.25.4-alpine@sha256:d3f0cf7723f3429e3f9ed846243970b20a2de7bae6a5b66fc5914e228d831bbb

# renovate: datasource=github-releases depName=hashicorp/terraform versioning=hashicorp
ARG DEFAULT_TERRAFORM_VERSION=1.14.5
# renovate: datasource=github-releases depName=opentofu/opentofu versioning=hashicorp
ARG DEFAULT_OPENTOFU_VERSION=1.11.5
# renovate: datasource=github-releases depName=open-policy-agent/conftest
ARG DEFAULT_CONFTEST_VERSION=0.66.0

# Stage 1: build artifact and download deps

FROM --platform=$BUILDPLATFORM golang:${GOLANG_TAG} AS builder

# These are automatically populated by Docker
ARG TARGETOS
ARG TARGETARCH

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
# renovate: datasource=repology depName=alpine_3_22/bash versioning=loose
ENV BUILDER_BASH_VERSION="5.2.37-r0"
RUN apk add --no-cache \
        bash=${BUILDER_BASH_VERSION}
COPY go.mod go.sum ./
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod graph | awk '{if ($1 !~ "@") print $2}' | xargs go get

COPY . /app
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags "-s -w -X 'main.version=${ATLANTIS_VERSION}' -X 'main.commit=${ATLANTIS_COMMIT}' -X 'main.date=${ATLANTIS_DATE}'" -v -o atlantis .

FROM debian:${DEBIAN_TAG} AS debian-base

# Define package versions for Debian
# renovate: datasource=repology depName=debian_12/ca-certificates versioning=loose
ENV DEBIAN_CA_CERTIFICATES_VERSION="20230311+deb12u1"
# renovate: datasource=repology depName=debian_12/curl versioning=loose
ENV DEBIAN_CURL_VERSION="7.88.1-10+deb12u14"
# renovate: datasource=repology depName=debian_12/git versioning=loose
ENV DEBIAN_GIT_VERSION="1:2.39.5-0+deb12u2"
# renovate: datasource=repology depName=debian_12/unzip versioning=loose
ENV DEBIAN_UNZIP_VERSION="6.0-28"
# renovate: datasource=repology depName=debian_12/openssh-server versioning=loose
ENV DEBIAN_OPENSSH_SERVER_VERSION="1:9.2p1-2+deb12u7"
# renovate: datasource=repology depName=debian_12/dumb-init versioning=loose
ENV DEBIAN_DUMB_INIT_VERSION="1.2.5-2"
# renovate: datasource=repology depName=debian_12/gnupg versioning=loose
ENV DEBIAN_GNUPG_VERSION="2.2.40-1.1+deb12u2"
# renovate: datasource=repology depName=debian_12/openssl versioning=loose
ENV DEBIAN_OPENSSL_VERSION="3.0.17-1~deb12u2"

# Install packages needed to run Atlantis.
# We place this last as it will bust less docker layer caches when packages update
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates=${DEBIAN_CA_CERTIFICATES_VERSION} \
        curl=${DEBIAN_CURL_VERSION} \
        git=${DEBIAN_GIT_VERSION} \
        unzip=${DEBIAN_UNZIP_VERSION} \
        openssh-server=${DEBIAN_OPENSSH_SERVER_VERSION} \
        dumb-init=${DEBIAN_DUMB_INIT_VERSION} \
        gnupg=${DEBIAN_GNUPG_VERSION} \
        openssl=${DEBIAN_OPENSSL_VERSION} && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

FROM debian-base AS deps

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
ENV GIT_LFS_VERSION=3.7.1

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
        "1.8.5 1.9.8 1.10.5 ${DEFAULT_TERRAFORM_VERSION}" \
    && ./download-release.sh \
        "tofu" \
        "${TARGETPLATFORM}" \
        "${DEFAULT_OPENTOFU_VERSION}" \
        "${DEFAULT_OPENTOFU_VERSION}"

# Stage 2 - Alpine
# Creating the individual distro builds using targets
FROM alpine:${ALPINE_TAG} AS alpine-base

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
# copy dependencies
COPY --from=deps /usr/local/bin/conftest /usr/local/bin/conftest
COPY --from=deps /usr/bin/git-lfs /usr/bin/git-lfs
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

# renovate: datasource=repology depName=alpine_3_23/ca-certificates versioning=loose
ENV CA_CERTIFICATES_VERSION="20251003-r0"
# renovate: datasource=repology depName=alpine_3_23/curl versioning=loose
ENV CURL_VERSION="8.17.0-r1"
# renovate: datasource=repology depName=alpine_3_23/git versioning=loose
ENV GIT_VERSION="2.52.0-r0"
# renovate: datasource=repology depName=alpine_3_23/unzip versioning=loose
ENV UNZIP_VERSION="6.0-r16"
# renovate: datasource=repology depName=alpine_3_23/bash versioning=loose
ENV BASH_VERSION="5.3.3-r1"
# renovate: datasource=repology depName=alpine_3_23/openssh versioning=loose
ENV OPENSSH_VERSION="10.2_p1-r0"
# renovate: datasource=repology depName=alpine_3_23/dumb-init versioning=loose
ENV DUMB_INIT_VERSION="1.2.5-r3"
# renovate: datasource=repology depName=alpine_3_23/gcompat versioning=loose
ENV GCOMPAT_VERSION="1.1.0-r4"
# renovate: datasource=repology depName=alpine_3_23/coreutils versioning=loose
ENV COREUTILS_ENV_VERSION="9.8-r1"

# Install packages needed to run Atlantis.
# We place this last as it will bust less docker layer caches when packages update
RUN apk add --no-cache \
        ca-certificates=${CA_CERTIFICATES_VERSION} \
        curl=${CURL_VERSION} \
        git=${GIT_VERSION} \
        unzip=${UNZIP_VERSION} \
        bash=${BASH_VERSION} \
        openssh=${OPENSSH_VERSION} \
        dumb-init=${DUMB_INIT_VERSION} \
        gcompat=${GCOMPAT_VERSION} \
        coreutils-env=${COREUTILS_ENV_VERSION}

ARG DEFAULT_CONFTEST_VERSION
ENV DEFAULT_CONFTEST_VERSION=${DEFAULT_CONFTEST_VERSION}

# Set the entry point to the atlantis user and run the atlantis command
USER atlantis
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["server"]

FROM alpine-base as alpine
# copy terraform binaries
COPY --from=deps /usr/local/bin/terraform/terraform* /usr/local/bin/
COPY --from=deps /usr/local/bin/tofu/tofu* /usr/local/bin/

FROM alpine-base as alpine-slim-terraform
COPY --from=deps /usr/local/bin/terraform/terraform /usr/local/bin/terraform

FROM alpine-base as alpine-slim-tofu
COPY --from=deps /usr/local/bin/tofu/tofu /usr/local/bin/tofu

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

ARG DEFAULT_CONFTEST_VERSION
ENV DEFAULT_CONFTEST_VERSION=${DEFAULT_CONFTEST_VERSION}

# Set the entry point to the atlantis user and run the atlantis command
USER atlantis
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["server"]
