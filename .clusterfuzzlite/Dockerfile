FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:939c41f02015510f186bc1ae2d39d84736585aecd7e9628cacde6d3bc87c3f61
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
