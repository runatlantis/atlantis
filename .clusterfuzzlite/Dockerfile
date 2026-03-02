FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:3023a59112a77c74961406840198a791bd68cb341d977314c8728301d91f6c8d
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
