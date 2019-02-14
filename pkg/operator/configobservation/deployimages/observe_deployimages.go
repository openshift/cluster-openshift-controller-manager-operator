package deployimages

import (
	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/events"
)

func ObserveControllerManagerImagesConfig(genericListers configobserver.Listers, recorder events.Recorder, existingConfig map[string]interface{}) (map[string]interface{}, []error) {
	listers := genericListers.(configobservation.Listers)
	var errs []error
	prevObservedConfig := map[string]interface{}{}

	// first observe all the existing config values so that if we get any errors
	// we can at least return those.
	builderImagePath := []string{"build", "imageTemplateFormat", "format"}
	currentBuilderImage, _, err := unstructured.NestedString(existingConfig, builderImagePath...)
	if err != nil {
		return prevObservedConfig, append(errs, err)
	}
	if len(currentBuilderImage) > 0 {
		err := unstructured.SetNestedField(prevObservedConfig, currentBuilderImage, builderImagePath...)
		if err != nil {
			return prevObservedConfig, append(errs, err)
		}
	}

	deployerImagePath := []string{"deployer", "imageTemplateFormat", "format"}
	currentDeployerImage, _, err := unstructured.NestedString(existingConfig, deployerImagePath...)
	if err != nil {
		return prevObservedConfig, append(errs, err)
	}
	if len(currentDeployerImage) > 0 {
		err := unstructured.SetNestedField(prevObservedConfig, currentDeployerImage, deployerImagePath...)
		if err != nil {
			return prevObservedConfig, append(errs, err)
		}
	}

	// now gather the cluster config and turn it into the observed config
	observedConfig := map[string]interface{}{}
	controllerManagerImagesConfigMap, err := listers.ConfigMapLister.ConfigMaps(util.OperatorNamespaceName).Get("openshift-controller-manager-images")
	if errors.IsNotFound(err) {
		glog.V(2).Infof("configmap/openshift-controller-manager-images: not found")
		return observedConfig, errs
	}
	if err != nil {
		return prevObservedConfig, append(errs, err)
	}
	if controllerManagerImagesConfigMap != nil {
		// TODO(juanvallejo): reflect any issues in operator status
		if err = configobservation.ObserveField(observedConfig, controllerManagerImagesConfigMap.Data["builderImage"], "build.imageTemplateFormat.format", true); err != nil {
			return nil, append(errs, err)
		}
		if err = configobservation.ObserveField(observedConfig, controllerManagerImagesConfigMap.Data["deployerImage"], "deployer.imageTemplateFormat.format", true); err != nil {
			return nil, append(errs, err)
		}
	}

	return observedConfig, errs
}
