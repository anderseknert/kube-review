package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/scheme"
)

func main() {
	f, err := ioutil.ReadFile("deployment.yaml")
	if err != nil {
		panic(err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	object, kind, err := decode(f, nil, nil)
	if err != nil {
		panic(err)
	}

	userInfo := authenticationv1.UserInfo{
		Username: "kube-review",
		Groups:   []string{"kube-review"},
		UID:      string(uuid.NewUUID()),
		Extra:    map[string]authenticationv1.ExtraValue{},
	}

	dryRun := true

	metaKind := &metav1.GroupVersionKind{
		Group:   kind.Group,
		Version: kind.Version,
		Kind:    kind.Kind,
	}

	var name, namespace string
	unstructured, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	metadata := unstructured["metadata"]
	switch t := metadata.(type) {
	case map[string]interface{}:
		switch n := t["name"].(type) {
		case string:
			name = n
		}
		switch n := t["namespace"].(type) {
		case string:
			namespace = n
		}
	}

	admissionRequest := &admissionv1.AdmissionRequest{
		UID:                uuid.NewUUID(),
		Kind:               *metaKind,
		Resource:           metav1.GroupVersionResource{Group: kind.Group, Version: kind.Version, Resource: "deployments"}, // TODO
		SubResource:        "",                                                                                             // TODO
		RequestKind:        metaKind,
		RequestResource:    &metav1.GroupVersionResource{Group: kind.Group, Version: kind.Version, Resource: "deployments"}, // TODO
		RequestSubResource: "",                                                                                              // TODO
		Name:               name,
		Namespace:          namespace,
		Operation:          admissionv1.Create, // TODO
		UserInfo:           userInfo,
		Object: runtime.RawExtension{
			Object: object.DeepCopyObject(),
		},
		OldObject: runtime.RawExtension{
			Object: object.DeepCopyObject(),
		},
		DryRun: &dryRun,
		Options: runtime.RawExtension{
			Object: &metav1.CreateOptions{},
		},
	}

	admissionReview := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{APIVersion: "admission.k8s.io/v1", Kind: "AdmissionReview"},
		Request:  admissionRequest,
	}

	requestJson, err := json.MarshalIndent(&admissionReview, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(requestJson))
}
