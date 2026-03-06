FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:c5c6517ee82c8de02c341dcaddcd03d239fa3a399c4b3627dab5d1d5c32444d3
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
