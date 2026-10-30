# Image for Anthropic's OSS Scanner (https://github.com/anthropics/oss-scanner, projects/atlantis).
# It lives here, not in that repository, so Renovate keeps its pins in step with go.mod and
# testing/Dockerfile. The scanner builds it with the repository root as context. The root
# .dockerignore only allows what the release image needs, so Dockerfile.dockerignore next to this
# file replaces it and COPY sees the whole checkout.
# Atlantis is a Go server that receives VCS webhooks (GitHub, GitLab, Bitbucket, Azure DevOps, Gitea),
# clones pull requests and runs Terraform/OpenTofu on them. While the network is available this build
# fetches every Go module, installs the Terraform and conftest binaries the tests call, builds the
# server and compiles the test binaries, so the audit can rebuild and rerun tests offline.
# threat_model.md, next to this file, describes how to exercise the result.

# Go toolchain matching go.mod. Renovate groups golang with go.mod, so they bump together: the
# tests below run with GOTOOLCHAIN=local and fail if this falls behind go.mod.
FROM golang:1.27.3-trixie@sha256:fbb7ae40b62ff8afe31662df570d39b4f5052ac32b90dae8e60d7657a20a74f4

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates git make unzip curl \
    && rm -rf /var/lib/apt/lists/*

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# GOPATH and GOCACHE live in a world-writable /go so the root scanner shell and the non-root
# atlantis user (see below) share the same module and build caches.
RUN install -d -m 1777 /go /go/cache
ENV GOPATH=/go \
    GOCACHE=/go/cache

# Terraform and conftest, as testing/Dockerfile installs them. The integration tests in
# server/controllers/events look terraform up on PATH; policy-check tests call conftest.
# renovate: datasource=github-releases depName=hashicorp/terraform versioning=hashicorp
ENV TERRAFORM_VERSION=1.16.6
# renovate: datasource=github-releases depName=open-policy-agent/conftest
ENV CONFTEST_VERSION=0.69.0
RUN case "$(uname -m)" in x86_64) TF_ARCH=amd64 CFT_ARCH=x86_64 ;; aarch64) TF_ARCH=arm64 CFT_ARCH=arm64 ;; \
      *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;; esac \
    && cd /tmp \
    && curl -fsSLO "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${TF_ARCH}.zip" \
    && curl -fsSL -o tf.sums "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_SHA256SUMS" \
    && grep " terraform_${TERRAFORM_VERSION}_linux_${TF_ARCH}.zip$" tf.sums | sha256sum -c - \
    && unzip -q "terraform_${TERRAFORM_VERSION}_linux_${TF_ARCH}.zip" terraform -d /usr/local/bin \
    && curl -fsSLO "https://github.com/open-policy-agent/conftest/releases/download/v${CONFTEST_VERSION}/conftest_${CONFTEST_VERSION}_Linux_${CFT_ARCH}.tar.gz" \
    && curl -fsSL -o cft.sums "https://github.com/open-policy-agent/conftest/releases/download/v${CONFTEST_VERSION}/checksums.txt" \
    && grep " conftest_${CONFTEST_VERSION}_Linux_${CFT_ARCH}.tar.gz$" cft.sums | sha256sum -c - \
    && tar -xzf "conftest_${CONFTEST_VERSION}_Linux_${CFT_ARCH}.tar.gz" -C /usr/local/bin conftest \
    && rm -f /tmp/*.zip /tmp/*.tar.gz /tmp/*.sums

# scanner contract: the checkout lives inside the image, at /src
COPY . /src
WORKDIR /src

# Fetch every module the build or the tests need (the root module plus the nested e2e and CI-script
# modules); after the image build there is no network.
RUN go mod download \
    && find . -name go.mod -not -path ./go.mod -execdir go mod download \;

# Build the server the way `make build-service` does (Go keeps DWARF debug info by default).
RUN CGO_ENABLED=0 go build -o /usr/local/bin/atlantis . \
    && atlantis version

# Compile every test binary without running it, so the suite stays buildable offline.
RUN go test -count=1 -run '^$' ./... >/dev/null

# Everything is fetched. From here on the go command must not reach a proxy or download a toolchain,
# so the test run below also proves the image works offline.
ENV GOPROXY=off \
    GOSUMDB=off \
    GOTOOLCHAIN=local

RUN git config --system --add safe.directory '*' \
    && git config --system user.name scanner \
    && git config --system user.email scanner@localhost

# Some tests check that unreadable files are rejected, which root can always read, so the suite runs
# as a non-root user (uid 1001, as in the project's testing/Dockerfile). threat_model.md explains how
# to rerun it from the scanner's root shell. A failing test does not fail the image.
RUN useradd -u 1001 --create-home --shell /bin/bash atlantis \
    && chown -R atlantis:atlantis /src /go
RUN runuser -u atlantis -- env HOME=/home/atlantis timeout 1200 make test \
    || echo "WARNING: make test reported failures"
