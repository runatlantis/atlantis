FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:e7292bef11ae12cbe41550299c54dad85decc77c2b7bd262087b4d955f5db4c8
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
