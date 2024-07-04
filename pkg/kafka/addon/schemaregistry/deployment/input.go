package deployment

import (
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/adminuser"
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/listener"
	kubernetesv1model "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubecluster/model"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	// DockerImage https://hub.docker.com/layers/confluentinc/cp-schema-registry/7.2.6
	DockerImage         = "confluentinc/cp-schema-registry:7.2.6"
	ContainerPort       = 8081
	KafkaStoreTopicName = "schema-registry"
	Name                = "schema-registry"
)

type input struct {
	labels                           map[string]string
	namespace                        *pulumikubernetescorev1.Namespace
	schemaRegistryDeploymentName     string
	namespaceName                    string
	bootstrapServerHostname          string
	bootstrapServerPort              int32
	saslJaasConfigKeyInSecret        string
	saslJaasConfigSecretName         string
	schemaRegistryContainerResources *kubernetesv1model.ContainerResources
}

func extractInput(ctx *pulumi.Context) *input {
	var contextState = ctx.Value(kafkacontextstate.Key).(kafkacontextstate.ContextState)

	return &input{
		labels:                           contextState.Spec.Labels,
		namespace:                        contextState.Status.AddedResources.Namespace,
		schemaRegistryDeploymentName:     Name,
		namespaceName:                    contextState.Spec.NamespaceName,
		bootstrapServerHostname:          contextState.Spec.ExternalBootstrapHostname,
		bootstrapServerPort:              listener.ExternalPublicListenerPortNumber,
		saslJaasConfigSecretName:         adminuser.SaslPasswordSecretName,
		saslJaasConfigKeyInSecret:        adminuser.SaslJaasConfigKeyInSecret,
		schemaRegistryContainerResources: contextState.Spec.SchemaRegistryContainerSpec.Resources,
	}
}
