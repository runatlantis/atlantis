FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:82b4c1a985784f6e640de1da7bed831495c32620d17ad86342fdd71e54601cab
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
