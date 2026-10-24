package admission

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	v1 "k8s.io/api/admission/v1"
)

func TestBasicReview(t *testing.T) {
	t.Parallel()

	manifest := mustReadFileString(t, "testdata/in.yaml")

	reviewBytes, err := CreateAdmissionReviewRequest(
		[]byte(manifest),
		"create",
		"kube-review",
		[]string{"system:masters"},
		2,
	)
	if err != nil {
		t.Fatal(err)
	}

	var review v1.AdmissionReview

	err = json.Unmarshal(reviewBytes, &review)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReviewMultipleDocsFirstOneEmpty(t *testing.T) {
	t.Parallel()

	manifest := mustReadFileString(t, "testdata/twodocsfirstempty.yaml")

	reviewBytes, err := CreateAdmissionReviewRequest(
		[]byte(manifest),
		"create",
		"kube-review",
		[]string{"system:masters"},
		2,
	)
	if err != nil {
		t.Fatal(err)
	}

	var review v1.AdmissionReview

	err = json.Unmarshal(reviewBytes, &review)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReviewUnknownResourceRequiresAPIVersionAndKind(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
		wantErr  string
	}{
		{
			name:     "missing apiVersion",
			manifest: "kind: Synthetic\nmetadata:\n  name: sample\n",
			wantErr:  "apiVersion must be a non-empty string",
		},
		{
			name:     "missing kind",
			manifest: "apiVersion: example.test/v1\nmetadata:\n  name: sample\n",
			wantErr:  "kind must be a non-empty string",
		},
		{
			name:     "non-string apiVersion",
			manifest: "apiVersion: 1\nkind: Synthetic\n",
			wantErr:  "apiVersion must be a non-empty string",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := CreateAdmissionReviewRequest([]byte(test.manifest), "create", "kube-review", nil, 2)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("expected error containing %q, got %v", test.wantErr, err)
			}
		})
	}
}

func mustReadFileString(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}
