FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:aa0051c72cbb89145b94d369b5fdb37fb0e346d1193bb04c15c8e671f2d1916c
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
