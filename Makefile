# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

ENSURE_GARDENER_MOD := $(shell go get github.com/gardener/gardener@$$(go list -m -f "{{.Version}}" github.com/gardener/gardener))
GARDENER_HACK_DIR   := $(shell go list -m -f "{{.Dir}}" github.com/gardener/gardener)/hack


REPO_ROOT           := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
HACK_DIR            := $(REPO_ROOT)/hack
COVERPROFILE        := test/output/coverprofile.out
NAME                := machine-controller-manager-provider-hcloud
IMAGE_NAME          := $(IMAGE_PREFIX)/$(NAME)
VERSION             := $(shell cat VERSION)

LEADER_ELECT 	    := "true"
# If Integration Test Suite is to be run locally against clusters then export the below variable
# with MCM deployment name in the cluster
MACHINE_CONTROLLER_MANAGER_DEPLOYMENT_NAME := machine-controller-manager
PLATFORM ?= linux/amd64

#########################################
# Tools & Cleanup
#########################################

TOOLS_DIR := $(HACK_DIR)/tools
include $(GARDENER_HACK_DIR)/tools.mk
-include .env

.PHONY: tidy
tidy:
	@go mod tidy
	@mkdir -p $(REPO_ROOT)/.ci/hack && cp $(GARDENER_HACK_DIR)/.ci/* $(REPO_ROOT)/.ci/hack/ && chmod +xw $(REPO_ROOT)/.ci/hack/*
	@GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) bash $(REPO_ROOT)/hack/update-github-templates.sh
	@cp $(GARDENER_HACK_DIR)/cherry-pick-pull.sh $(HACK_DIR)/cherry-pick-pull.sh && chmod +xw $(HACK_DIR)/cherry-pick-pull.sh

#################################################
# Rules for starting machine-controller locally
#################################################

.PHONY: start
start:
	go run \
		cmd/machine-controller/main.go \
		--control-kubeconfig=$(CONTROL_KUBECONFIG) \
		--target-kubeconfig=$(TARGET_KUBECONFIG) \
		--namespace=$(CONTROL_NAMESPACE) \
		--machine-creation-timeout=20m \
		--machine-drain-timeout=5m \
		--machine-health-timeout=10m \
		--machine-pv-detach-timeout=2m \
		--machine-safety-apiserver-statuscheck-timeout=30s \
		--machine-safety-apiserver-statuscheck-period=1m \
		--machine-safety-orphan-vms-period=30m \
		--leader-elect=$(LEADER_ELECT) \
		--v=3

#####################################################################
# Rules for verification, formatting, linting, testing and cleaning
#####################################################################

.PHONY: install
install:
	@LD_FLAGS="-w -X github.com/23technologies/$(NAME)/pkg/version.Version=$(VERSION)" \
	bash $(GARDENER_HACK_DIR)/install.sh ./...

.PHONY: generate
generate: $(VGOPATH) $(CONTROLLER_GEN) $(GEN_CRD_API_REFERENCE_DOCS) $(HELM) $(MOCKGEN)
	@REPO_ROOT=$(REPO_ROOT) VGOPATH=$(VGOPATH) GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) bash $(GARDENER_HACK_DIR)/generate-sequential.sh ./cmd/... ./pkg/...
	$(MAKE) format

.PHONY: check-generate
check-generate:
	@bash $(GARDENER_HACK_DIR)/check-generate.sh $(REPO_ROOT)

.PHONY: format
format: $(GOIMPORTS) $(GOIMPORTSREVISER)
	@bash $(GARDENER_HACK_DIR)/format.sh ./cmd ./pkg ./test

.PHONY: sast
sast: $(GOSEC)
	@bash $(GARDENER_HACK_DIR)/sast.sh

.PHONY: sast-report
sast-report: $(GOSEC)
	@bash $(GARDENER_HACK_DIR)/sast.sh --gosec-report true

.PHONY: check
check: $(GO_ADD_LICENSE) $(GOIMPORTS) $(GOLANGCI_LINT)
	@bash $(GARDENER_HACK_DIR)/check.sh --golangci-lint-config=./.golangci.yaml ./cmd/... ./pkg/... ./test/...

	@bash $(GARDENER_HACK_DIR)/check-charts.sh ./charts
	@bash $(GARDENER_HACK_DIR)/check-license-header.sh

.PHONY: test
test:
	@SKIP_FETCH_TOOLS=1 bash $(GARDENER_HACK_DIR)/test.sh ./cmd/... ./pkg/...

.PHONY: test-cov
test-cov:
	@SKIP_FETCH_TOOLS=1 bash $(GARDENER_HACK_DIR)/test-cover.sh ./cmd/... ./pkg/...

.PHONY: test-clean
test-clean:
	@bash $(GARDENER_HACK_DIR)/test-cover-clean.sh

.PHONY: verify
verify: check format test sast

.PHONY: verify-extended
verify-extended: check-generate check format test-cov test-clean sast-report

.PHONY: clean
clean:
	@bash $(GARDENER_HACK_DIR)/clean.sh ./cmd/... ./pkg/...

.PHONY: test-integration
test-integration:
	.ci/local_integration_test

.PHONY: add-license-headers
add-license-headers: $(GO_ADD_LICENSE)
	@bash $(GARDENER_HACK_DIR)/add-license-header.sh
