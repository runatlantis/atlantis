FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:c90cf6fc942d2e0f20e420d29e961c7137ea4d3cca239833dadff14fc0c5bd46
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
