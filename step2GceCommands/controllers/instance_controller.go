/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	computev1 "gce-operator/api/v1"

	////
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	ownerKey = ".metadata.controller"
	apiGVStr = computev1.GroupVersion.String()
)

////
var log = logf.Log.WithName("controller_instance")

//var _ reconcile.Reconciler = &InstanceReconciler{}

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

/*
// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &InstanceReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
}
*/

// +kubebuilder:rbac:groups=compute.gce.infradvisor.fr,resources=instances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compute.gce.infradvisor.fr,resources=instances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods/status,verbs=get;update;patch

func (r *InstanceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("instance", req.NamespacedName)

	// your logic here

	lbls := labels.Set{
		"app":     req.Name,
		"version": computev1.GroupVersion.Version,
	}

	// test logging
	reqLogger := log.WithValues("Req.Namespace", req.Namespace, "Req.Name", req.Name)

	// get instance

	instance := &computev1.Instance{}

	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)

	if err != nil {

		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue

			//reqLogger.Info("Error:" + err.Error())
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// get existing pods

	existingPods := &corev1.PodList{}
	_ = r.Client.List(context.TODO(),
		existingPods,
		&client.ListOptions{
			Namespace:     req.Namespace,
			LabelSelector: labels.SelectorFromSet(lbls),
		})

	existingPodNames := []string{}
	failingPodNames := []string{}
	// Count the pods that are pending or running as available
	for _, pod := range existingPods.Items {
		if pod.GetObjectMeta().GetDeletionTimestamp() != nil {
			continue
		}
		if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodRunning {
			existingPodNames = append(existingPodNames, pod.GetObjectMeta().GetName())
		} else {
			failingPodNames = append(failingPodNames, pod.GetObjectMeta().GetName())
		}
	}

	//reqLogger.Info("get existing pods count= " + strconv.Itoa(len(existingPods.Items)))

	// too many pods ???

	if int32(len(existingPodNames)) > instance.Spec.Replicas {
		// delete a pod. Just one at a time (this reconciler will be called again afterwards)
		reqLogger.Info("Deleting a pod in the instance", "expected replicas", instance.Spec.Replicas, "Pod.Names", existingPodNames)

		pod := existingPods.Items[0]
		err = r.Client.Delete(context.TODO(), &pod)
		if err != nil {
			reqLogger.Error(err, "failed to delete a pod")
			return ctrl.Result{}, err
		}
	}

	// new pod

	if int32(len(existingPodNames)) < instance.Spec.Replicas {
		newpod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "gceinstance-pod",
				Namespace:    req.Namespace,
				Labels:       lbls,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "busybox",
						Image:   "busybox",
						Command: []string{"sleep", "3600"},
					},
				},
			},
		}

		if err := ctrl.SetControllerReference(instance, &newpod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		err = r.Client.Create(context.TODO(), &newpod)
		if err != nil {
			reqLogger.Error(err, "failed to add a pod")
			return ctrl.Result{}, err
		}

	}

	return ctrl.Result{}, nil
}

func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	/// Try adding pods

	if err := mgr.GetFieldIndexer().IndexField(&corev1.Pod{}, ownerKey, func(rawObj runtime.Object) []string {
		// grab the pods, extract the owner...
		pod := rawObj.(*corev1.Pod)
		owner := metav1.GetControllerOf(pod)
		if owner == nil {
			return nil
		}
		// ...make sure it's my objects
		if owner.APIVersion != apiGVStr || owner.Kind != "Instance" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	///
	return ctrl.NewControllerManagedBy(mgr).
		For(&computev1.Instance{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
