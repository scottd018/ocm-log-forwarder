# OCM-LOG-FORWARDER

This project is not supported by Red Hat.  It was a thought experiment on how to take Service Logs from 
[OpenShift Cluster Manager](https://docs.openshift.com/rosa/ocm/ocm-overview.html) and send them to a 
backend system for audit purposes.  It is specifically tested with ROSA OpenShift clusters that are 
registered with OCM.

## Walkthrough

1. Create a ROSA cluster using your preferred method.

2. Enable the ElasticSearch operator:

**NOTE:** for now the ElasticSearch operator is all that is supported by the code base, however it should be 
flexible enough to add more backends moving forward, as needed.

```bash
make es-operator
```

3. Create the Namespace:

```bash
make es-namespace
```

4. Deploy the ElasticSearch components:

```bash
make es-components
```

5. Retrieve you cluster-id and store it as an environment variable:

```bash
export CLUSTER_NAME="my-cluster-name"
export OCM_CLUSTER_ID=$(rosa describe cluster -c ${CLUSTER_NAME} | grep '^ID:' | awk '{print $NF}')
```

6. Create your OCM token as a secret in OpenShift:

```bash
export OCM_TOKEN_PATH="/path/to/my/ocm/token.json"
oc -n elastic-system create secret generic ocm-token --from-file=$OCM_CLUSTER_ID=$OCM_TOKEN_PATH
```

### Testing (Without Deploying the Controller)

7. During development, I found it beneficial to be be able to test outside of deploying to an 
actual cluster.  To do so, you simply need to set some environment variables and run a 
command:

**NOTE:** the above follows the naming of components created during the walkthrough.  If you
have adjusted part of the walkthrough for your use case, be sure to adjust your workflow 
appropriately

```bash
export OCM_CLUSTER_ID=$(rosa describe cluster -c ${CLUSTER_NAME} | grep '^ID:' | awk '{print $NF}')
export BACKEND_ES_SECRET_NAME="elasticsearch-es-elastic-user"
export BACKEND_ES_SECRET_NAMESPACE="elastic-system"
export BACKEND_ES_URL="https://$(oc -n ${BACKEND_ES_SECRET_NAMESPACE} get route elasticsearch --no-headers | awk '{print $2}')"

# lower the poll interval from default to see things happen quicker
export c
```

8. Run the test:

```bash
go run main.go

...
#1 2023-04-05 10:44:53 processor.go:55 ▶ INF initializing kubernetes cluster config: cluster=[cluster-id]
#2 2023-04-05 10:44:53 processor.go:67 ▶ WAR unable to initialize cluster config: cluster=[cluster-id], attempting file initialization
#3 2023-04-05 10:44:53 processor.go:71 ▶ INF initializing kubernetes file config: cluster=[cluster-id], file=[/Users/dscott/.kube/config]
#4 2023-04-05 10:44:53 controller.go:28 ▶ INF initializing backend: cluster=[cluster-id], type=[elasticsearch]
#5 2023-04-05 10:44:54 controller.go:35 ▶ INF initializing poller: cluster=[cluster-id], interval=[1 minutes]
#6 2023-04-05 10:44:54 controller.go:57 ▶ INF starting main program loop
#7 2023-04-05 10:44:54 controller.go:87 ▶ INF polling openshift cluster manager: cluster=[cluster-id]
#8 2023-04-05 10:44:54 poller.go:27 ▶ INF refreshing token: cluster=[cluster-id]
#9 2023-04-05 10:44:54 token.go:31 ▶ INF retrieving cluster secret: cluster=[cluster-id], secret=[ocm-token]
#10 2023-04-05 10:44:54 token.go:37 ▶ INF retrieving bearer token: cluster=[cluster-id]
#11 2023-04-05 10:44:54 poller.go:34 ▶ INF retrieving service logs: cluster=[cluster-id]
#12 2023-04-05 10:44:54 elasticsearch.go:119 ▶ INF sending items to elasticsearch: cluster=cluster-id, event_stream_id=event-stream-id, index=ocm_service_logs
#13 2023-04-05 10:44:54 elasticsearch.go:119 ▶ INF sending items to elasticsearch: cluster=cluster-id, event_stream_id=event-stream-id, index=ocm_service_logs
#14 2023-04-05 10:44:54 elasticsearch.go:119 ▶ INF sending items to elasticsearch: cluster=cluster-id, event_stream_id=event-stream-id, index=ocm_service_logs
#15 2023-04-05 10:44:54 elasticsearch.go:119 ▶ INF sending items to elasticsearch: cluster=cluster-id, event_stream_id=event-stream-id, index=ocm_service_logs
...
```

9. Open Kibana and view the logs (index name is `ocm_service_logs` or can be specified by `BACKEND_ES_INDEX`).  You can get the URL by running the following:

```bash
KIBANA_URL="https://$(oc -n ${BACKEND_ES_SECRET_NAMESPACE} get route kibana --no-headers | awk '{print $2}')"
echo $KIBANA_URL
```
