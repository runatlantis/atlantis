FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:ff0fd2c3f71575b627cc642b6c19df4c824a43e351d9fb8d0ca19062469ec219
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
