# syntax=docker/dockerfile:1@sha256:38387523653efa0039f8e1c89bb74a30504e76ee9f565e25c9a09841f9427b05
# what distro is the image being built for
ARG ALPINE_TAG=3.21.3@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c
ARG DEBIAN_TAG=12.10-slim@sha256:4b50eb66f977b4062683ff434ef18ac191da862dbe966961bc11990cf5791a8d
# renovate: datasource=docker depName=golang versioning=docker
ARG GOLANG_TAG=1.24.4-alpine@sha256:68932fa6d4d4059845c8f40ad7e654e626f3ebd3706eef7846f319293ab5cb7a

# renovate: datasource=github-releases depName=hashicorp/terraform versioning=hashicorp
ARG DEFAULT_TERRAFORM_VERSION=1.11.4
# renovate: datasource=github-releases depName=opentofu/opentofu versioning=hashicorp
ARG DEFAULT_OPENTOFU_VERSION=1.10.5
# renovate: datasource=github-releases depName=open-policy-agent/conftest
ARG DEFAULT_CONFTEST_VERSION=0.59.0

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
# renovate: datasource=repology depName=alpine_3_21/bash versioning=loose
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
ENV DEBIAN_CURL_VERSION="7.88.1-10+deb12u12"
# renovate: datasource=repology depName=debian_12/git versioning=loose
ENV DEBIAN_GIT_VERSION="1:2.39.5-0+deb12u2"
# renovate: datasource=repology depName=debian_12/unzip versioning=loose
ENV DEBIAN_UNZIP_VERSION="6.0-28"
# renovate: datasource=repology depName=debian_12/openssh-server versioning=loose
ENV DEBIAN_OPENSSH_SERVER_VERSION="1:9.2p1-2+deb12u7"
# renovate: datasource=repology depName=debian_12/dumb-init versioning=loose
ENV DEBIAN_DUMB_INIT_VERSION="1.2.5-2"
# renovate: datasource=repology depName=debian_12/gnupg versioning=loose
ENV DEBIAN_GNUPG_VERSION="2.2.40-1.1"
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
ENV GIT_LFS_VERSION=3.6.1

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

# renovate: datasource=repology depName=alpine_3_21/ca-certificates versioning=loose
ENV CA_CERTIFICATES_VERSION="20250619-r0"
# renovate: datasource=repology depName=alpine_3_21/curl versioning=loose
ENV CURL_VERSION="8.12.1-r1"
# renovate: datasource=repology depName=alpine_3_21/git versioning=loose
ENV GIT_VERSION="2.47.3-r0"
# renovate: datasource=repology depName=alpine_3_21/unzip versioning=loose
ENV UNZIP_VERSION="6.0-r15"
# renovate: datasource=repology depName=alpine_3_21/bash versioning=loose
ENV BASH_VERSION="5.2.37-r0"
# renovate: datasource=repology depName=alpine_3_21/openssh versioning=loose
ENV OPENSSH_VERSION="9.9_p2-r0"
# renovate: datasource=repology depName=alpine_3_21/dumb-init versioning=loose
ENV DUMB_INIT_VERSION="1.2.5-r3"
# renovate: datasource=repology depName=alpine_3_21/gcompat versioning=loose
ENV GCOMPAT_VERSION="1.1.0-r4"
# renovate: datasource=repology depName=alpine_3_21/coreutils versioning=loose
ENV COREUTILS_ENV_VERSION="9.5-r2"

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
