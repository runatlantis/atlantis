ARG ATLANTIS_BASE=ghcr.io/runatlantis/atlantis-base
ARG ATLANTIS_BASE_TAG_DATE=latest
ARG ATLANTIS_BASE_TAG_TYPE=alpine

# Stage 1: build artifact

FROM golang:1.19.5-alpine AS builder

ARG ATLANTIS_VERSION=dev
ENV ATLANTIS_VERSION=${ATLANTIS_VERSION}
ARG ATLANTIS_COMMIT=none
ENV ATLANTIS_COMMIT=${ATLANTIS_COMMIT}
ARG ATLANTIS_DATE=unknown
ENV ATLANTIS_DATE=${ATLANTIS_DATE}

WORKDIR /app
COPY . /app

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X 'main.version=${ATLANTIS_VERSION}' -X 'main.commit=${ATLANTIS_COMMIT}' -X 'main.date=${ATLANTIS_DATE}'" -v -o atlantis .

# Stage 2
# The runatlantis/atlantis-base is created by docker-base/Dockerfile
FROM ${ATLANTIS_BASE}:${ATLANTIS_BASE_TAG_DATE}-${ATLANTIS_BASE_TAG_TYPE} AS base

# Get the architecture the image is being built for
ARG TARGETPLATFORM

# install terraform binaries
# renovate: datasource=github-releases depName=hashicorp/terraform versioning=hashicorp
ENV DEFAULT_TERRAFORM_VERSION=1.3.7

# In the official Atlantis image we only have the latest of each Terraform version.
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN AVAILABLE_TERRAFORM_VERSIONS="1.0.11 1.1.9 1.2.9 ${DEFAULT_TERRAFORM_VERSION}" && \
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

# renovate: datasource=github-releases depName=open-policy-agent/conftest
ENV DEFAULT_CONFTEST_VERSION=0.38.0

RUN AVAILABLE_CONFTEST_VERSIONS="${DEFAULT_CONFTEST_VERSION}" && \
    case "${TARGETPLATFORM}" in \
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
        ln -s "/usr/local/bin/cft/versions/${VERSION}/conftest" "/usr/local/bin/conftest${VERSION}" && \
        rm "conftest_${VERSION}_Linux_${CONFTEST_ARCH}.tar.gz" && \
        rm checksums.txt; \
    done

RUN ln -s /usr/local/bin/cft/versions/${DEFAULT_CONFTEST_VERSION}/conftest /usr/local/bin/conftest

# copy binary
COPY --from=builder /app/atlantis /usr/local/bin/atlantis

# copy docker entrypoint
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["server"]
