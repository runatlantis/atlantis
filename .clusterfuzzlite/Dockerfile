FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:65b72e0dfcc817d94ce50022a0691b4378076aaf2b45c5a2dc650a7b73f7d9ed
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
