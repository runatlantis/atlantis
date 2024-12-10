FROM golang:1.23.3@sha256:73f06be4578c9987ce560087e2e2ea6485fb605e3910542cadd8fa09fc5f3e31

RUN apt-get update && apt-get --no-install-recommends -y install unzip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Terraform
# renovate: datasource=github-releases depName=hashicorp/terraform versioning=hashicorp
ENV TERRAFORM_VERSION=1.10.1
RUN case $(uname -m) in x86_64|amd64) ARCH="amd64" ;; aarch64|arm64|armv7l) ARCH="arm64" ;; esac && \
    wget -nv -O terraform.zip https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${ARCH}.zip && \
    mkdir -p /usr/local/bin/tf/versions/${TERRAFORM_VERSION} && \
    unzip terraform.zip -d /usr/local/bin/tf/versions/${TERRAFORM_VERSION} && \
    ln -s /usr/local/bin/tf/versions/${TERRAFORM_VERSION}/terraform /usr/local/bin/terraform && \
    rm terraform.zip

# Install conftest
# renovate: datasource=github-releases depName=open-policy-agent/conftest
ENV CONFTEST_VERSION=0.56.0
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN case $(uname -m) in x86_64|amd64) ARCH="x86_64" ;; aarch64|arm64|armv7l) ARCH="arm64" ;; esac && \
    curl -LOs https://github.com/open-policy-agent/conftest/releases/download/v${CONFTEST_VERSION}/conftest_${CONFTEST_VERSION}_Linux_${ARCH}.tar.gz && \
    curl -LOs https://github.com/open-policy-agent/conftest/releases/download/v${CONFTEST_VERSION}/checksums.txt && \
    sed -n "/conftest_${CONFTEST_VERSION}_Linux_${ARCH}.tar.gz/p" checksums.txt | sha256sum -c && \
    mkdir -p /usr/local/bin/cft/versions/${CONFTEST_VERSION} && \
    tar -C  /usr/local/bin/cft/versions/${CONFTEST_VERSION} -xzf conftest_${CONFTEST_VERSION}_Linux_${ARCH}.tar.gz && \
    # Generally Atlantis requires `conftest$version` command. But we use `conftest` command in test.
    # `conftest$version` command blocks upgrading conftest operation cause e2e test use this image.
    ln -s /usr/local/bin/cft/versions/${CONFTEST_VERSION}/conftest /usr/local/bin/conftest && \
    rm conftest_${CONFTEST_VERSION}_Linux_${ARCH}.tar.gz && \
    rm checksums.txt

RUN useradd -u 1001 -m atlantis

USER atlantis
