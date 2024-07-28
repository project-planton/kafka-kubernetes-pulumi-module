package pkg

import (
	strimzitypes "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta2"
	"github.com/pkg/errors"
	"github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/strimzioperator/kafka/v1beta2"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func kafkaAdminUser(ctx *pulumi.Context, locals *Locals, createdNamespace *kubernetescorev1.Namespace,
	labels map[string]string) error {

	_, err := v1beta2.NewKafkaUser(ctx, "admin-user", &v1beta2.KafkaUserArgs{
		Kind:       pulumi.String("KafkaUser"),
		ApiVersion: pulumi.String("kafka.strimzi.io/v1beta2"),
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.KafkaKubernetes.Metadata.Id),
			Namespace: createdNamespace.Metadata.Name(),
			Labels:    pulumi.ToStringMap(labels),
		},
		Spec: v1beta2.KafkaUserSpecArgs{
			Authentication: v1beta2.KafkaUserSpecAuthenticationArgs{
				Type: pulumi.String(strimzitypes.KafkaUserSpecAuthenticationTypeScramSha512),
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to create kafka admin user")
	}
	return nil
}
