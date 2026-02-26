FROM gcr.io/oss-fuzz-base/base-builder-go@sha256:153f35fc70b091cb68d7243476b452c30ea5ff40504f970d60ac49d2f949293c
COPY . $SRC/atlantis
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/atlantis
