# This Dockerfile builds the docker image used for running Jenkins tests.
# It is based upon ../testing/Dockerfile which is used for circledi
FROM docker-internal.repo.sonatype.com/cdi/golang-1.14:1

RUN apt-get install -y unzip

# Install Terraform
ENV TERRAFORM_VERSION=0.13.3
RUN curl -LOks https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    mkdir -p /usr/local/bin/tf/versions/${TERRAFORM_VERSION} && \
    unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin/tf/versions/${TERRAFORM_VERSION} && \
    ln -s /usr/local/bin/tf/versions/${TERRAFORM_VERSION}/terraform /usr/local/bin/terraform && \
    rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip
RUN go get golang.org/x/tools/cmd/goimports
