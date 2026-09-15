package svm

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

// newSVM builds the plugin without Init, which would create a ZAPI client. The
// label logic under test reads only the plugin's own maps, so the client stays
// nil. This mirrors cmd/collectors/rest/plugins/svm/svm_test.go.
func newSVM() *SVM {
	return &SVM{AbstractPlugin: plugin.New("svm", options.New(), nil, nil, "svm", nil)}
}

func newSVMInstance(t *testing.T, data *matrix.Matrix, name string, state string) *matrix.Instance {
	t.Helper()
	i, err := data.NewInstance(name)
	if err != nil {
		t.Fatalf("NewInstance(%s): %v", name, err)
	}
	i.SetLabel("svm", name)
	i.SetLabel("state", state)
	return i
}

// TestMCCSVMs covers the MetroCluster export rule, which is identical to the
// REST plugin's and was the subject of a dedicated fix: an SVM whose name ends
// in "-mc" belongs to the partner cluster and is only exported while running.
// This is the ZAPI side of the behaviour GitHub issue #4420 reported for
// KeyPerf.
func TestMCCSVMs(t *testing.T) {
	data := matrix.New("svm", "svm", "svm")
	running := newSVMInstance(t, data, "svm1-mc", "running")
	stopped := newSVMInstance(t, data, "svm2-mc", "stopped")
	plain := newSVMInstance(t, data, "svm-test", "stopped")

	newSVM().updateSVM(data)

	assert.True(t, running.IsExportable())
	assert.False(t, stopped.IsExportable())
	// A non-MetroCluster SVM is exported regardless of state.
	assert.True(t, plain.IsExportable())
}

// TestAuditProtocolEnabledLabel covers GitHub issue #4365
// "Missing label `audit_protocol_enabled` from `svm_labels` metric".
//
// This label is set unconditionally from the map, so an SVM absent from the
// audit response gets an empty value rather than no label at all -- which is
// what keeps the label present on every svm_labels series.
func TestAuditProtocolEnabledLabel(t *testing.T) {
	data := matrix.New("svm", "svm", "svm")
	withAudit := newSVMInstance(t, data, "svm1", "running")
	withoutAudit := newSVMInstance(t, data, "svm2", "running")

	s := newSVM()
	s.auditProtocols = map[string]string{"svm1": "true"}

	s.updateSVM(data)

	assert.Equal(t, withAudit.GetLabel("audit_protocol_enabled"), "true")
	assert.Equal(t, withoutAudit.GetLabel("audit_protocol_enabled"), "")
}

// TestCifsAndProtocolLabels covers the per-SVM protocol labels that make up
// svm_labels, the metric behind GitHub issue #3122 (cluster svm_labels missing)
// and #2710 (the ZAPI collector reporting extra svm_labels).
func TestCifsAndProtocolLabels(t *testing.T) {
	data := matrix.New("svm", "svm", "svm")
	inst := newSVMInstance(t, data, "svm1", "running")

	s := newSVM()
	s.cifsProtocols = map[string]CifsSecurity{
		"svm1": {cifsNtlmEnabled: "true", smbEncryption: "false", smbSigning: "true"},
	}
	s.cifsEnabled = map[string]bool{"svm1": true}
	s.nfsEnabled = map[string]string{"svm1": "true"}
	s.nisInfo = map[string]string{"svm1": "nis.example.com"}
	s.iscsiAuth = map[string]string{"svm1": "chap"}
	s.iscsiService = map[string]string{"svm1": "true"}
	s.fpolicyData = map[string]Fpolicy{"svm1": {name: "fp1", enable: "true"}}
	s.ldapData = map[string]string{"svm1": "sign"}
	s.kerberosConfig = map[string]string{"svm1": "true"}
	s.sshData = map[string]SSHInfo{"svm1": {ciphers: "aes256", isInsecure: "false"}}

	s.updateSVM(data)

	assert.Equal(t, inst.GetLabel("cifs_ntlm_enabled"), "true")
	assert.Equal(t, inst.GetLabel("smb_encryption_required"), "false")
	assert.Equal(t, inst.GetLabel("smb_signing_required"), "true")
	assert.Equal(t, inst.GetLabel("cifs_protocol_enabled"), "true")
	assert.Equal(t, inst.GetLabel("nfs_protocol_enabled"), "true")
	assert.Equal(t, inst.GetLabel("nis_domain"), "nis.example.com")
	assert.Equal(t, inst.GetLabel("iscsi_authentication_type"), "chap")
	assert.Equal(t, inst.GetLabel("iscsi_service_enabled"), "true")
	assert.Equal(t, inst.GetLabel("fpolicy_enabled"), "true")
	assert.Equal(t, inst.GetLabel("fpolicy_name"), "fp1")
	assert.Equal(t, inst.GetLabel("ldap_session_security"), "sign")
	assert.Equal(t, inst.GetLabel("nfs_kerberos_protocol_enabled"), "true")
	assert.Equal(t, inst.GetLabel("ciphers"), "aes256")
	assert.Equal(t, inst.GetLabel("insecured"), "false")
}

