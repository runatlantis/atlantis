FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:bb738df2f5838003c1987203d4ad456a13fca20916c09eb875dace4205922ea2
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
