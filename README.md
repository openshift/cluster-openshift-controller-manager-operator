# cluster-openshift-controller-manager-operator
The openshift-controller-manager operator is an 
[OpenShift ClusterOperator](https://github.com/openshift/enhancements/blob/master/enhancements/operator-dev-doc.md#what-is-an-openshift-clusteroperator).    
It installs and maintains the OpenShiftControllerManager [Custom Resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) in a cluster and can be viewed with:     
```
oc get clusteroperator openshift-controller-manager -o yaml
```

The [Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#customresourcedefinitions)
`openshiftcontrollermanagers.operator.openshift.io`    
can be viewed in a cluster with:

```console
$ oc get crd openshiftcontrollermanagers.operator.openshift.io -o yaml
```

Many OpenShift ClusterOperators share common build, test, deployment, and update methods.    
For more information about how to build, deploy, test, update, and develop OpenShift ClusterOperators, see      
[OpenShift ClusterOperator and Operand Developer Document](https://github.com/openshift/enhancements/blob/master/enhancements/operator-dev-doc.md#how-do-i-buildupdateverifyrun-unit-tests)

This section explains how to deploy OpenShift with your test openshift-controller-manager image:        
[Testing a ClusterOperator/Operand image in a cluster](https://github.com/openshift/enhancements/blob/master/enhancements/operator-dev-doc.md#how-can-i-test-changes-to-an-openshift-operatoroperandrelease-component)
