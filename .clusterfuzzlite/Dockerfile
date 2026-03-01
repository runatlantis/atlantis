FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:be4ca29b2d8855779c893406caf4a9881b7dbf4fb92437e0887e124fcd8a5572
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
