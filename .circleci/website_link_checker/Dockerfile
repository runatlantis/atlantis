# This Dockerfile builds runatlantis/ci-link-checker.
# It is used in CircleCI to check if the website has any broken links.
FROM node:14
ENV DOCKERIZE_VERSION v0.6.1

# Muffet is used to check for broken links.
COPY --from=raviqqe/muffet:2.4.0 muffet /usr/local/bin/muffet

# http-server is used to serve the website locally as muffet checks it.
RUN yarn global add http-server

# Dockerize is used to wait until the server is up and running before running
# muffet.
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz
