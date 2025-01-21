package admission

import (
	"encoding/json"
	"os"
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

func mustReadFileString(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}
