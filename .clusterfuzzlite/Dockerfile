FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:bb153df7a0308f72195d3f8b3d48c98a06ae0db7a7e350a1bfea68a15ec1b785
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
