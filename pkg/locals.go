package pkg

import (
	"fmt"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-module/pkg/outputs"
	kafkakubernetesmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/kafkakubernetes/model"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	IngressCertClusterIssuerName string
	// bootstrap
	IngressBootstrapCertSecretName   string
	IngressExternalBootstrapHostname string
	IngressInternalBootstrapHostname string
	IngressExternalBrokerHostnames   []string
	IngressInternalBrokerHostnames   []string
	IngressHostnames                 []string
	BootstrapKubeServiceFqdn         string
	BootstrapKubeServiceName         string

	// kowl dashboard
	IngressKowlCertSecretName   string
	IngressExternalKowlHostname string
	IngressInternalKowlHostname string
	IngressKowlHostnames        []string
	KowlKubeServiceFqdn         string

	// schema registry
	IngressSchemaRegistryCertSecretName   string
	IngressExternalSchemaRegistryHostname string
	IngressInternalSchemaRegistryHostname string
	IngressSchemaRegistryHostnames        []string
	SchemaRegistryKubeServiceFqdn         string

	Namespace       string
	KafkaKubernetes *kafkakubernetesmodel.KafkaKubernetes
}

func initializeLocals(ctx *pulumi.Context, stackInput *kafkakubernetesmodel.KafkaKubernetesStackInput) *Locals {
	locals := &Locals{}

	ctx.Export(outputs.KafkaSaslUsername, pulumi.String(vars.AdminUsername))

	//assign value for the locals variable to make it available across the project
	locals.KafkaKubernetes = stackInput.ApiResource

	kafkaKubernetes := stackInput.ApiResource

	//decide on the namespace
	locals.Namespace = kafkaKubernetes.Metadata.Id
	ctx.Export(outputs.Namespace, pulumi.String(locals.Namespace))

	locals.BootstrapKubeServiceName = fmt.Sprintf("%s-kafka-%s-bootstrap", kafkaKubernetes.Metadata.Id, vars.ExternalPublicListenerName)

	locals.BootstrapKubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", locals.BootstrapKubeServiceName, locals.Namespace)

	// schema registry related locals data
	if locals.KafkaKubernetes.Spec.SchemaRegistryContainer != nil &&
		locals.KafkaKubernetes.Spec.SchemaRegistryContainer.IsEnabled {

		locals.IngressSchemaRegistryCertSecretName = fmt.Sprintf("schema-registry-%s", kafkaKubernetes.Metadata.Id)

		locals.IngressExternalSchemaRegistryHostname = fmt.Sprintf("%s-schema-registry.%s", kafkaKubernetes.Metadata.Id, kafkaKubernetes.Spec.Ingress.EndpointDomainName)

		locals.IngressInternalSchemaRegistryHostname = fmt.Sprintf("%s-schema-registry-internal.%s", kafkaKubernetes.Metadata.Id, kafkaKubernetes.Spec.Ingress.EndpointDomainName)

		ctx.Export(outputs.IngressExternalSchemaRegistryUrl, pulumi.Sprintf("https://%s", locals.IngressExternalSchemaRegistryHostname))
		ctx.Export(outputs.IngressInternalSchemaRegistryUrl, pulumi.Sprintf("https://%s", locals.IngressInternalSchemaRegistryHostname))

		locals.IngressSchemaRegistryHostnames = []string{
			locals.IngressExternalSchemaRegistryHostname,
			locals.IngressInternalSchemaRegistryHostname,
		}
		locals.SchemaRegistryKubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", vars.SchemaRegistryKubeServiceName, locals.Namespace)
	}

	// kowl related locals data
	if locals.KafkaKubernetes.Spec.IsKowlDashboardEnabled {

		locals.IngressKowlCertSecretName = fmt.Sprintf("kowl-%s", kafkaKubernetes.Metadata.Id)

		locals.IngressExternalKowlHostname = fmt.Sprintf("%s-kowl.%s", kafkaKubernetes.Metadata.Id, kafkaKubernetes.Spec.Ingress.EndpointDomainName)

		locals.IngressInternalKowlHostname = fmt.Sprintf("%s-kowl-internal.%s", kafkaKubernetes.Metadata.Id, kafkaKubernetes.Spec.Ingress.EndpointDomainName)

		ctx.Export(outputs.IngressExternalKowlUrl, pulumi.Sprintf("https://%s", locals.IngressExternalKowlHostname))
		ctx.Export(outputs.IngressInternalKowlUrl, pulumi.Sprintf("https://%s", locals.IngressInternalKowlHostname))

		locals.IngressSchemaRegistryHostnames = []string{
			locals.IngressExternalKowlHostname,
			locals.IngressInternalKowlHostname,
		}
		locals.KowlKubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", vars.KowlKubeServiceName, locals.Namespace)
	}

	if kafkaKubernetes.Spec.Ingress == nil ||
		!kafkaKubernetes.Spec.Ingress.IsEnabled ||
		kafkaKubernetes.Spec.Ingress.EndpointDomainName == "" {
		return locals
	}

	locals.IngressExternalBootstrapHostname = fmt.Sprintf("%s-bootstrap.%s", kafkaKubernetes.Metadata.Id, kafkaKubernetes.Spec.Ingress.EndpointDomainName)

	locals.IngressInternalBootstrapHostname = fmt.Sprintf("%s-bootstrap-internal.%s", kafkaKubernetes.Metadata.Id, kafkaKubernetes.Spec.Ingress.EndpointDomainName)

	ctx.Export(outputs.IngressExternalBootStrapHostname, pulumi.String(locals.IngressExternalBootstrapHostname))
	ctx.Export(outputs.IngressInternalBootStrapHostname, pulumi.String(locals.IngressInternalBootstrapHostname))

	// Creating internal broker hostnames
	ingressInternalBrokerHostnames := make([]string, int(kafkaKubernetes.Spec.BrokerContainer.Replicas))
	for i := 0; i < int(kafkaKubernetes.Spec.BrokerContainer.Replicas); i++ {
		ingressInternalBrokerHostnames[i] = fmt.Sprintf("%s-broker-b%d-internal.%s", kafkaKubernetes.Metadata.Id, i, kafkaKubernetes.Spec.Ingress.EndpointDomainName)
	}
	locals.IngressInternalBrokerHostnames = ingressInternalBrokerHostnames

	// Creating external broker hostnames
	ingressExternalBrokerHostnames := make([]string, int(kafkaKubernetes.Spec.BrokerContainer.Replicas))
	for i := 0; i < int(kafkaKubernetes.Spec.BrokerContainer.Replicas); i++ {
		ingressExternalBrokerHostnames[i] = fmt.Sprintf("%s-broker-b%d.%s", kafkaKubernetes.Metadata.Id, i, kafkaKubernetes.Spec.Ingress.EndpointDomainName)
	}
	locals.IngressExternalBrokerHostnames = ingressExternalBrokerHostnames

	var ingressHostnames = []string{
		locals.IngressInternalBootstrapHostname,
		locals.IngressExternalBootstrapHostname,
	}

	ingressHostnames = append(ingressHostnames, locals.IngressInternalBrokerHostnames...)
	ingressHostnames = append(ingressHostnames, locals.IngressExternalBrokerHostnames...)
	locals.IngressHostnames = ingressHostnames

	//export ingress hostnames
	//ctx.Export(outputs.IngressExternalHostname, pulumi.String(locals.IngressExternalHostname))
	//ctx.Export(outputs.IngressInternalHostname, pulumi.String(locals.IngressInternalHostname))

	//note: a ClusterIssuer resource should have already exist on the kubernetes-cluster.
	//this is typically taken care of by the kubernetes cluster administrator.
	//if the kubernetes-cluster is created using Planton Cloud, then the cluster-issuer name will be
	//same as the ingress-domain-name as long as the same ingress-domain-name is added to the list of
	//ingress-domain-names for the GkeCluster/EksCluster/AksCluster spec.
	locals.IngressCertClusterIssuerName = kafkaKubernetes.Spec.Ingress.EndpointDomainName

	locals.IngressBootstrapCertSecretName = kafkaKubernetes.Metadata.Id

	locals.IngressBootstrapCertSecretName = kafkaKubernetes.Metadata.Id

	return locals
}
