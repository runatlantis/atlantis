# The runatlantis/atlantis-base is created by docker-base/Dockerfile.
FROM golang:1.12 as builder
WORKDIR /go/src/github.com/runatlantis/atlantis
COPY . /go/src/github.com/runatlantis/atlantis
RUN CGO_ENABLED=0 go build -o atlantis main.go
FROM runatlantis/atlantis-base:v3.0
LABEL authors="Anubhav Mishra, Luke Kysow"

# install terraform binaries
ENV DEFAULT_TERRAFORM_VERSION=0.11.13

# In the official Atlantis image we only have the latest of each Terrafrom version.
RUN AVAILABLE_TERRAFORM_VERSIONS="0.8.8 0.9.11 0.10.8 ${DEFAULT_TERRAFORM_VERSION}" && \
    for VERSION in ${AVAILABLE_TERRAFORM_VERSIONS}; do \
        curl -LOks https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_linux_amd64.zip && \
        curl -LOks https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_SHA256SUMS && \
        sed -n "/terraform_${VERSION}_linux_amd64.zip/p" terraform_${VERSION}_SHA256SUMS | sha256sum -c && \
        mkdir -p /usr/local/bin/tf/versions/${VERSION} && \
        unzip terraform_${VERSION}_linux_amd64.zip -d /usr/local/bin/tf/versions/${VERSION} && \
        ln -s /usr/local/bin/tf/versions/${VERSION}/terraform /usr/local/bin/terraform${VERSION} && \
        rm terraform_${VERSION}_linux_amd64.zip && \
        rm terraform_${VERSION}_SHA256SUMS; \
    done && \
    ln -s /usr/local/bin/tf/versions/${DEFAULT_TERRAFORM_VERSION}/terraform /usr/local/bin/terraform

# copy binary
COPY --from=builder /go/src/github.com/runatlantis/atlantis/atlantis /usr/local/bin/atlantis

# copy docker entrypoint
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["server"]
