package svm

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

func TestMCCSVMs(t *testing.T) {
	data := populatedData()
	opts := options.New()
	s := &SVM{AbstractPlugin: plugin.New("svm", opts, nil, nil, "svm", nil)}

	s.updateSVM(data)

	// svm1-mc should be exportable
	instance := data.GetInstance("svm1-mc")
	assert.True(t, instance.IsExportable())

	// svm2-mc should not be exportable
	instance = data.GetInstance("svm2-mc")
	assert.False(t, instance.IsExportable())

	// svm-test should be exportable
	instance = data.GetInstance("svm-test")
	assert.True(t, instance.IsExportable())
}

func populatedData() *matrix.Matrix {
	// Create test data
	data := matrix.New("svm", "svm", "svm")

	instance1, _ := data.NewInstance("svm1-mc")
	instance1.SetLabel("svm", "svm1-mc")
	instance1.SetLabel("state", "running")

	instance2, _ := data.NewInstance("svm2-mc")
	instance2.SetLabel("svm", "svm2-mc")
	instance2.SetLabel("state", "stopped")

	instance3, _ := data.NewInstance("svm-test")
	instance3.SetLabel("svm", "svm-test")
	instance3.SetLabel("state", "stopped")

	return data
}

// The queries the four fetching getters use. A fixture is registered per query,
// so Run can drive all of them offline.
const (
	kerberosQuery         = "api/protocols/nfs/kerberos/interfaces"
	fpolicyQuery          = "api/protocols/fpolicy"
	iscsiServicesQuery    = "api/protocols/san/iscsi/services"
	iscsiCredentialsQuery = "api/protocols/san/iscsi/credentials"
)

