# This Dockerfile builds the docker image used for running circle ci tests.
# We need terraform installed for our full test suite so it installs that.
# It's updated by running make build-testing-image which will also push a new
# image.
FROM circleci/golang:1.15

# Install Terraform
ENV TERRAFORM_VERSION=0.14.6
RUN curl -LOks https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    sudo mkdir -p /usr/local/bin/tf/versions/${TERRAFORM_VERSION} && \
    sudo unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin/tf/versions/${TERRAFORM_VERSION} && \
    sudo ln -s /usr/local/bin/tf/versions/${TERRAFORM_VERSION}/terraform /usr/local/bin/terraform && \
    rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip

# Install conftest
ENV DEFAULT_CONFTEST_VERSION=0.21.0

RUN AVAILABLE_CONFTEST_VERSIONS="${DEFAULT_CONFTEST_VERSION}" && \
    for VERSION in ${AVAILABLE_CONFTEST_VERSIONS}; do \
        curl -LOs https://github.com/open-policy-agent/conftest/releases/download/v${VERSION}/conftest_${VERSION}_Linux_x86_64.tar.gz && \
        curl -LOs https://github.com/open-policy-agent/conftest/releases/download/v${VERSION}/checksums.txt && \
        sed -n "/conftest_${VERSION}_Linux_x86_64.tar.gz/p" checksums.txt | sha256sum -c && \
        sudo mkdir -p /usr/local/bin/cft/versions/${VERSION} && \
        sudo tar -C  /usr/local/bin/cft/versions/${VERSION} -xzf conftest_${VERSION}_Linux_x86_64.tar.gz && \
        sudo ln -s /usr/local/bin/cft/versions/${VERSION}/conftest /usr/local/bin/conftest${VERSION} && \
        rm conftest_${VERSION}_Linux_x86_64.tar.gz && \
        rm checksums.txt; \
    done
RUN go get golang.org/x/tools/cmd/goimports
