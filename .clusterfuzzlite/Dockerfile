FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:d510f8656e9447c2f7293a179676c809cd8261078c80cc08a55699a3262fea0d
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
