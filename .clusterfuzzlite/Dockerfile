FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:b940188f7306d865c69127852ec33e9c821287e6fbebaa9feb89ccc09a6e9b50
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
