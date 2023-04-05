#
# elasticsearch
#
es-operator:
	@oc apply -f deploy/operator.yaml

es-namespace:
	@oc apply -f deploy/namespace.yaml

es-components:
	@oc apply -f deploy/elasticsearch.yaml

ocm-secret:
	@