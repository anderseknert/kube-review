package admission

import (
	"bytes"
	"encoding/json"
	"fmt"

	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
)

//goland:noinspection GoNameStartsWithPackageName
func CreateAdmissionReviewRequest(input []byte, action string, username string, groups []string) ([]byte, error) {
	operation, err := actionToOperation(action)
	if err != nil {
		return nil, err
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	object, kind, err := decode(input, nil, nil)
	if err != nil {
		// Failure to decode, likely due to unrecognized type, try unstructured
		return fromUnstructured(input, operation, username, groups)
	}

	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		return nil, fmt.Errorf("failed to parse object, %w", err)
	}

	userInfo := getUserInfo(username, groups)
	newObject := getNewObject(object, *operation)
	oldObject := getOldObject(object, *operation)

	return createAdmissionRequest(unstructured, *kind, operation, userInfo, newObject, oldObject)
}

func fromUnstructured(input []byte, operation *admissionv1.Operation, username string, groups []string) ([]byte, error) {
	var object interface{}

	// Try "brute force" serialization of unknown type
	err := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(input), 4096).Decode(&object)
	if err != nil {
		return nil, err
	}
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&object)
	if err != nil {
		return nil, fmt.Errorf("failed to parse object, %w", err)
	}

	version, err := schema.ParseGroupVersion(unstructured["apiVersion"].(string))
	if err != nil {
		return nil, err
	}
	withKind := version.WithKind(unstructured["kind"].(string))
	kind := &schema.GroupVersionKind{
		Group:   withKind.Group,
		Version: withKind.Version,
		Kind:    withKind.Kind,
	}
	userInfo := getUserInfo(username, groups)

	var unknown runtime.Unknown
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &unknown)
	if err != nil {
		return nil, err
	}

	newObject := getUnknownRaw(&unknown, *operation)
	oldObject := getOldUnknownRaw(&unknown, *operation)

	return createAdmissionRequest(unstructured, *kind, operation, userInfo, newObject, oldObject)
}

func createAdmissionRequest(unstructured map[string]interface{}, gvk schema.GroupVersionKind, operation *admissionv1.Operation, user v1.UserInfo, object, oldObject runtime.RawExtension) ([]byte, error) {
	dryRun := true

	name, namespace := getNameAndNamespace(unstructured)
	r, _ := meta.UnsafeGuessKindToResource(gvk)
	resource := &metav1.GroupVersionResource{Group: r.Group, Version: r.Version, Resource: r.Resource}

	kind := gvkMeta(gvk.Group, gvk.Version, gvk.Kind)

	admissionRequest := &admissionv1.AdmissionRequest{
		UID:                uuid.NewUUID(),
		Kind:               *kind,
		Resource:           *resource,
		SubResource:        "", // TODO
		RequestKind:        kind,
		RequestResource:    resource,
		RequestSubResource: "", // TODO
		Name:               name,
		Operation:          *operation,
		UserInfo:           user,
		Object:             object,
		OldObject:          oldObject,
		DryRun:             &dryRun,
		Options:            getOptions(*operation),
	}
	if namespace != "" {
		admissionRequest.Namespace = namespace
	}

	admissionReview := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{APIVersion: "admission.k8s.io/v1", Kind: "AdmissionReview"},
		Request:  admissionRequest,
	}

	requestJSON, err := json.MarshalIndent(&admissionReview, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed encoding object to JSON %w", err)
	}

	return requestJSON, nil
}

func gvkMeta(group, version, kind string) *metav1.GroupVersionKind {
	return &metav1.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
}

func actionToOperation(action string) (*admissionv1.Operation, error) {
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

	return &admissionAction, nil
}

func getNameAndNamespace(unstructured map[string]interface{}) (name string, namespace string) {
	metadata := unstructured["metadata"]
	if t, ok := metadata.(map[string]interface{}); ok {
		if n, ok2 := t["name"].(string); ok2 {
			name = n
		}
		if n, ok2 := t["namespace"].(string); ok2 {
			namespace = n
		}
	}

	return name, namespace
}

func getUserInfo(username string, groups []string) v1.UserInfo {
	return v1.UserInfo{
		Username: username,
		Groups:   groups,
		UID:      string(uuid.NewUUID()),
		Extra:    map[string]v1.ExtraValue{},
	}
}

func getUnknownRaw(unknown *runtime.Unknown, action admissionv1.Operation) runtime.RawExtension {
	if action == admissionv1.Delete {
		return runtime.RawExtension{}
	}

	return runtime.RawExtension{
		Object: unknown,
	}
}

func getNewObject(object runtime.Object, action admissionv1.Operation) runtime.RawExtension {
	if action == admissionv1.Delete {
		return runtime.RawExtension{}
	}

	return runtime.RawExtension{
		Object: object.DeepCopyObject(),
	}
}

func getOldUnknownRaw(unknown *runtime.Unknown, action admissionv1.Operation) runtime.RawExtension {
	if action == admissionv1.Create || action == admissionv1.Connect {
		return runtime.RawExtension{}
	}

	return runtime.RawExtension{
		Object: unknown,
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
		return runtime.RawExtension{Object: &metav1.CreateOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CreateOptions",
				APIVersion: "meta.k8s.io/v1",
			},
		}}
	case admissionv1.Update:
		return runtime.RawExtension{Object: &metav1.UpdateOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "UpdateOptions",
				APIVersion: "meta.k8s.io/v1",
			},
		}}
	case admissionv1.Delete:
		return runtime.RawExtension{Object: &metav1.DeleteOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeleteOptions",
				APIVersion: "meta.k8s.io/v1",
			},
		}}
	default:
		// CONNECT
		return runtime.RawExtension{}
	}
}
