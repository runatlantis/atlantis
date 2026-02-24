FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:4fdcdc77a3f18074faea7cacabb1af430d8cab57e536b129fac02838535d4e83
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
