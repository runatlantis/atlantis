FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:3a8404e215d13241ab0a1296102e9ddcb174db67a990a56af4f3d19cb38807d1
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
