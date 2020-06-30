*To initiate the kubebuilder structure:

go mod init gce-operator
export GO111MODULE=on
kubebuilder init --domain gce.infradvisor.fr
kubebuilder create api --group compute --version v1 --kind Instance

**In the last structure:

*To install CRD on k8s cluster & run local controller & manager

make install;make run

*To create a sample "Instance" object

kubectl apply -f config/samples/compute_v1_instance.yaml

*To biuld controller/manager image

make docker-build IMG=k8s-controller-compute.gce.infradvisor.fr:0.3

*To deploy the controller on current k8s cluster

make deploy IMG=k8s-controller-compute.gce.infradvisor.fr:0.3


