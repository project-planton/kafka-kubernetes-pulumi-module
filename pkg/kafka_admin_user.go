package pkg

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/strimzioperator/kafka/v1beta2"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func kafkaAdminUser(ctx *pulumi.Context, locals *Locals, createdNamespace *kubernetescorev1.Namespace,
	createdKafkaCluster *v1beta2.Kafka, labels map[string]string) error {

	//add the label required to create the admin secret for the target kafka-cluster
	labels["strimzi.io/cluster"] = locals.KafkaKubernetes.Metadata.Id

	_, err := v1beta2.NewKafkaUser(ctx,
		"admin-user",
		&v1beta2.KafkaUserArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(vars.AdminUsername),
				Namespace: createdNamespace.Metadata.Name(),
				Labels:    pulumi.ToStringMap(labels),
			},
			Spec: v1beta2.KafkaUserSpecArgs{
				Authentication: v1beta2.KafkaUserSpecAuthenticationArgs{
					Type: pulumi.String("scram-sha-512"),
				},
			},
		}, pulumi.Parent(createdKafkaCluster))
	if err != nil {
		return errors.Wrap(err, "failed to create kafka admin user")
	}
	return nil
}