// TestAbsentMapsLeaveLabelsUnset is the control for the test above: every label
// except audit_protocol_enabled is behind a map lookup, so an SVM missing from
// a response must not acquire an empty label. Setting one would publish a label
// the ZAPI collector never collected, which is the #2710 complaint.
func TestAbsentMapsLeaveLabelsUnset(t *testing.T) {
	data := matrix.New("svm", "svm", "svm")
	inst := newSVMInstance(t, data, "svm1", "running")

	newSVM().updateSVM(data)

	for _, label := range []string{
		"cifs_ntlm_enabled", "smb_encryption_required", "smb_signing_required",
		"cifs_protocol_enabled", "nfs_protocol_enabled",
		"iscsi_authentication_type", "iscsi_service_enabled",
		"fpolicy_enabled", "fpolicy_name", "ldap_session_security",
		"nfs_kerberos_protocol_enabled", "ciphers", "insecured",
		"ns_source", "ns_db",
	} {
		if got := inst.GetLabel(label); got != "" {
			t.Errorf("label %q: got %q, want it unset", label, got)
		}
	}
}

// TestNameserviceSwitchLabels pins that the nsdb and nssource lists are sorted
// before being joined, so the label value is stable across polls. An unstable
// label value creates a new time series on every poll.
func TestNameserviceSwitchLabels(t *testing.T) {
	data := matrix.New("svm", "svm", "svm")
	inst := newSVMInstance(t, data, "svm1", "running")

	s := newSVM()
	// Deliberately out of order.
	s.nsswitchInfo = map[string]Nsswitch{
		"svm1": {nsdb: []string{"passwd", "group", "hosts"}, nssource: []string{"nis", "files", "ldap"}},
	}
	s.nisInfo = map[string]string{"svm1": "nis.example.com"}

	s.updateSVM(data)

	assert.Equal(t, inst.GetLabel("ns_db"), "group,hosts,passwd")
	assert.Equal(t, inst.GetLabel("ns_source"), "files,ldap,nis")
}

// TestNonExportableInstanceSkipped pins that an SVM already filtered out by the
// template is not given labels, so it cannot reappear downstream.
func TestNonExportableInstanceSkipped(t *testing.T) {
	data := matrix.New("svm", "svm", "svm")
	inst := newSVMInstance(t, data, "svm1", "running")
	inst.SetExportable(false)

	s := newSVM()
	s.auditProtocols = map[string]string{"svm1": "true"}

	s.updateSVM(data)

	assert.Equal(t, inst.GetLabel("audit_protocol_enabled"), "")
}

