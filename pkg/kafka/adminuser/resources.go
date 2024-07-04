package adminuser

import (
	"path/filepath"

	strimzitypes "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta2"
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Username                  = "admin"
	ClusterLabelKey           = "strimzi.io/cluster"
	SaslPasswordSecretName    = "admin"
	SaslJaasConfigKeyInSecret = "sasl.jaas.config"
	SaslPasswordKeyInSecret   = "password"
)

func Resources(ctx *pulumi.Context) error {
	i := extractInput(ctx)
	strimziCacheLoc := filepath.Join(i.workspaceDir, "strimzi")
	adminUserYamlPath := filepath.Join(strimziCacheLoc, "admin-user.yaml")
	adminUserObject := buildAdminUserObject(i)
	if err := manifest.Create(adminUserYamlPath, adminUserObject); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", adminUserYamlPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, Username, &pulumik8syaml.ConfigFileArgs{
		File: adminUserYamlPath,
	}, pulumi.DependsOn([]pulumi.Resource{i.namespace}), pulumi.Parent(i.namespace))
	if err != nil {
		return errors.Wrap(err, "failed to add kubernetes config file")
	}
	return nil
}

func buildAdminUserObject(i *input) *strimzitypes.KafkaUser {
	//warning: if this does not match the name of the kafka resource, the user will not be created
	i.labels[ClusterLabelKey] = i.resourceId
	return &strimzitypes.KafkaUser{
		TypeMeta: v1.TypeMeta{
			Kind:       "KafkaUser",
			APIVersion: "kafka.strimzi.io/v1beta2",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      Username,
			Namespace: i.namespaceName,
			Labels:    i.labels,
		},
		Spec: &strimzitypes.KafkaUserSpec{
			Authentication: &strimzitypes.KafkaUserSpecAuthentication{
				Type: strimzitypes.KafkaUserSpecAuthenticationTypeScramSha512,
			},
		},
	}
}
