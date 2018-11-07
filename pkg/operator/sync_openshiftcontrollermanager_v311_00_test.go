package operator

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	configv1 "github.com/openshift/api/config/v1"
	fakeConfig "github.com/openshift/client-go/config/clientset/versioned/fake"
	v1alpha1 "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/apis/openshiftcontrollermanager/v1alpha1"
)

const dummyCA = `-----BEGIN CERTIFICATE-----
VEhJUyBJUyBBIERVTU1ZIENFUlRJRklDQVRFIFdJVEggQkFTRTY0IEVOQ09ERUQg
REFUQSBBTkQgU0hPVUxEIE5PVCBCRSBVU0VECg==
-----END CERTIFICATE-----
`

func TestManageBuildAdditionalCAConfigMap(t *testing.T) {

	tests := []struct {
		name           string
		inputCAMap     *corev1.ConfigMap
		statusCAMap    *corev1.ConfigMap
		expectModified bool
		expectError    bool
	}{
		{
			name:           "no-ca",
			expectModified: false,
		},
		{
			name: "observed-cm-no-user-cm",
			inputCAMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "openshift-build-additional-ca",
					Namespace: openshiftConfigNamespaceName,
				},
				Data: map[string]string{
					"my-ca.crt": dummyCA,
				},
			},
			expectModified: true,
		},
		{
			name: "observed-cm-user-cm-mismatch",
			inputCAMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "openshift-build-additional-ca",
					Namespace: openshiftConfigNamespaceName,
				},
				Data: map[string]string{
					"foo.crt": dummyCA + dummyCA,
				},
			},
			statusCAMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: openshiftConfigNamespaceName,
					Name:      "openshift-build-additional-ca",
				},
				Data: map[string]string{
					"bar.crt": dummyCA,
				},
			},
			expectModified: true,
		},
		{
			name: "observed-cm-user-cm-match",
			inputCAMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "openshift-build-additional-ca",
					Namespace: openshiftConfigNamespaceName,
				},
				Data: map[string]string{
					"foo.crt": dummyCA,
				},
			},
			statusCAMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "openshift-build-additional-ca",
					Namespace: openshiftConfigNamespaceName,
				},
				Data: map[string]string{
					"foo.crt": dummyCA,
				},
			},
		},
		{
			name: "remove-deployed-ca",
			statusCAMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "openshift-build-additional-ca",
					Namespace: openshiftConfigNamespaceName,
				},
				Data: map[string]string{
					"foo.crt": dummyCA,
				},
			},
			expectModified: true,
		},
		{
			name: "error if bad input CA",
			inputCAMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "openshift-build-additional-ca",
					Namespace: openshiftConfigNamespaceName,
				},
				Data: map[string]string{
					"foo.crt": "THIS IS AN INVALID PEM FORMAT",
				},
			},
			expectError: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			coreObjs := []runtime.Object{}
			configObjs := []runtime.Object{}
			expectedHash := ""
			opConfig := &v1alpha1.OpenShiftControllerManagerOperatorConfig{}
			ctrlMgrCAMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: targetNamespaceName,
					Name:      "build-additional-ca",
				},
			}
			if tc.inputCAMap != nil {
				expectedHash = hashCAMap(tc.inputCAMap)
				coreObjs = append(coreObjs, tc.inputCAMap)
				bc := &configv1.Build{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster",
					},
					Spec: configv1.BuildSpec{
						AdditionalTrustedCA: configv1.ConfigMapReference{
							Name:      tc.inputCAMap.Name,
							Namespace: tc.inputCAMap.Namespace,
						},
					},
				}
				configObjs = append(configObjs, bc)
				opConfig.Spec.AdditionalTrustedCA = &v1alpha1.AdditionalTrustedCA{
					SHA1Hash:      expectedHash,
					ConfigMapName: tc.inputCAMap.Name,
				}
			}
			if tc.statusCAMap != nil {
				opConfig.Status.AdditionalTrustedCA = &v1alpha1.AdditionalTrustedCA{
					SHA1Hash:      hashCAMap(tc.statusCAMap),
					ConfigMapName: tc.statusCAMap.Name,
				}
				ctrlMgrCAMap.Data = tc.statusCAMap.Data
			}
			coreObjs = append(coreObjs, ctrlMgrCAMap)

			testClient := fake.NewSimpleClientset(coreObjs...)
			configClient := fakeConfig.NewSimpleClientset(configObjs...)
			result, modified, err := manageBuildAdditionalCAConfigMap(testClient.CoreV1(), configClient.ConfigV1(), opConfig)
			if tc.expectError {
				if err == nil {
					t.Fatal("expected error did not occur")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error occurred: %v", err)
			}
			if tc.expectModified != modified {
				t.Errorf("expected configMap to be modified: %v, got %v", tc.expectModified, modified)
			}
			if expectedHash != result {
				cm, _ := testClient.CoreV1().ConfigMaps("openshift-controller-manager").Get("build-additional-ca", metav1.GetOptions{})
				t.Logf("expected data:\n%v\ngot:\n%v", tc.inputCAMap.Data, cm.Data)
				t.Errorf("expected hash %s, got %s", expectedHash, result)
			}
			if len(expectedHash) > 0 {
				if opConfig.Status.AdditionalTrustedCA == nil {
					t.Error("expected operator config status to report AdditionalTrustedCA data, got nil")
				} else if expectedHash != opConfig.Status.AdditionalTrustedCA.SHA1Hash {
					t.Errorf("expected operator config status to have SHA1Hash %s, got %s", expectedHash, opConfig.Status.AdditionalTrustedCA.SHA1Hash)
				}
			}
		})
	}
}

func hashCAMap(cm *corev1.ConfigMap) string {
	h := sha1.New()
	data := &bytes.Buffer{}
	if cm != nil {
		for _, v := range cm.Data {
			if len(v) > 0 {
				data.WriteString(v)
			}
		}
	}
	toHash := data.Bytes()
	if len(toHash) == 0 {
		return ""
	}
	h.Write(toHash)
	return fmt.Sprintf("%x", h.Sum(nil))
}
