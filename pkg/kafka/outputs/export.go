package outputs

import (
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/adminuser"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"
	"github.com/plantoncloud/pulumi-blueprint-golang-commons/pkg/kubernetes/pulumikubernetesprovider"
	"github.com/plantoncloud/pulumi-blueprint-golang-commons/pkg/pulumi/pulumicustomoutput"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Export(ctx *pulumi.Context) error {
	var i = extractInput(ctx)

	ctx.Export(GetNamespaceNameOutputName(), pulumi.String(i.namespaceName))
	ctx.Export(GetSaslUsernameOutputName(), pulumi.String(adminuser.Username))

	ctx.Export(GetExternalBootstrapServerHostnameOutputName(), pulumi.String(i.externalBootstrapHostname))
	ctx.Export(GetInternalBootstrapServerHostnameOutputName(), pulumi.String(i.internalBootstrapHostname))

	ctx.Export(GetExternalSchemaRegistryUrlOutputName(), pulumi.Sprintf("https://%s", i.externalSchemaRegistryHostname))
	ctx.Export(GetInternalSchemaRegistryUrlOutputName(), pulumi.Sprintf("https://%s", i.internalSchemaRegistryHostname))

	ctx.Export(GetExternalKowlDashboardUrlOutputName(), pulumi.Sprintf("https://%s", i.externalKowlDashboardHostname))
	ctx.Export(GetInternalKowlDashboardUrlOutputName(), pulumi.Sprintf("https://%s", i.internalKowlDashboardHostname))

	return nil
}

func GetExternalBootstrapServerHostnameOutputName() string {
	return pulumicustomoutput.Name("external-bootstrap-server-hostname")
}

func GetInternalBootstrapServerHostnameOutputName() string {
	return pulumicustomoutput.Name("internal-bootstrap-server-hostname")
}

func GetExternalSchemaRegistryUrlOutputName() string {
	return pulumicustomoutput.Name("external-schema-registry-url")
}

func GetInternalSchemaRegistryUrlOutputName() string {
	return pulumicustomoutput.Name("internal-schema-registry-url")
}

func GetExternalKowlDashboardUrlOutputName() string {
	return pulumicustomoutput.Name("external-kowl-dashboard-url")
}

func GetInternalKowlDashboardUrlOutputName() string {
	return pulumicustomoutput.Name("internal-kowl-dashboard-url")
}

func GetNamespaceNameOutputName() string {
	return pulumikubernetesprovider.PulumiOutputName(kubernetescorev1.Namespace{}, englishword.EnglishWord_namespace.String())
}

func GetSaslUsernameOutputName() string {
	return pulumicustomoutput.Name(englishword.EnglishWord_username.String())
}
