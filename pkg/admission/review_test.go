package admission

import (
	"encoding/json"
	v1 "k8s.io/api/admission/v1"
	"testing"
)

func TestBasicReview(t *testing.T) {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
        ports:
        - containerPort: 8080`

	reviewBytes, err := CreateAdmissionReviewRequest(
		[]byte(manifest), "create", "kube-review", []string{"system:masters"},
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
