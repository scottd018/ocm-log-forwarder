# OCM-LOG-FORWARDER

This project is not supported by Red Hat.  It was a thought experiment on how to take Service Logs from 
[OpenShift Cluster Manager](https://docs.openshift.com/rosa/ocm/ocm-overview.html) and send them to a 
backend system for audit purposes.  It is specifically tested with ROSA OpenShift clusters that are 
registered with OCM.

## Walkthrough (Testing Out of Cluster)

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

