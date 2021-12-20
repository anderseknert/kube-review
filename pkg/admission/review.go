package admission

import (
	"encoding/json"
	"fmt"

	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/scheme"
)

//goland:noinspection GoNameStartsWithPackageName
func AdmissionReviewRequest(input []byte, action string) ([]byte, error) {
	actionMapper := map[string]admissionv1.Operation{
		"create":  admissionv1.Create,
		"update":  admissionv1.Update,
		"delete":  admissionv1.Delete,
		"connect": admissionv1.Connect,
	}
	var admissionAction admissionv1.Operation
	var found bool
	if admissionAction, found = actionMapper[action]; !found {
		return nil, fmt.Errorf("unknown action: %v, choose one of 'create', 'update', 'delete' or 'connect'", action)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	object, kind, err := decode(input, nil, nil)
	if err != nil {
		return nil, err
	}

	userInfo := v1.UserInfo{
		Username: "kube-review",
		Groups:   []string{"kube-review"},
		UID:      string(uuid.NewUUID()),
		Extra:    map[string]v1.ExtraValue{},
	}

	metaKind := &metav1.GroupVersionKind{
		Group:   kind.Group,
		Version: kind.Version,
		Kind:    kind.Kind,
	}

	var name, namespace string
	unstructured, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	metadata := unstructured["metadata"]
	if t, ok := metadata.(map[string]interface{}); ok {
		if n, ok2 := t["name"].(string); ok2 {
			name = n
		}
		if n, ok2 := t["namespace"].(string); ok2 {
			namespace = n
		}
	}

	// TODO: Must be a better way?
	r, _ := meta.UnsafeGuessKindToResource(*kind)
	resource := &metav1.GroupVersionResource{Group: r.Group, Version: r.Version, Resource: r.Resource}

	dryRun := true

	admissionRequest := &admissionv1.AdmissionRequest{
		UID:                uuid.NewUUID(),
		Kind:               *metaKind,
		Resource:           *resource,
		SubResource:        "", // TODO
		RequestKind:        metaKind,
		RequestResource:    resource,
		RequestSubResource: "", // TODO
		Name:               name,
		Namespace:          namespace,
		Operation:          admissionAction,
		UserInfo:           userInfo,
		Object:             getNewObject(object, admissionAction),
		OldObject:          getOldObject(object, admissionAction),
		DryRun:             &dryRun,
		Options:            getOptions(admissionAction),
	}

	admissionReview := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{APIVersion: "admission.k8s.io/v1", Kind: "AdmissionReview"},
		Request:  admissionRequest,
	}

	requestJSON, err := json.MarshalIndent(&admissionReview, "", "    ")
	if err != nil {
		return nil, err
	}

	return requestJSON, nil
}

func getNewObject(object runtime.Object, action admissionv1.Operation) runtime.RawExtension {
	if action == admissionv1.Delete {
		return runtime.RawExtension{}
	}
	return runtime.RawExtension{
		Object: object.DeepCopyObject(),
	}
}

func getOldObject(object runtime.Object, action admissionv1.Operation) runtime.RawExtension {
	if action == admissionv1.Create || action == admissionv1.Connect {
		return runtime.RawExtension{}
	}
	return runtime.RawExtension{
		Object: object.DeepCopyObject(),
	}
}

func getOptions(action admissionv1.Operation) runtime.RawExtension {
	switch action {
	case admissionv1.Create:
		return runtime.RawExtension{Object: &metav1.CreateOptions{}}
	case admissionv1.Update:
		return runtime.RawExtension{Object: &metav1.UpdateOptions{}}
	case admissionv1.Delete:
		return runtime.RawExtension{Object: &metav1.DeleteOptions{}}
	default:
		// CONNECT
		return runtime.RawExtension{}
	}
}
