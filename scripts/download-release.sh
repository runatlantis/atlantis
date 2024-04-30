#!/bin/sh
COMMAND_NAME=${1:-terraform}
TARGETPLATFORM=${2:-"linux/amd64"}
DEFAULT_VERSION=${3:-"1.8.0"}
AVAILABLE_VERSIONS=${4:-"1.8.0"}
case "${TARGETPLATFORM}" in
  "linux/amd64") ARCH=amd64 ;;
  "linux/arm64") ARCH=arm64 ;;
  "linux/arm/v7") ARCH=arm ;;
  *) echo "ERROR: 'TARGETPLATFORM' value unexpected: ${TARGETPLATFORM}"; exit 1 ;;
esac
for VERSION in ${AVAILABLE_VERSIONS}; do
  case "${COMMAND_NAME}" in
    "terraform")
      DOWNLOAD_URL_FORMAT=$(printf 'https://releases.hashicorp.com/terraform/%s/%s_%s' "$VERSION" "$COMMAND_NAME" "$VERSION")
      COMMAND_DIR=/usr/local/bin/terraform
      ;;
    "tofu")
      DOWNLOAD_URL_FORMAT=$(printf 'https://github.com/opentofu/opentofu/releases/download/v%s/%s_%s' "$VERSION" "$COMMAND_NAME" "$VERSION")
      COMMAND_DIR=/usr/local/bin/tofu
      ;;
    *) echo "ERROR: 'COMMAND_NAME' value unexpected: ${COMMAND_NAME}"; exit 1 ;;
  esac
  curl -LOs "${DOWNLOAD_URL_FORMAT}_linux_${ARCH}.zip"
  curl -LOs "${DOWNLOAD_URL_FORMAT}_SHA256SUMS"
  sed -n "/${COMMAND_NAME}_${VERSION}_linux_${ARCH}.zip/p" "${COMMAND_NAME}_${VERSION}_SHA256SUMS" | sha256sum -c
  mkdir -p "${COMMAND_DIR}/${VERSION}"
  unzip "${COMMAND_NAME}_${VERSION}_linux_${ARCH}.zip" -d "${COMMAND_DIR}/${VERSION}"
  ln -s "${COMMAND_DIR}/${VERSION}/${COMMAND_NAME}" "${COMMAND_DIR}/${COMMAND_NAME}${VERSION}"
  rm "${COMMAND_NAME}_${VERSION}_linux_${ARCH}.zip"
  rm "${COMMAND_NAME}_${VERSION}_SHA256SUMS"
done
ln -s "${COMMAND_DIR}/${DEFAULT_VERSION}/${COMMAND_NAME}" "${COMMAND_DIR}/${COMMAND_NAME}"
