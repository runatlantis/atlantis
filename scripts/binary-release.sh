#!/bin/bash

# define architecture we want to build
XC_ARCH=${XC_ARCH:-"386 amd64 arm arm64"}
XC_OS=${XC_OS:-linux darwin}
XC_EXCLUDE_OSARCH="!darwin/arm !darwin/386 !darwin/arm64"

# clean up
echo "-> running clean up...."
rm -rf output/*

if ! which gox > /dev/null; then
    echo "-> installing gox..."
    # Need to run go get in a separate dir
    # so it doesn't modify our go.mod.
    SRC_DIR=$(pwd)
    cd $(mktemp -d)
    go mod init example.com/m
    go get -u github.com/mitchellh/gox
    cd "$SRC_DIR"
fi

# build
# we want to build statically linked binaries
export CGO_ENABLED=0
echo "-> building..."
gox \
    -os="${XC_OS}" \
    -arch="${XC_ARCH}" \
    -osarch="${XC_EXCLUDE_OSARCH}" \
    -output "output/{{.OS}}_{{.Arch}}/atlantis" \
    .

# Zip and copy to the dist dir
echo ""
echo "Packaging..."
for PLATFORM in $(find ./output -mindepth 1 -maxdepth 1 -type d); do
    OSARCH=$(basename ${PLATFORM})
    echo "--> ${OSARCH}"

    pushd $PLATFORM >/dev/null 2>&1
    zip ../atlantis_${OSARCH}.zip ./*
    popd >/dev/null 2>&1
done

echo ""
echo ""
echo "-----------------------------------"
echo "Output:"
ls -alh output/
