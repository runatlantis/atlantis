FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:b0a1a6f705621381e4e0acf74a56920bf7888256693ffde78fa793a5413b4cb3
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
