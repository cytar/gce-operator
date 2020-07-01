*To initiate the kubebuilder structure:

go mod init gce-operator
export GO111MODULE=on
kubebuilder init --domain gce.infradvisor.fr
kubebuilder create api --group compute --version v1 --kind Instance

**In the last structure:

*To install CRD on k8s cluster & run local controller & manager

#kustomize build config/crd | kubectl apply -f -
make install

make run

*To create a sample "Instance" object

kubectl apply -f config/samples/compute_v1_instance.yaml

*To biuld controller/manager image

#docker build . -t k8s-controller-compute.gce.infradvisor.fr:0.3
#make docker-build docker-push IMG=cytar/k8s-controller-compute.gce.infradvisor.fr:0.3
make docker-build IMG=k8s-controller-compute.gce.infradvisor.fr:0.3

*To deploy the controller on current k8s cluster

#kustomize build config/default | kubectl apply -f -
#make deploy IMG=cytar/k8s-controller-compute.gce.infradvisor.fr:0.3
make deploy IMG=k8s-controller-compute.gce.infradvisor.fr:0.3