// TestStoppedMCSVMGetsNoLabels pins that a partner-cluster SVM that is not
// running is skipped before any label is applied, not merely marked
// non-exportable afterwards.
func TestStoppedMCSVMGetsNoLabels(t *testing.T) {
	data := matrix.New("svm", "svm", "svm")
	inst := newSVMInstance(t, data, "svm1-mc", "stopped")

	s := newSVM()
	s.auditProtocols = map[string]string{"svm1-mc": "true"}
	s.nisInfo = map[string]string{"svm1-mc": "nis.example.com"}

	s.updateSVM(data)

	assert.False(t, inst.IsExportable())
	assert.Equal(t, inst.GetLabel("audit_protocol_enabled"), "")
	assert.Equal(t, inst.GetLabel("nis_domain"), "")
}

// newSVMForRun builds a plugin whose Run can be called without a cluster.
//
// Run's twelve ZAPI lookups all sit behind the cache gate
// (currentVal >= PluginInvocationRate), so leaving currentVal below the rate
// skips every one of them. The only other client use is
// s.client.Metadata.Reset(), which zapi.NewTestClient satisfies with a non-nil
// Metadata and no network. This is the same gate the certificate plugin tests
// use to reach Run offline.
func newSVMForRun(t *testing.T) *SVM {
	t.Helper()
	params := node.NewS("SVM")
	p := plugin.New("Zapi", options.New(), params, nil, "svm", nil)
	s := &SVM{AbstractPlugin: p}
	if err := s.InitAbc(); err != nil {
		t.Fatalf("InitAbc: %v", err)
	}
	s.client = zapi.NewTestClient()
	s.currentVal = 0
	s.PluginInvocationRate = 10
	return s
}

// TestRunAppliesLabels covers the wiring from Run to updateSVM.
//
// Without this, updateSVM could be correct in isolation while Run no longer
// called it: replacing the s.updateSVM(data) line with any other use of data
// left every other test in this file green.
func TestRunAppliesLabels(t *testing.T) {
	s := newSVMForRun(t)
	s.auditProtocols = map[string]string{"svm1": "true"}
	s.nisInfo = map[string]string{"svm1": "nis.example.com"}

	data := matrix.New("svm", "svm", "svm")
	inst := newSVMInstance(t, data, "svm1", "running")

	got, meta, err := s.Run(map[string]*matrix.Matrix{"svm": data})

	assert.Nil(t, err)
	assert.Nil(t, got)
	// Run returns the client's metadata, not nil.
	assert.NotNil(t, meta)

	// The labels prove updateSVM actually ran as part of Run.
	assert.Equal(t, inst.GetLabel("audit_protocol_enabled"), "true")
	assert.Equal(t, inst.GetLabel("nis_domain"), "nis.example.com")
}

// TestRunAppliesMCCExportRule covers the MetroCluster rule through Run, since
// that is the behaviour with the filed-bug history and it must survive the
// orchestration, not just updateSVM in isolation.
func TestRunAppliesMCCExportRule(t *testing.T) {
	s := newSVMForRun(t)

	data := matrix.New("svm", "svm", "svm")
	running := newSVMInstance(t, data, "svm1-mc", "running")
	stopped := newSVMInstance(t, data, "svm2-mc", "stopped")

	_, _, err := s.Run(map[string]*matrix.Matrix{"svm": data})
	assert.Nil(t, err)

	assert.True(t, running.IsExportable())
	assert.False(t, stopped.IsExportable())
}

// TestRunCacheGating pins that the expensive ZAPI lookups stay behind the
// invocation-rate gate. Run increments currentVal each poll and only refreshes
// once the rate is reached -- which is also what keeps this test offline, so a
// regression that removed the gate would fail here rather than silently
// issuing twelve ZAPI calls per poll.
func TestRunCacheGating(t *testing.T) {
	s := newSVMForRun(t)

	data := matrix.New("svm", "svm", "svm")
	newSVMInstance(t, data, "svm1", "running")
	dataMap := map[string]*matrix.Matrix{"svm": data}

	for i := range 5 {
		_, _, err := s.Run(dataMap)
		assert.Nil(t, err)
		assert.Equal(t, s.currentVal, i+1)
	}
}
