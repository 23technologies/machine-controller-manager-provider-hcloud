# Catch Up to Extension Toolchain Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring `machine-controller-manager-provider-hcloud` up to the exact toolchain and `providerSpec` contract of `opendefensecloud/gardener-extension-provider-hcloud`, migrate to hcloud-go v2, rename the module to `opendefensecloud`, and fold in the P0 correctness fixes — producing a new release the extension can pin.

**Architecture:** The Go source imports **no `gardener/gardener` packages** (verified) — only MCM, hcloud-go, and k8s. So the version catch-up is (a) a build-tooling/go.sum bump for gardener + code-generation, (b) a straight recompile against MCM v0.62 (driver interface unchanged), (c) a real SDK migration hcloud-go v1→v2 where resource IDs change `int`→`int64`, and (d) additive `providerSpec` fields (`tags`, `extraConfig`) plus the P0 bug fixes. Everything below must land together for the module to compile, so it ships as one coordinated branch cut as `v0.4.0-odp.1`.

**Tech Stack:** Go 1.26.2, gardener/gardener v1.146.4 (build tooling only), machine-controller-manager v0.62.1, hetznercloud/hcloud-go/v2 v2.44.0, k8s.io/* v0.35.5, sigs.k8s.io/controller-runtime v0.23.3, onsi/ginkgo/v2 v2.32.0 + gomega v1.42.0, go.uber.org/mock v0.6.0.

## Global Constraints

- Module path MUST become `github.com/opendefensecloud/machine-controller-manager-provider-hcloud` (was `github.com/23technologies/...`). No `23technologies` string may remain in `cmd/`, `pkg/`, `Makefile`, or `Dockerfile` after Task 1.
- Dependency versions are pinned to the EXACT values the extension's `go.mod` uses (listed in Tech Stack). Do not float to newer.
- hcloud-go MUST be the `/v2` module (`github.com/hetznercloud/hcloud-go/v2/hcloud`). No import of the v1 path (`github.com/hetznercloud/hcloud-go/hcloud`) may remain.
- All resource IDs (server, floating IP, placement group) are `int64` after Task 3.
- The MCM driver interface is unchanged v0.58→v0.62: keep `CreateMachine`, `InitializeMachine`, `DeleteMachine`, `GetMachineStatus`, `ListMachines`, `GetVolumeIDs`; keep the `Unimplemented` stubs.
- `providerSpec` fields the provider MUST accept (contract from the extension's `charts/internal/machineclass/templates/machineclass.yaml`): `cluster`, `zone`, `serverType`, `imageName`, `sshFingerprint`, `placementGroupID`, `networkName`, `floatingPoolName`, `tags` (map[string]string), `extraConfig` (map[string]string).
- After release, the extension's `charts/images.yaml` tag `machine-controller-manager-provider-hcloud` (currently `v0.3.2-odp.1`) must be bumped to the new tag — tracked as a handoff item, not part of this repo's tasks.
- Every task ends green: `go build ./...` and `go test ./...` pass before commit.

---

## Task 1: Rename module 23technologies → opendefensecloud

**Files:**
- Modify: `go.mod:1` (module line)
- Modify: all `.go` files under `cmd/` and `pkg/` importing the old path (13 import sites across `pkg/hcloud/*`, `pkg/hcloud/apis/*`, `cmd/machine-controller/main.go`, `pkg/cmd/app.go`)
- Modify: `Makefile:64` (LD_FLAGS `-X` path)
- Modify: `Dockerfile:4` (WORKDIR)

**Interfaces:**
- Produces: the module import prefix `github.com/opendefensecloud/machine-controller-manager-provider-hcloud`, used by every later task's import lines.

- [ ] **Step 1: Rewrite the module path and all references**

```bash
OLD='github.com/23technologies/machine-controller-manager-provider-hcloud'
NEW='github.com/opendefensecloud/machine-controller-manager-provider-hcloud'
go mod edit -module "$NEW"
grep -rl "$OLD" cmd pkg Makefile Dockerfile | xargs sed -i '' "s#$OLD#$NEW#g"
```

- [ ] **Step 2: Verify no stale references remain**

Run: `grep -rn "23technologies" cmd pkg Makefile Dockerfile`
Expected: no output (exit 1).

- [ ] **Step 3: Build to confirm the rename compiles (still on old deps)**

Run: `go build ./...`
Expected: exit 0 (this task changes no dependencies, only the self-import prefix).

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "chore: rename module 23technologies -> opendefensecloud"
```

---

## Task 2: Bump dependencies to the extension's pinned versions

This task is **compile-driven and iterative** — bump, tidy, build, fix. The known likely breakage points are called out; fix them as the compiler surfaces them.

**Files:**
- Modify: `go.mod` (go directive + require versions), `go.sum` (regenerated)
- Modify: `hack/tools.go` (gardener hack tooling moved to submodules in recent gardener; drop deprecated `golang.org/x/lint`)
- Modify: `Dockerfile:2` (builder image → `golang:1.26.2`)

**Interfaces:**
- Produces: go.mod on `go 1.26.2`, gardener v1.146.4, MCM v0.62.1, hcloud-go/v2 v2.44.0, k8s v0.35.5, controller-runtime v0.23.3, ginkgo v2.32.0, gomega v1.42.0, go.uber.org/mock v0.6.0. Task 3 depends on the hcloud-go/v2 module being present.

- [ ] **Step 1: Set the go directive and toolchain**

```bash
go mod edit -go=1.26.2 -toolchain=go1.26.2
```

- [ ] **Step 2: Pin the core dependencies to the extension's versions**

```bash
go mod edit \
  -require=github.com/gardener/gardener@v1.146.4 \
  -require=github.com/gardener/machine-controller-manager@v0.62.1 \
  -require=github.com/hetznercloud/hcloud-go/v2@v2.44.0 \
  -require=k8s.io/api@v0.35.5 \
  -require=k8s.io/apimachinery@v0.35.5 \
  -require=k8s.io/client-go@v0.35.5 \
  -require=k8s.io/code-generator@v0.35.5 \
  -require=k8s.io/component-base@v0.35.5 \
  -require=sigs.k8s.io/controller-runtime@v0.23.3 \
  -require=github.com/onsi/ginkgo/v2@v2.32.0 \
  -require=github.com/onsi/gomega@v1.42.0 \
  -require=go.uber.org/mock@v0.6.0
go mod edit -droprequire=golang.org/x/lint
```

- [ ] **Step 3: Update `hack/tools.go` for the new gardener module layout and drop golint**

In recent gardener the `hack` tooling is consumed via the `github.com/gardener/gardener/hack/tools` submodule and `golang.org/x/lint/golint` is dropped. Replace the file body's import block with:

```go
import (
	_ "github.com/ahmetb/gen-crd-api-reference-docs"
	_ "github.com/gardener/gardener/hack"
	_ "github.com/gardener/gardener/hack/api-reference/template"
	_ "github.com/onsi/ginkgo/v2"
	_ "github.com/onsi/gomega"
	_ "k8s.io/code-generator"
	_ "sigs.k8s.io/controller-runtime"
)
```

> If `go mod tidy` in Step 4 reports that `github.com/gardener/gardener/hack` or `.../hack/.ci` has moved to a submodule (e.g. `github.com/gardener/gardener/hack/tools`), follow the error: add the submodule `require` the extension's go.mod uses (`github.com/gardener/gardener/hack/tools v1.146.4`, `github.com/gardener/gardener/pkg/apis v1.146.4`) and adjust the blank imports to the paths that actually resolve. The extension repo's `go.mod` + `hack/tools.go` are the reference.

- [ ] **Step 4: Tidy and download**

Run: `go mod tidy && go mod download`
Expected: completes; go.sum regenerated. Resolve any "missing go.sum entry" / "module moved" errors by following the tidy output (see note in Step 3).

- [ ] **Step 5: Build — expect hcloud-go v1 breakage only**

Run: `go build ./... 2>&1 | head -40`
Expected: the ONLY errors are in the 5 files still importing the hcloud-go **v1** path (`client.go`, `utils.go`, `machine_controller.go`, `mock/test_env.go`, `mock/machine_controller.go`) — those are fixed in Task 3. There should be no MCM or k8s API breakage (MCM interface unchanged; k8s minor bump is source-compatible for the small surface used in `cmd/`/`pkg/cmd`). If MCM/k8s errors appear, fix them minimally here (e.g. renamed helper in `pkg/cmd/app.go`).

- [ ] **Step 6: Bump the Dockerfile builder image**

In `Dockerfile:2` change `FROM golang:1.24.4 AS builder` to:

```dockerfile
FROM golang:1.26.2 AS builder
```

- [ ] **Step 7: Commit (build still red on hcloud-go — that's expected and noted in the message)**

```bash
git add go.mod go.sum hack/tools.go Dockerfile
git commit -m "chore: bump deps to extension pins (gardener v1.146.4, MCM v0.62.1, k8s v0.35.5, go 1.26.2); hcloud-go v2 migration follows"
```

---

## Task 3: Migrate hcloud-go v1 → v2 (int → int64)

**Files:**
- Modify: `pkg/hcloud/apis/client.go:24`, `pkg/hcloud/apis/utils.go:28-29`, `pkg/hcloud/machine_controller.go:30`, `pkg/hcloud/apis/mock/test_env.go:24`, `pkg/hcloud/apis/mock/machine_controller.go` (imports)
- Modify: `pkg/hcloud/types.go:20-23` (ID field types)
- Modify: `pkg/hcloud/apis/transcoder/provider_id.go` (ID type + string conversions)
- Test: `pkg/hcloud/apis/transcoder/provider_id_test.go` (int64 literals)

**Interfaces:**
- Produces: `transcoder.DecodeServerIDFromProviderID(providerID string) (int64, error)`, `transcoder.EncodeProviderID(zone string, serverID int64) string`, `transcoder.ServerData{ID int64; Zone string}`, `hcloud.CreateMachineMethodData{ServerID int64; FloatingIPID int64}` — consumed by `machine_controller.go` in Tasks 4 & 5.

- [ ] **Step 1: Update the test to expect int64 (write the failing expectation first)**

In `pkg/hcloud/apis/transcoder/provider_id_test.go`, ensure the decode assertions compare against `int64`. Add/adjust:

```go
It("should decode the server ID as int64", func() {
	id, err := DecodeServerIDFromProviderID("hcloud:///hel1-dc2/12345")
	Expect(err).NotTo(HaveOccurred())
	Expect(id).To(Equal(int64(12345)))
})
```

- [ ] **Step 2: Run it to confirm it fails to compile (old signature returns int)**

Run: `go test ./pkg/hcloud/apis/transcoder/... 2>&1 | head`
Expected: compile error — `int64` does not match the current `int` return of `DecodeServerIDFromProviderID`.

- [ ] **Step 3: Rewrite the hcloud-go import path in all 5 files**

```bash
grep -rl 'hetznercloud/hcloud-go/hcloud' pkg | \
  xargs sed -i '' 's#github.com/hetznercloud/hcloud-go/hcloud#github.com/hetznercloud/hcloud-go/v2/hcloud#g'
```

- [ ] **Step 4: Change ID field types to int64 in `pkg/hcloud/types.go`**

```go
type CreateMachineMethodData struct {
	ServerID     int64
	FloatingIPID int64
}
```

- [ ] **Step 5: Update `pkg/hcloud/apis/transcoder/provider_id.go` to int64**

`ServerData.ID` → `int64`; `DecodeServerIDFromProviderID` returns `(int64, error)` with error sentinel `-1`; parse with `ParseInt`; format with `FormatInt`; `EncodeProviderID` takes `int64`:

```go
// ServerData describes a server
type ServerData struct {
	ID   int64
	Zone string
}
```

In `DecodeServerDataFromProviderID`, replace the `strconv.Atoi` block:

```go
	serverID, err := strconv.ParseInt(providerIDData[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("ServerID found is invalid: %v", err)
	}
```

`DecodeServerIDFromProviderID`:

```go
func DecodeServerIDFromProviderID(providerID string) (int64, error) {
	serverData, err := DecodeServerDataFromProviderID(providerID)
	if err != nil {
		return -1, err
	}
	return serverData.ID, nil
}
```

`DecodeServerIDAsStringFromProviderID` conversion:

```go
	return strconv.FormatInt(serverData.ID, 10), nil
```

`EncodeProviderID`:

```go
func EncodeProviderID(zone string, serverID int64) string {
	return fmt.Sprintf("hcloud:///%s/%d", url.PathEscape(zone), serverID)
}
```

- [ ] **Step 6: Fix remaining int64 flow-through in `machine_controller.go`**

Build and fix each site the compiler flags. Known sites: the `serverID` local in `GetMachineStatus` (declared via `transcoder.DecodeServerIDFromProviderID`, now int64 — remove the redundant `int` declaration at the top `var serverID int` and let it be int64), the `PlacementGroupID` parse (`strconv.Atoi` → `strconv.ParseInt(providerSpec.PlacementGroupID, 10, 64)` returning int64 for `PlacementGroup.GetByID`), and `resultData.ServerID`/`FloatingIPID` assignments (now int64 from `serverResult.Server.ID` / `ipResult.FloatingIP.ID`, which are int64 in v2).

- [ ] **Step 7: Build the whole module**

Run: `go build ./... 2>&1 | head -40`
Expected: exit 0. If the mock package fails, fix `mock/test_env.go` / `mock/machine_controller.go` ID literals to int64 and any v2-renamed helper (the `datacenter` field was already removed upstream — see commit `3e2b178` — so the mock should not reference it).

- [ ] **Step 8: Run the full test suite**

Run: `go test ./... 2>&1 | tail -30`
Expected: PASS. Fix any int/int64 literal mismatches in `*_test.go` and mock handlers.

- [ ] **Step 9: Commit**

```bash
git add -A
git commit -m "feat: migrate to hcloud-go v2 (resource IDs int -> int64)"
```

---

## Task 4: Add `tags` and `extraConfig` to ProviderSpec and apply tags to server labels

**Files:**
- Modify: `pkg/hcloud/apis/provider_spec.go` (add fields)
- Modify: `pkg/hcloud/machine_controller.go:127-140` (merge `Tags` into `ServerCreateOpts.Labels`)
- Test: `pkg/hcloud/apis/transcoder/machine_class_json_test.go` (decode tags/extraConfig)

**Interfaces:**
- Consumes: `apis.ProviderSpec` from Task 1.
- Produces: `apis.ProviderSpec.Tags map[string]string` and `apis.ProviderSpec.ExtraConfig map[string]string`, consumed by `CreateMachine`.

- [ ] **Step 1: Write the failing decode test**

In `pkg/hcloud/apis/transcoder/machine_class_json_test.go`, add a case that decodes a providerSpec carrying `tags` and asserts they survive. Use the mock provider spec builder or inline JSON:

```go
It("should decode tags and extraConfig", func() {
	raw := []byte(`{"cluster":"c","zone":"hel1-dc2","serverType":"cx22","imageName":"ubuntu-24.04","sshFingerprint":"aa:bb","tags":{"mcm.gardener.cloud/cluster":"c","mcm.gardener.cloud/role":"node"},"extraConfig":{"foo":"bar"}}`)
	mc := &v1alpha1.MachineClass{ProviderSpec: runtime.RawExtension{Raw: raw}}
	spec, err := DecodeProviderSpecFromMachineClass(mc, providerSecret)
	Expect(err).NotTo(HaveOccurred())
	Expect(spec.Tags).To(HaveKeyWithValue("mcm.gardener.cloud/role", "node"))
	Expect(spec.ExtraConfig).To(HaveKeyWithValue("foo", "bar"))
})
```

- [ ] **Step 2: Run it to confirm it fails**

Run: `go test ./pkg/hcloud/apis/transcoder/... -run Decode 2>&1 | head`
Expected: compile error — `spec.Tags` / `spec.ExtraConfig` undefined.

- [ ] **Step 3: Add the fields to `pkg/hcloud/apis/provider_spec.go`**

```go
type ProviderSpec struct {
	Cluster        string `json:"cluster"`
	Zone           string `json:"zone"`
	ServerType     string `json:"serverType"`
	ImageName      string `json:"imageName"`
	SSHFingerprint string `json:"sshFingerprint"`

	PlacementGroupID string `json:"placementGroupID,omitempty"`
	FloatingPoolName string `json:"floatingPoolName,omitempty"`
	NetworkName      string `json:"networkName,omitempty"`

	// Tags are additional labels applied to the created Hetzner server. The extension
	// supplies mcm.gardener.cloud/cluster and mcm.gardener.cloud/role here.
	Tags map[string]string `json:"tags,omitempty"`
	// ExtraConfig carries per-machine-type "additional VM options" from the CloudProfile's
	// MachineTypeOptions. Reserved: currently accepted but not applied (no Hetzner semantic yet).
	ExtraConfig map[string]string `json:"extraConfig,omitempty"`
}
```

- [ ] **Step 4: Run the decode test to confirm it passes**

Run: `go test ./pkg/hcloud/apis/transcoder/... -run Decode -v 2>&1 | tail`
Expected: PASS.

- [ ] **Step 5: Merge `Tags` into the server labels in `CreateMachine`**

In `pkg/hcloud/machine_controller.go`, after the `opts` literal (which sets the topology + cluster/role labels), merge provider-supplied tags so they augment (and can override) the defaults, without dropping the keys `ListMachines` selects on:

```go
	for k, v := range providerSpec.Tags {
		opts.Labels[k] = v
	}
```

Insert this immediately after the `opts := hcloud.ServerCreateOpts{...}` block (after line ~140) and before the SSH-key lookup.

- [ ] **Step 6: Build and test**

Run: `go build ./... && go test ./... 2>&1 | tail -20`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: accept providerSpec.tags/extraConfig; apply tags to server labels (extension contract)"
```

---

## Task 5: P0 correctness fixes

Five independent fixes. Each is test-first where a unit test is feasible; the placement-group and cleanup fixes are exercised via the mock.

### Task 5a: Decouple placement group from FloatingPoolName

**Files:** Modify `pkg/hcloud/machine_controller.go:174-194`. Test: `pkg/hcloud/machine_controller_test.go`.

- [ ] **Step 1: Fix the gate and the parse**

Change the placement-group block guard from `if providerSpec.FloatingPoolName != ""` to `if providerSpec.PlacementGroupID != ""`, and parse the ID as int64:

```go
	if providerSpec.PlacementGroupID != "" {
		placementGroupID, err := strconv.ParseInt(providerSpec.PlacementGroupID, 10, 64)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		placementGroup, _, err := client.PlacementGroup.GetByID(ctx, placementGroupID)
		if err != nil {
			return nil, status.Error(codes.Unavailable, err.Error())
		}

		_, _, err = client.Server.AddToPlacementGroup(ctx, server, placementGroup)
		if err != nil {
			return nil, status.Error(codes.Unavailable, err.Error())
		}

		server, err = apis.WaitForActionsAndGetServer(ctx, client, server)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}
```

- [ ] **Step 2: Add a CreateMachine test that sets PlacementGroupID (no floating pool)**

Extend the create test / mock so the provider spec has `PlacementGroupID: "42"` and no floating pool, and assert `CreateMachine` succeeds and the mock's `/placement_groups/42` + `AddToPlacementGroup` handler was hit (the mock scaffolding already exists at `mock/machine_controller.go` ~L401-477). Follow the existing test style in `machine_controller_test.go`.

- [ ] **Step 3: Run and confirm pass**

Run: `go test ./pkg/hcloud/... -run CreateMachine 2>&1 | tail`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "fix: apply placement group based on PlacementGroupID, not FloatingPoolName"
```

### Task 5b: Never delete a foreign server in create-error cleanup

**Files:** Modify `pkg/hcloud/machine_controller.go:331-351` (`createMachineOnErrorCleanup`).

- [ ] **Step 1: Delete only servers this request created**

Remove the name-based fallback so cleanup acts solely on the ID this request recorded:

```go
func (p *MachineProvider) createMachineOnErrorCleanup(ctx context.Context, req *driver.CreateMachineRequest) {
	client := apis.GetClientForToken(string(req.Secret.Data["token"]))
	resultData := ctx.Value(CtxWrapDataKey("MethodData")).(*CreateMachineMethodData)

	if resultData.ServerID != 0 {
		server, _, _ := client.Server.GetByID(ctx, resultData.ServerID)
		if nil != server {
			_, _ = client.Server.Delete(ctx, server)
		}
	}

	if resultData.FloatingIPID != 0 {
		floatingIP, _, _ := client.FloatingIP.GetByID(ctx, resultData.FloatingIPID)
		if nil != floatingIP {
			_, _ = client.FloatingIP.Delete(ctx, floatingIP)
		}
	}
}
```

- [ ] **Step 2: Build and run existing tests**

Run: `go test ./pkg/hcloud/... 2>&1 | tail`
Expected: PASS (existing cleanup tests still pass; foreign-server name collision no longer deletes).

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "fix: cleanup only deletes servers this request created (never by name)"
```

### Task 5c: Check the GetByName error before create

**Files:** Modify `pkg/hcloud/machine_controller.go:87`.

- [ ] **Step 1: Handle the lookup error**

```go
	server, _, err := client.Server.GetByName(ctx, machine.Name)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
```

(Change the `:=` on the existing `providerSpec, err :=` flow as needed so `err` is reused, not shadowed; the function already has `err` in scope from the `DecodeProviderSpecFromMachineClass` call above.)

- [ ] **Step 2: Build and test**

Run: `go build ./... && go test ./pkg/hcloud/... 2>&1 | tail`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "fix: fail create on transient Server.GetByName error instead of ignoring it"
```

### Task 5d: Fix nil-panic on null providerSpec + re-enable secret validation

**Files:** Modify `pkg/hcloud/apis/transcoder/machine_class_json.go:39-48`, `pkg/hcloud/apis/validation/validation.go`. Test: `pkg/hcloud/apis/validation/validation_test.go`, `pkg/hcloud/apis/transcoder/machine_class_json_test.go`.

- [ ] **Step 1: Write failing tests**

In `validation_test.go`, add a case: a spec that is valid but a secret missing `token` yields an error. In `machine_class_json_test.go`, add: `ProviderSpec.Raw = []byte("null")` returns an error (not a panic).

```go
It("should reject a null providerSpec without panicking", func() {
	mc := &v1alpha1.MachineClass{ProviderSpec: runtime.RawExtension{Raw: []byte("null")}}
	_, err := DecodeProviderSpecFromMachineClass(mc, providerSecret)
	Expect(err).To(HaveOccurred())
})
```

- [ ] **Step 2: Run to confirm failure (panic / missing validation)**

Run: `go test ./pkg/hcloud/apis/... 2>&1 | head`
Expected: the null case panics or the secret case has no error.

- [ ] **Step 3: Decode into a value + nil-guard in `machine_class_json.go`**

```go
	// Extract providerSpec
	var providerSpec apis.ProviderSpec

	if machineClass == nil {
		return nil, errors.New("MachineClass provided is nil")
	}

	jsonErr := json.Unmarshal(machineClass.ProviderSpec.Raw, &providerSpec)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to parse JSON data provided as ProviderSpec: %v", jsonErr)
	}

	validationErr := validation.ValidateHCloudProviderSpec(&providerSpec, secret)
	if len(validationErr) > 0 {
		return nil, fmt.Errorf("error while validating ProviderSpec %v", validationErr)
	}

	return &providerSpec, nil
```

(Decoding `null` into a value struct leaves it zero-valued, and the required-field validation then rejects it cleanly. Also switch the guard from `!= nil` to `len(...) > 0`.)

- [ ] **Step 4: Re-enable secret validation in `validation.go`**

Replace the commented line with real checks and add a `PlacementGroupID` numeric check:

```go
	if spec.SSHFingerprint == "" {
		allErrs = append(allErrs, fmt.Errorf("sshFingerprint is a required field"))
	}
	if spec.PlacementGroupID != "" {
		if _, err := strconv.ParseInt(spec.PlacementGroupID, 10, 64); err != nil {
			allErrs = append(allErrs, fmt.Errorf("placementGroupID must be numeric: %v", err))
		}
	}
	if secrets != nil {
		if len(secrets.Data["token"]) == 0 {
			allErrs = append(allErrs, fmt.Errorf("token is required in the secret"))
		}
		if len(secrets.Data["userData"]) == 0 {
			allErrs = append(allErrs, fmt.Errorf("userData is required in the secret"))
		}
	}

	return allErrs
```

Add `"strconv"` to the imports.

- [ ] **Step 5: Run tests**

Run: `go test ./pkg/hcloud/apis/... 2>&1 | tail`
Expected: PASS. (Note: existing tests that call `DecodeProviderSpecFromMachineClass` with a valid spec must also supply a secret with `token`+`userData` — update fixtures in `mock/provider_spec.go` / test setup if they now fail.)

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "fix: reject null providerSpec without panic; validate secret token/userData and numeric placementGroupID"
```

### Task 5e: Map transient API errors to retryable codes

**Files:** Modify `pkg/hcloud/machine_controller.go` (lookup error sites: image/sshkey/network/server-by-id lookups at ~107,113,144,154,383,385,391).

- [ ] **Step 1: Reclassify lookup/API failures**

For each `client.*.Get*`/lookup call whose `err` currently returns `codes.InvalidArgument`, change transient API failures to `codes.Unavailable` (retryable), keeping `InvalidArgument` only for genuinely malformed input (e.g. an unparseable providerID, image-not-found which is a config error stays `InvalidArgument`). Concretely: image lookup API error → `Unavailable`; SSH key lookup API error → `Unavailable`; network lookup API error → `Unavailable`; `Server.GetByID` API errors in `DeleteMachine`/`GetMachineStatus` → `Unavailable`. Leave the "not found / nil result" cases as their current explicit codes.

- [ ] **Step 2: Build and test**

Run: `go build ./... && go test ./... 2>&1 | tail`
Expected: PASS (update any test asserting the old code to expect `Unavailable`).

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "fix: return retryable codes.Unavailable for transient hcloud API lookup failures"
```

---

## Task 6: Version package + version stamping

**Files:**
- Create: `pkg/version/version.go`
- Verify: `Makefile:64` `-X` path now resolves (it already targets `.../pkg/version.Version`)

**Interfaces:**
- Produces: `version.Version` string var, set at build time via `-ldflags -X`.

- [ ] **Step 1: Create the version package**

`pkg/version/version.go`:

```go
// Package version holds the build-time version of the provider binary.
package version

// Version is set at build time via -ldflags "-X .../pkg/version.Version=<v>".
var Version = "dev"
```

- [ ] **Step 2: Confirm the LD_FLAGS path matches**

Run: `grep -n 'pkg/version.Version' Makefile`
Expected: shows `-X github.com/opendefensecloud/machine-controller-manager-provider-hcloud/pkg/version.Version=$(VERSION)` (opendefensecloud after Task 1).

- [ ] **Step 3: Build with the stamp and verify it embeds**

Run: `make install && go version -m $(go env GOPATH)/bin/machine-controller 2>/dev/null | grep -i version || echo "built"`
Expected: build succeeds; `VERSION` is embedded (no longer a silent no-op).

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "fix: add pkg/version so -ldflags version stamping is no longer a no-op"
```

---

## Task 7: CI hardening and release-workflow modernization

**Files:**
- Modify: `.github/workflows/on-pr-test.yml`, `.github/workflows/on-release-publish.yml`
- Create: `.github/dependabot.yml`

**Interfaces:** none (CI only).

- [ ] **Step 1: Harden the PR workflow**

Rewrite `.github/workflows/on-pr-test.yml` to Go 1.26, current actions, and add build + vet + lint gates alongside tests:

```yaml
name: go:test

on:
  pull_request:
    paths-ignore:
      - "docs/**"

permissions:
  contents: read

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
          cache: true
      - name: Build
        run: go build ./...
      - name: Vet
        run: go vet ./...
      - name: Test
        run: go test ./... -race
```

> A full `make check` (golangci-lint via gardener hack) + `make sast` gate is desirable but depends on the gardener hack scripts resolving `./charts` (`check-charts.sh`) which this repo lacks — track that as follow-up (Task 8 note). The `go build/vet/test -race` gate above is the immediate win.

- [ ] **Step 2: Modernize the release workflow**

In `.github/workflows/on-release-publish.yml`: bump `actions/checkout@v2`→`@v4`, `docker/setup-buildx-action@v1`→`@v3`, `docker/login-action@v1`→`@v3`, `docker/build-push-action@v2`→`@v6`, `softprops/action-gh-release@v1`→`@v2`; add `platforms: linux/amd64,linux/arm64` to the build-push step; add `needs: create-release` to `on-publish` so a release never publishes without its image.

- [ ] **Step 3: Add dependabot**

`.github/dependabot.yml`:

```yaml
version: 2
updates:
  - package-ecosystem: gomod
    directory: "/"
    schedule: { interval: weekly }
  - package-ecosystem: github-actions
    directory: "/"
    schedule: { interval: weekly }
  - package-ecosystem: docker
    directory: "/"
    schedule: { interval: weekly }
```

- [ ] **Step 4: Commit**

```bash
git add .github
git commit -m "ci: build/vet/test -race gate on PRs; bump actions; multi-arch release; dependabot"
```

---

## Task 8: Full verification and release cut

**Files:** Modify `VERSION`.

- [ ] **Step 1: Run the full local verification the same way CI/Make will**

Run: `go build ./... && go vet ./... && go test ./... -race 2>&1 | tail -30`
Expected: all PASS.

- [ ] **Step 2: Build the container image locally (amd64) to confirm the Dockerfile**

Run: `docker build -t mcm-provider-hcloud:local .`
Expected: image builds; `make install` runs inside on Go 1.26.2.

- [ ] **Step 3: Bump VERSION**

Set `VERSION` to `v0.4.0-odp.1`.

- [ ] **Step 4: Commit and tag**

```bash
git add VERSION
git commit -m "release: v0.4.0-odp.1 (extension toolchain catch-up + hcloud-go v2 + P0 fixes)"
```

(Tag `v0.4.0-odp.1` is pushed by the maintainer to trigger the release workflow — do not push tags without confirmation.)

- [ ] **Step 5: Record the handoff item**

The extension repo's `charts/images.yaml` entry `machine-controller-manager-provider-hcloud` (tag `v0.3.2-odp.1`) must be bumped to `v0.4.0-odp.1` after the image publishes. Renovate will also propose this. Note it in the PR description.

---

## Deferred to follow-up plans (out of scope here)

These are independent, lower-risk, and not required for "compiles + tests pass + matches the extension contract." Track separately:

- **Sample manifests** (`kubernetes/*.yaml`): they deploy the AWS provider image + retired `cx11`/`ubuntu-20.04`; rewrite to the real image + `cx22`/`ubuntu-24.04` with placement-group/floating-pool/network examples.
- **Docs & governance:** `docs/README.md` lists a non-existent interface method; README "work-in-progress" banner; `CODEOWNERS` → a real opendefensecloud team; drop orphaned `.ci/hack` Concourse scripts and the dead `make test-integration` target.
- **hcloud client cache + race** (`apis/client.go`): never writes back to `singletons`; add `sync.RWMutex`/`sync.Map` keyed on a token hash.
- **Broader test coverage:** floating-IP/private-network create paths, both cleanup branches, `apis` package (`GetRegionFromZone`, token newline strip, `HCLOUD_ENDPOINT`), response-payload assertions.
- **Full `make check`/`make sast` CI gate** once the `./charts` dependency of the gardener hack scripts is resolved.

## Self-Review

- **Spec coverage:** match extension versions (Task 2) ✓; module rename (Task 1) ✓; verify + implement tags/extraConfig schema (Task 4) ✓; P0 fixes folded post-revendor (Task 5, after Tasks 2–3) ✓; hcloud-go v2 (Task 3) ✓; CI (Task 7) ✓; release/handoff (Task 8) ✓.
- **Placeholder scan:** dependency-bump/hcloud-v2 tasks are explicitly compile-driven with named likely-breakage points rather than vague "fix errors"; all deterministic edits show exact code.
- **Type consistency:** `DecodeServerIDFromProviderID` int64 (Task 3) matches its use in `machine_controller.go`; `ProviderSpec.Tags`/`ExtraConfig` map[string]string (Task 4) match the extension's `MachineTypeOptions.ExtraConfig` and machineclass `tags` types; `validation` returns `[]error` checked with `len(...) > 0` (Task 5d).
