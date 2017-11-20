#!/bin/bash

set -ex
cd $(dirname $0)/..

PACKAGES_TO_SKIP='generate_test_mocks/xtools_go_loader,generate_test_mocks/gomock_reflect,generate_test_mocks/gomock_source'
rm -f mock_display_test.go
rm -rf matchers
$GOPATH/bin/ginkgo -succinct generate_test_mocks/xtools_go_loader
$GOPATH/bin/ginkgo -r -skipPackage=$PACKAGES_TO_SKIP --randomizeAllSpecs --randomizeSuites --race --trace -cover

rm -f mock_display_test.go
rm -rf matchers
$GOPATH/bin/ginkgo -succinct generate_test_mocks/gomock_reflect
$GOPATH/bin/ginkgo --randomizeAllSpecs --randomizeSuites --race --trace -cover

rm -f mock_display_test.go
rm -rf matchers
$GOPATH/bin/ginkgo -succinct generate_test_mocks/gomock_source
$GOPATH/bin/ginkgo --randomizeAllSpecs --randomizeSuites --race --trace -cover
