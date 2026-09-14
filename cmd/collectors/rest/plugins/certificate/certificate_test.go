package certificate

import (
	"testing"
	"time"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

// newCertificate builds the plugin without calling Init, because Init creates a
// REST client unconditionally. currentVal is left below PluginInvocationRate so
// Run skips refreshAdminSerial and never touches the client -- which is what
// makes the rest of Run testable offline.
func newCertificate(t *testing.T, adminSerial string) *Certificate {
	t.Helper()
	params := node.NewS("Certificate")
	p := plugin.New("Rest", options.New(), params, nil, "certificate", nil)
	c := &Certificate{AbstractPlugin: p}
	if err := c.InitAbc(); err != nil {
		t.Fatalf("InitAbc: %v", err)
	}
	c.adminVserverSerial = adminSerial
	c.currentVal = 0
	c.PluginInvocationRate = 10
	return c
}

func certMatrix(t *testing.T) (*matrix.Matrix, *matrix.Metric) {
	t.Helper()
	data := matrix.New("Certificate", "certificate", "certificate")
	expiration, err := data.NewMetricFloat64("expiration")
	if err != nil {
		t.Fatalf("NewMetricFloat64: %v", err)
	}
	return data, expiration
}

func newCertInstance(t *testing.T, data *matrix.Matrix, labels map[string]string) *matrix.Instance {
	t.Helper()
	i, err := data.NewInstance("cert-1")
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}
	for k, v := range labels {
		i.SetLabel(k, v)
	}
	return i
}

// TestRunSetsExpiryTime covers the RFC3339 expiry_time label, which the
// certificate dashboard renders directly.
func TestRunSetsExpiryTime(t *testing.T) {
	c := newCertificate(t, "")
	data, expiration := certMatrix(t)

	inst := newCertInstance(t, data, map[string]string{
		"serial_number": "ABC123", "type": "server",
	})
	// 2030-01-01T00:00:00Z
	expiration.SetValueFloat64(inst, 1893456000)

	_, meta, err := c.Run(map[string]*matrix.Matrix{"certificate": data})

	assert.Nil(t, err)
	assert.NotNil(t, meta)
	assert.Equal(t, inst.GetLabel("expiry_time"), "2030-01-01T00:00:00Z")
}

// TestRunWithoutAdminSerialLeavesIssuerUnset covers GitHub issue #3398
// "SecurityCert plugin error", which shipped with the status/untested label.
//
// When the admin SVM serial could not be resolved, adminVserverSerial is empty.
// The plugin must then skip the issuer and validity classification rather than
// attributing them to an arbitrary certificate.
func TestRunWithoutAdminSerialLeavesIssuerUnset(t *testing.T) {
	c := newCertificate(t, "")
	data, expiration := certMatrix(t)

	inst := newCertInstance(t, data, map[string]string{
		"serial_number": "ABC123", "type": "server",
	})
	expiration.SetValueFloat64(inst, 1893456000)

	_, _, err := c.Run(map[string]*matrix.Matrix{"certificate": data})
	assert.Nil(t, err)

	// expiry_time is still set -- it does not depend on the admin SVM.
	assert.Equal(t, inst.GetLabel("expiry_time"), "2030-01-01T00:00:00Z")
	// But the admin-SVM-only labels must not be set.
	assert.Equal(t, inst.GetLabel("certificateIssuerType"), "")
	assert.Equal(t, inst.GetLabel("certificateExpiryStatus"), "")
}

// TestRunOnlyClassifiesAdminServerCert pins the three-way gate: the serial must
// match the admin SVM's and the type must be "server". Getting this wrong
// labels the wrong certificate as the cluster's.
func TestRunOnlyClassifiesAdminServerCert(t *testing.T) {
	tests := []struct {
		name         string
		adminSerial  string
		serialNumber string
		certType     string
		wantLabelled bool
	}{
		{name: "admin server cert is classified", adminSerial: "ABC123", serialNumber: "ABC123", certType: "server", wantLabelled: true},
		{name: "different serial is not", adminSerial: "ABC123", serialNumber: "XYZ789", certType: "server", wantLabelled: false},
		{name: "client cert is not", adminSerial: "ABC123", serialNumber: "ABC123", certType: "client", wantLabelled: false},
		{name: "root_ca is not", adminSerial: "ABC123", serialNumber: "ABC123", certType: "root_ca", wantLabelled: false},
		{name: "no admin serial resolved", adminSerial: "", serialNumber: "", certType: "server", wantLabelled: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newCertificate(t, tc.adminSerial)
			data, expiration := certMatrix(t)

			inst := newCertInstance(t, data, map[string]string{
				"serial_number": tc.serialNumber, "type": tc.certType,
			})
			expiration.SetValueFloat64(inst, float64(time.Now().Add(365*24*time.Hour).Unix()))

			_, _, err := c.Run(map[string]*matrix.Matrix{"certificate": data})
			assert.Nil(t, err)

			if tc.wantLabelled {
				// No PEM present, so the issuer type falls back to unknown.
				assert.Equal(t, inst.GetLabel("certificateIssuerType"), "unknown")
				assert.Equal(t, inst.GetLabel("certificateExpiryStatus"), "active")
			} else {
				assert.Equal(t, inst.GetLabel("certificateIssuerType"), "")
				assert.Equal(t, inst.GetLabel("certificateExpiryStatus"), "")
			}
		})
	}
}

