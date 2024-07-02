package adminuser

import (
	"path/filepath"

	strimzitypes "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta2"
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	code2cloudv1deploykfcstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/stack/kubernetes/model"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"
	"github.com/plantoncloud/pulumi-stack-runner-go-sdk/pkg/name/output/custom"
	pulk8scv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
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

type Input struct {
	WorkspaceDir                     string
	KafkaKubernetesKubernetesStackInput *code2cloudv1deploykfcstackk8smodel.KafkaKubernetesKubernetesStackInput
	NamespaceName                    string
	Namespace                        *pulk8scv1.Namespace
	Labels                           map[string]string
}

func Resources(ctx *pulumi.Context, input *Input) error {
	strimziCacheLoc := filepath.Join(input.WorkspaceDir, "strimzi")
	adminUserYamlPath := filepath.Join(strimziCacheLoc, "admin-user.yaml")
	adminUserObject := buildAdminUserObject(input)
	if err := manifest.Create(adminUserYamlPath, adminUserObject); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", adminUserYamlPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, Username, &pulumik8syaml.ConfigFileArgs{
		File: adminUserYamlPath,
	}, pulumi.DependsOn([]pulumi.Resource{input.Namespace}), pulumi.Parent(input.Namespace))
	if err != nil {
		return errors.Wrap(err, "failed to add kubernetes config file")
	}
	exportOutputs(ctx)
	return nil
}

func exportOutputs(ctx *pulumi.Context) {
	ctx.Export(GetSaslUsernameOutputName(), pulumi.String(Username))
}

func GetSaslUsernameOutputName() string {
	return custom.Name(englishword.EnglishWord_username.String())
}

func buildAdminUserObject(input *Input) *strimzitypes.KafkaUser {
	//warning: if this does not match the name of the kafka resource, the user will not be created
	input.Labels[ClusterLabelKey] = input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Metadata.Id
	return &strimzitypes.KafkaUser{
		TypeMeta: v1.TypeMeta{
			Kind:       "KafkaUser",
			APIVersion: "kafka.strimzi.io/v1beta2",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      Username,
			Namespace: input.NamespaceName,
			Labels:    input.Labels,
		},
		Spec: &strimzitypes.KafkaUserSpec{
			Authentication: &strimzitypes.KafkaUserSpecAuthentication{
				Type: strimzitypes.KafkaUserSpecAuthenticationTypeScramSha512,
			},
		},
	}
}
