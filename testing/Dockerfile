FROM circleci/golang:1.17

# Install Terraform
ENV TERRAFORM_VERSION=1.0.11
RUN curl -LOks https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    sudo mkdir -p /usr/local/bin/tf/versions/${TERRAFORM_VERSION} && \
    sudo unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin/tf/versions/${TERRAFORM_VERSION} && \
    sudo ln -s /usr/local/bin/tf/versions/${TERRAFORM_VERSION}/terraform /usr/local/bin/terraform && \
    rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip

# Install conftest
ENV CONFTEST_VERSION=0.28.3
RUN curl -LOs https://github.com/open-policy-agent/conftest/releases/download/v${CONFTEST_VERSION}/conftest_${CONFTEST_VERSION}_Linux_x86_64.tar.gz && \
    curl -LOs https://github.com/open-policy-agent/conftest/releases/download/v${CONFTEST_VERSION}/checksums.txt && \
    sed -n "/conftest_${CONFTEST_VERSION}_Linux_x86_64.tar.gz/p" checksums.txt | sha256sum -c && \
    sudo mkdir -p /usr/local/bin/cft/versions/${CONFTEST_VERSION} && \
    sudo tar -C  /usr/local/bin/cft/versions/${CONFTEST_VERSION} -xzf conftest_${CONFTEST_VERSION}_Linux_x86_64.tar.gz && \
    sudo ln -s /usr/local/bin/cft/versions/${CONFTEST_VERSION}/conftest /usr/local/bin/conftest${CONFTEST_VERSION} && \
    rm conftest_${CONFTEST_VERSION}_Linux_x86_64.tar.gz && \
    rm checksums.txt

RUN go get golang.org/x/tools/cmd/goimports
