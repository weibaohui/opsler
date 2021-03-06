#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

vendor/k8s.io/code-generator/generate-groups.sh \
deepcopy \
github.com/opsler/opsler/opsler-operator/pkg/generated \
github.com/opsler/opsler/opsler-operator/pkg/apis \
opsler:v1alpha1 \
--go-header-file "./tmp/codegen/boilerplate.go.txt"


vendor/k8s.io/code-generator/generate-groups.sh \
deepcopy \
github.com/opsler/opsler/opsler-operator/pkg/generated2 \
github.com/opsler/opsler/opsler-operator/pkg/istio/apis \
istio:v1alpha3 \
--go-header-file "./tmp/codegen/boilerplate.go.txt"
