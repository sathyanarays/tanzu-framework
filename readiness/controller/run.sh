export KIND_CLUSTER_NAME=kind
export KUBE_VERSION=v1.26.3 
pushd ../..
kind delete clusters -A  && make create-kind-cluster && make deploy-local-readiness
popd