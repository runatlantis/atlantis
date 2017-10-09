#!/bin/bash

set -ex

rm -f mock_display_test.go
$GOPATH/bin/ginkgo -succinct generate_test_mocks/xtools_go_loader
$GOPATH/bin/ginkgo -r -skipPackage=generate_test_mocks/xtools_go_loader,generate_test_mocks/gomock_reflect,generate_test_mocks/gomock_source --randomizeAllSpecs --randomizeSuites --race --trace -cover

rm -f mock_display_test.go
$GOPATH/bin/ginkgo -succinct generate_test_mocks/gomock_reflect
$GOPATH/bin/ginkgo -r -skipPackage=generate_test_mocks/xtools_go_loader,generate_test_mocks/gomock_reflect,generate_test_mocks/gomock_source --randomizeAllSpecs --randomizeSuites --race --trace -cover

rm -f mock_display_test.go
$GOPATH/bin/ginkgo -succinct generate_test_mocks/gomock_source
$GOPATH/bin/ginkgo -r -skipPackage=generate_test_mocks/xtools_go_loader,generate_test_mocks/gomock_reflect,generate_test_mocks/gomock_source --randomizeAllSpecs --randomizeSuites --race --trace -cover
