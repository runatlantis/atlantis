FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:2b3cf9ef42c497b71d3916a15e3062b1c0600264bf0e4c02de96bb440b04042f
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
