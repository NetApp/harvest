package vscan

import "testing"

func Test_splitOntapName(t *testing.T) {
	tests := []struct {
		name      string
		ontapName string
		svm       string
		scanner   string
		node      string
		isValid   bool
	}{
		{
			name:      "valid",
			ontapName: "svm:scanner:node",
			svm:       "svm",
			scanner:   "scanner",
			node:      "node",
			isValid:   true,
		},
		{
			name:      "ipv6",
			ontapName: "moon-ad:2a03:1e80:a15:60c::1:2a5:moon-02",
			svm:       "moon-ad",
			scanner:   "2a03:1e80:a15:60c::1:2a5",
			node:      "moon-02",
			isValid:   true,
		},
		{
			name:      "invalid zero colon",
			ontapName: "svm",
			svm:       "",
			scanner:   "",
			node:      "",
			isValid:   false,
		},
		{
			name:      "invalid one colon",
			ontapName: "svm:scanner",
			svm:       "",
			scanner:   "",
			node:      "",
			isValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSVM, gotScanner, gotNode, ok := splitOntapName(tt.ontapName)
			if gotSVM != tt.svm {
				t.Errorf("splitOntapName() got = %v, want %v", gotSVM, tt.svm)
			}
			if gotScanner != tt.scanner {
				t.Errorf("splitOntapName() got = %v, want %v", gotScanner, tt.scanner)
			}
			if gotNode != tt.node {
				t.Errorf("splitOntapName() got = %v, want %v", gotNode, tt.node)
			}
			if ok != tt.isValid {
				t.Errorf("splitOntapName() got = %v, want %v", ok, tt.isValid)
			}
		})
	}
}
