## Usage

### Sample YAML Configuration

Create a YAML file (`kafka-cluster.yaml`) with the desired configuration:

```yaml
apiVersion: code2cloud.planton.cloud/v1
kind: KafkaKubernetes
metadata:
  id: my-kafka-cluster
spec:
  kubernetes_cluster_credential_id: your-cluster-credential-id
  broker_container:
    replicas: 3
    resources:
      requests:
        cpu: "500m"
        memory: "1Gi"
      limits:
        cpu: "1"
        memory: "2Gi"
    disk_size: "10Gi"
  zookeeper_container:
    replicas: 3
    resources:
      requests:
        cpu: "500m"
        memory: "1Gi"
      limits:
        cpu: "1"
        memory: "2Gi"
    disk_size: "5Gi"
  kafka_topics:
    - name: topic1
      partitions: 3
      replicas: 3
    - name: topic2
      partitions: 1
      replicas: 1
  schema_registry_container:
    is_enabled: true
    replicas: 1
    resources:
      requests:
        cpu: "100m"
        memory: "256Mi"
      limits:
        cpu: "200m"
        memory: "512Mi"
  ingress:
    is_enabled: true
    endpoint_domain_name: "example.com"
  is_deploy_kafka_ui: true
```

### Deploying with CLI

Use the provided CLI tool to deploy the Kafka cluster:

```bash
platon pulumi up --stack-input kafka-cluster.yaml
```

If no Pulumi module is specified, the CLI uses the default module corresponding to the API resource.