// newSVMForRun builds the plugin with Init called for real. Options.IsTest makes
// Init skip client creation, and the registered fixtures make the four REST
// getters read files instead of dialling, so Run works end to end offline.
// This follows the testFilePath idiom from the restperf nic plugin.
func newSVMForRun(t *testing.T, fixtures map[string]string) *SVM {
	t.Helper()

	opts := options.New()
	opts.IsTest = true
	params := node.NewS("SVM")
	p := plugin.New("Rest", opts, params, nil, "svm", nil)

	s := &SVM{AbstractPlugin: p}
	if err := s.Init(conf.Remote{}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	s.testFilePaths = fixtures
	return s
}

// #nosec G101 -- these are fixture file paths, not credentials; gosec matches
// on the "credentials" in the iSCSI endpoint's name
func allFixtures() map[string]string {
	return map[string]string{
		kerberosQuery:         "testdata/kerberos.json",
		fpolicyQuery:          "testdata/fpolicy.json",
		iscsiServicesQuery:    "testdata/iscsi_services.json",
		iscsiCredentialsQuery: "testdata/iscsi_credentials.json",
	}
}

func runSVMMatrix(t *testing.T, names ...string) (*matrix.Matrix, map[string]*matrix.Instance) {
	t.Helper()
	data := matrix.New("svm", "svm", "svm")
	instances := make(map[string]*matrix.Instance, len(names))
	for _, name := range names {
		inst, err := data.NewInstance(name)
		if err != nil {
			t.Fatalf("NewInstance(%s): %v", name, err)
		}
		inst.SetLabel("svm", name)
		inst.SetLabel("state", "running")
		instances[name] = inst
	}
	return data, instances
}

// TestRunAppliesLabelsFromAllEndpoints covers Run end to end: the four REST
// lookups, their merge semantics, and the handoff to updateSVM.
//
// Before the Options.IsTest guard and the fixture seam, Run could not be called
// at all without a cluster, so nothing verified that it wired its five getters
// into updateSVM. This is the package behind GitHub issues #4365, #3122, #4242
// and #2710.
func TestRunAppliesLabelsFromAllEndpoints(t *testing.T) {
	s := newSVMForRun(t, allFixtures())
	data, inst := runSVMMatrix(t, "svm1", "svm2")

	got, meta, err := s.Run(map[string]*matrix.Matrix{"svm": data})

	assert.Nil(t, err)
	assert.Nil(t, got)
	assert.NotNil(t, meta)

	// Kerberos: true if any interface on the SVM has it enabled, in either order.
	assert.Equal(t, inst["svm1"].GetLabel("nfs_kerberos_protocol_enabled"), "true")
	assert.Equal(t, inst["svm2"].GetLabel("nfs_kerberos_protocol_enabled"), "true")

	// Fpolicy: a later record replaces an earlier one only when the earlier was
	// disabled, so svm1 ends up enabled and svm2 keeps its first, enabled policy.
	assert.Equal(t, inst["svm1"].GetLabel("fpolicy_enabled"), "true")
	assert.Equal(t, inst["svm1"].GetLabel("fpolicy_name"), "fp_on")
	assert.Equal(t, inst["svm2"].GetLabel("fpolicy_enabled"), "true")
	assert.Equal(t, inst["svm2"].GetLabel("fpolicy_name"), "fp_first")

	// iSCSI service: same upgrade-only-from-false rule.
	assert.Equal(t, inst["svm1"].GetLabel("iscsi_service_enabled"), "true")
	assert.Equal(t, inst["svm2"].GetLabel("iscsi_service_enabled"), "true")

	// iSCSI credentials: every authentication type is appended, comma separated.
	assert.Equal(t, inst["svm1"].GetLabel("iscsi_authentication_type"), "chap,none")
	assert.Equal(t, inst["svm2"].GetLabel("iscsi_authentication_type"), "deny")
}

// TestRunLeavesUnknownSVMsUnlabelled is the control: an SVM the endpoints never
// mention must not pick up labels from another SVM's response.
func TestRunLeavesUnknownSVMsUnlabelled(t *testing.T) {
	s := newSVMForRun(t, allFixtures())
	data, inst := runSVMMatrix(t, "svm_absent")

	_, _, err := s.Run(map[string]*matrix.Matrix{"svm": data})
	assert.Nil(t, err)

	for _, label := range []string{
		"nfs_kerberos_protocol_enabled", "fpolicy_enabled", "fpolicy_name",
		"iscsi_service_enabled", "iscsi_authentication_type",
	} {
		if got := inst["svm_absent"].GetLabel(label); got != "" {
			t.Errorf("label %q: got %q, want it unset", label, got)
		}
	}
}

// TestRunAppliesMCCExportRule covers the MetroCluster export rule through Run,
// so it is verified as part of the real orchestration and not only in
// updateSVM in isolation.
func TestRunAppliesMCCExportRule(t *testing.T) {
	s := newSVMForRun(t, allFixtures())

	data := matrix.New("svm", "svm", "svm")
	running, err := data.NewInstance("svm1-mc")
	assert.Nil(t, err)
	running.SetLabel("svm", "svm1-mc")
	running.SetLabel("state", "running")

	stopped, err := data.NewInstance("svm2-mc")
	assert.Nil(t, err)
	stopped.SetLabel("svm", "svm2-mc")
	stopped.SetLabel("state", "stopped")

	_, _, err = s.Run(map[string]*matrix.Matrix{"svm": data})
	assert.Nil(t, err)

	assert.True(t, running.IsExportable())
	assert.False(t, stopped.IsExportable())
}

// TestRunToleratesMissingEndpoints pins the error handling: each getter's
// failure is logged and skipped, not propagated, so one unsupported endpoint
// cannot stop the whole plugin. Registering a nonexistent fixture makes every
// fetch fail.
func TestRunToleratesMissingEndpoints(t *testing.T) {
	s := newSVMForRun(t, map[string]string{
		kerberosQuery:         "testdata/does_not_exist.json",
		fpolicyQuery:          "testdata/does_not_exist.json",
		iscsiServicesQuery:    "testdata/does_not_exist.json",
		iscsiCredentialsQuery: "testdata/does_not_exist.json",
	})
	data, inst := runSVMMatrix(t, "svm1")

	got, meta, err := s.Run(map[string]*matrix.Matrix{"svm": data})

	// Run still succeeds, and updateSVM still applied what it could.
	assert.Nil(t, err)
	assert.Nil(t, got)
	assert.NotNil(t, meta)
	assert.Equal(t, inst["svm1"].GetLabel("fpolicy_enabled"), "")
	assert.True(t, inst["svm1"].IsExportable())
}

// TestRunNameserviceSwitchFromLabel covers GetNSSwitchInfo, the one getter that
// makes no REST call: it parses the nameservice_switch label the template
// already collected. The nsdb and nssource lists must be sorted before joining,
// or the label value changes between polls and creates a new time series.
func TestRunNameserviceSwitchFromLabel(t *testing.T) {
	s := newSVMForRun(t, allFixtures())
	data, inst := runSVMMatrix(t, "svm1")

	// Deliberately unsorted, as ONTAP returns it.
	inst["svm1"].SetLabel("nameservice_switch", `{"passwd":["files","ldap"],"group":["nis"],"hosts":["dns"]}`)

	_, _, err := s.Run(map[string]*matrix.Matrix{"svm": data})
	assert.Nil(t, err)

	assert.Equal(t, inst["svm1"].GetLabel("ns_db"), "group,hosts,passwd")
	assert.Equal(t, inst["svm1"].GetLabel("ns_source"), "dns,files,ldap,nis")
}

// TestGetFpolicyParsesPoliciesArray is the regression test for the fpolicy
// array-parsing bug.
//
// The ONTAP REST schema for api/protocols/fpolicy declares "policies" as an
// array of objects, each with a "name" and a boolean "enabled" (verified
// against the published Swagger spec, definition fpolicy_policies, introduced
// in 9.6). The plugin read "policies.enabled" and "policies.name" directly,
// which gjson resolves only when the value is a single object -- so on any real
// response both lookups returned empty and every SVM silently got blank
// fpolicy_enabled and fpolicy_name labels, with no error to indicate it.
func TestGetFpolicyParsesPoliciesArray(t *testing.T) {
	s := newSVMForRun(t, map[string]string{fpolicyQuery: "testdata/fpolicy.json"})

	fpolicy, err := s.GetFpolicy()
	assert.Nil(t, err)

	// svm1's first policy is disabled, so the enabled one replaces it.
	assert.Equal(t, fpolicy["svm1"].enable, "true")
	assert.Equal(t, fpolicy["svm1"].name, "fp_on")

	// svm2's first policy is already enabled, so it is kept.
	assert.Equal(t, fpolicy["svm2"].enable, "true")
	assert.Equal(t, fpolicy["svm2"].name, "fp_first")

	// An SVM with an empty policies array contributes no entry at all, rather
	// than an entry with blank values.
	_, ok := fpolicy["svm3"]
	assert.False(t, ok)
}

// TestGetFpolicyParsesFlattenedPolicies pins that a lone policy serialized as
// an object rather than a one-element array is still parsed. gjson's Array()
// wraps a non-array in a single-element slice, so the fix covers both shapes
// and cannot regress if a response is flattened somewhere in transit.
func TestGetFpolicyParsesFlattenedPolicies(t *testing.T) {
	s := newSVMForRun(t, map[string]string{fpolicyQuery: "testdata/fpolicy_flattened.json"})

	fpolicy, err := s.GetFpolicy()
	assert.Nil(t, err)

	assert.Equal(t, fpolicy["svm1"].enable, "true")
	assert.Equal(t, fpolicy["svm1"].name, "fp_on")
}
