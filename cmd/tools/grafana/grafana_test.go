package grafana

import (
	"regexp"
	"testing"
)

func TestCheckVersion(t *testing.T) {

	inputVersion := []string{"7.2.3.4", "abc.1.3", "4.5.4", "7.1.0", "7.5.5"}
	expectedOutPut := []bool{true, false, false, true, true}
	// version length greater than 3

	for i, s := range inputVersion {
		c := checkVersion(s)
		if c != expectedOutPut[i] {
			t.Errorf("Expected %t but got %t for input %s", expectedOutPut[i], c, inputVersion[i])
		}
	}
}

func TestHttpsAddr(t *testing.T) {
	opts.addr = "https://1.1.1.1:3000"
	adjustOptions()
	if opts.addr != "https://1.1.1.1:3000" {
		t.Errorf("Expected opts.addr to be %s but got %s", "https://1.1.1.1:3000", opts.addr)
	}

	opts.addr = "https://1.1.1.1:3000"
	opts.useHttps = false // addr takes precedence over useHttps
	adjustOptions()
	if opts.addr != "https://1.1.1.1:3000" {
		t.Errorf("Expected opts.addr to be %s but got %s", "https://1.1.1.1:3000", opts.addr)
	}

	opts.addr = "http://1.1.1.1:3000"
	adjustOptions()
	if opts.addr != "http://1.1.1.1:3000" {
		t.Errorf("Expected opts.addr to be %s but got %s", "http://1.1.1.1:3000", opts.addr)
	}

	// Old way of specifying https
	opts.addr = "http://1.1.1.1:3000"
	opts.useHttps = true
	adjustOptions()
	if opts.addr != "https://1.1.1.1:3000" {
		t.Errorf("Expected opts.addr to be %s but got %s", "https://1.1.1.1:3000", opts.addr)
	}
}

// tests function addPrefixToMetricNames()
func TestAddPrefixToMetricNames(t *testing.T) {

	var (
		examples, expected [7]string
		prefix, result     string
		regex              *regexp.Regexp
		i                  int
	)

	regex = regexp.MustCompile(`([a-zA-Z_+]+)\{.+?}`)

	prefix = "xx_"

	examples = [7]string{
		`sum(volume_read_data{datacenter=\"$Datacenter\",cluster=~\"$Cluster\"}) by (cluster) + sum(volume_write_data{datacenter=\"$Datacenter\",cluster=~\"$Cluster\"}) by(cluster)`,
		`sum(topk($TopResources, volume_total_ops{datacenter=\"$Datacenter\",cluster=\"$Cluster\",svm=~\"$SVM\",volume=~\"$Volume\"}))`,
		`volume_size_used_percent{datacenter=\"$Datacenter\",cluster=\"$Cluster\",svm=~\"$SVM\",volume=~\"$Volume\"}`,
		`avg by(iscsi_lif) (iscsi_lif_iscsi_read_ops+iscsi_lif_iscsi_write_ops+iscsi_lif_iscsi_other_ops{datacenter=\"$Datacenter\",cluster=\"$Cluster\",node=~\"$Node\"})`,
		`label_values(metadata_component_status{type="collector",poller=~"$Poller"}, name)`,
		`label_values(poller_status, datacenter)`,
		`label_values(datacenter)`,
	}

	expected = [7]string{
		`sum(xx_volume_read_data{datacenter=\"$Datacenter\",cluster=~\"$Cluster\"}) by (cluster) + sum(xx_volume_write_data{datacenter=\"$Datacenter\",cluster=~\"$Cluster\"}) by(cluster)`,
		`sum(topk($TopResources, xx_volume_total_ops{datacenter=\"$Datacenter\",cluster=\"$Cluster\",svm=~\"$SVM\",volume=~\"$Volume\"}))`,
		`xx_volume_size_used_percent{datacenter=\"$Datacenter\",cluster=\"$Cluster\",svm=~\"$SVM\",volume=~\"$Volume\"}`,
		`avg by(iscsi_lif) (xx_iscsi_lif_iscsi_read_ops+xx_iscsi_lif_iscsi_write_ops+xx_iscsi_lif_iscsi_other_ops{datacenter=\"$Datacenter\",cluster=\"$Cluster\",node=~\"$Node\"})`,
		`label_values(xx_metadata_component_status{type="collector",poller=~"$Poller"}, name)`,
		`label_values(xx_poller_status, datacenter)`,
		`label_values(datacenter)`, // no metric name
	}

	for i = range examples {
		result = addPrefixToMetricNames(examples[i], prefix, regex)
		if result != expected[i] {
			t.Errorf("\nExpected: [%s]\n     Got: [%s]", expected[i], result)
		}
	}
}