// TestRunExpiryStatus pins the expiry classification thresholds, which drive
// the dashboard's alerting colours: expired, expiring within 60 days, active.
func TestRunExpiryStatus(t *testing.T) {
	tests := []struct {
		name   string
		expiry time.Time
		want   string
	}{
		{name: "already expired", expiry: time.Now().Add(-24 * time.Hour), want: "expired"},
		{name: "expiring inside 60 days", expiry: time.Now().Add(30 * 24 * time.Hour), want: "expiring"},
		{name: "just inside 60 days", expiry: time.Now().Add(59 * 24 * time.Hour), want: "expiring"},
		{name: "beyond 60 days is active", expiry: time.Now().Add(120 * 24 * time.Hour), want: "active"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newCertificate(t, "ABC123")
			data, expiration := certMatrix(t)

			inst := newCertInstance(t, data, map[string]string{
				"serial_number": "ABC123", "type": "server",
			})
			expiration.SetValueFloat64(inst, float64(tc.expiry.Unix()))

			_, _, err := c.Run(map[string]*matrix.Matrix{"certificate": data})
			assert.Nil(t, err)

			assert.Equal(t, inst.GetLabel("certificateExpiryStatus"), tc.want)
		})
	}
}

// TestRunMissingExpirationMetric pins the guard for a template that does not
// collect the expiration counter: the plugin must skip rather than panic.
func TestRunMissingExpirationMetric(t *testing.T) {
	c := newCertificate(t, "ABC123")
	data := matrix.New("Certificate", "certificate", "certificate")

	inst := newCertInstance(t, data, map[string]string{
		"serial_number": "ABC123", "type": "server",
	})

	_, _, err := c.Run(map[string]*matrix.Matrix{"certificate": data})

	assert.Nil(t, err)
	assert.Equal(t, inst.GetLabel("expiry_time"), "")
}

// TestRunNonExportableInstanceSkipped pins that a filtered-out certificate is
// left untouched.
func TestRunNonExportableInstanceSkipped(t *testing.T) {
	c := newCertificate(t, "ABC123")
	data, expiration := certMatrix(t)

	inst := newCertInstance(t, data, map[string]string{
		"serial_number": "ABC123", "type": "server",
	})
	inst.SetExportable(false)
	expiration.SetValueFloat64(inst, 1893456000)

	_, _, err := c.Run(map[string]*matrix.Matrix{"certificate": data})

	assert.Nil(t, err)
	assert.Equal(t, inst.GetLabel("expiry_time"), "")
}

// TestRunIncrementsCurrentVal covers the cache-refresh counter behind GitHub
// issue #4199 "Certificate plugin does not cache adminVserverSerial between
// polls". The serial is fetched with two REST calls, so it must be looked up
// once per PluginInvocationRate polls rather than on every poll.
func TestRunIncrementsCurrentVal(t *testing.T) {
	c := newCertificate(t, "ABC123")
	data, expiration := certMatrix(t)

	inst := newCertInstance(t, data, map[string]string{
		"serial_number": "ABC123", "type": "server",
	})
	expiration.SetValueFloat64(inst, 1893456000)

	dataMap := map[string]*matrix.Matrix{"certificate": data}

	for i := range 5 {
		_, _, err := c.Run(dataMap)
		assert.Nil(t, err)
		assert.Equal(t, c.currentVal, i+1)
		// The cached serial survives between polls: it is not re-fetched, and
		// not cleared, until the invocation rate is reached.
		assert.Equal(t, c.adminVserverSerial, "ABC123")
	}
}
