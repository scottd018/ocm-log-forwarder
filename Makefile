#
# developer workflow
#

# binary
BINARY_OUTPUT ?= bin/forwarder
OS ?= darwin
ARCH ?= arm64
build:
	@CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -a -o $(BINARY_OUTPUT)

# image
IMAGE_REPO ?= ghcr.io/scottd018/ocm-log-forwarder
IMAGE_TAG ?= unstable
image:
	@docker build . -t $(IMAGE_REPO):$(IMAGE_TAG)

# tests
DEBUG ?= true
CLUSTER_NAME ?= dscott
test-binary:
	@OCM_CLUSTER_ID=`rosa describe cluster -c $(CLUSTER_NAME) | grep '^ID:' | awk '{print $$NF}'` \
		OCM_POLL_INTERVAL_MINUTES=1 \
		BACKEND_ES_SECRET_NAME="elasticsearch-es-elastic-user" \
		BACKEND_ES_SECRET_NAMESPACE="elastic-system" \
		BACKEND_ES_URL="https://$$(oc -n $${BACKEND_ES_SECRET_NAMESPACE} get route elasticsearch --no-headers | awk '{print $$2}')" \
		DEBUG=$(DEBUG) \
		go run main.go

test-image:
	@docker run \
		-e OCM_CLUSTER_ID=`rosa describe cluster -c $(CLUSTER_NAME) | grep '^ID:' | awk '{print $$NF}'` \
		-e OCM_POLL_INTERVAL_MINUTES=1 \
		-e BACKEND_ES_SECRET_NAME="elasticsearch-es-elastic-user" \
		-e BACKEND_ES_SECRET_NAMESPACE="elastic-system" \
		-e BACKEND_ES_URL="https://$$(oc -n $${BACKEND_ES_SECRET_NAMESPACE} get route elasticsearch --no-headers | awk '{print $$2}')" \
		-e DEBUG=$(DEBUG) \
		-e KUBECONFIG=/kube-config \
		-v /Users/dscott/.kube/config:/kube-config \
		$(IMAGE_REPO):$(IMAGE_TAG)

#
# demo elasticsearch
#
es-operator:
	@oc apply -f deploy/operator.yaml

es-namespace:
	@oc apply -f deploy/namespace.yaml

es-components:
	@oc apply -f deploy/elasticsearch.yaml

OCM_TOKEN_PATH ?= /Users/dscott/.aws/ocm.json
ocm-secret:
	@OCM_CLUSTER_ID=`rosa describe cluster -c $(CLUSTER_NAME) | grep '^ID:' | awk '{print $$NF}'` \
		oc -n elastic-system create secret generic ocm-token --from-file=$$OCM_CLUSTER_ID=$(OCM_TOKEN_PATH)
