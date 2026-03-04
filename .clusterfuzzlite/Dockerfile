FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:92d6572f97afbb1ffdb0f14b0816f5fd7a1918d88d30fbf751b1d931bd86af3d
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
