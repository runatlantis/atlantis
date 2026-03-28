FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:19d69ca8d903b769fc2c72de77bec957ebf413c87b80a1c6ef2872afd34d97c7
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
