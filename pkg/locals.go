package pkg

import (
	"fmt"
	kafkakubernetesmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/kafkakubernetes/model"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	IngressCertClusterIssuerName string
	IngressCertSecretName        string
	// bootstrap
	BootstrapKubeServiceFqdn         string
	BootstrapKubeServiceName         string
	IngressExternalBootstrapHostname string
	IngressInternalBootstrapHostname string

	IngressExternalBrokerHostnames []string
	IngressInternalBrokerHostnames []string

	IngressHostnames []string

	// kowl dashboard
	IngressExternalKowlDashboardHostname string
	IngressInternalKowlDashboardHostname string
	// schema registry
	IngressExternalSchemaRegistryHostname string
	IngressInternalSchemaRegistryHostname string
	IngressSchemaRegistryHostnames        []string
	SchemaRegistryKubeServiceFqdn         string

	Namespace       string
	KafkaKubernetes *kafkakubernetesmodel.KafkaKubernetes
}

func initializeLocals(ctx *pulumi.Context, stackInput *kafkakubernetesmodel.KafkaKubernetesStackInput) *Locals {
	locals := &Locals{}
	//assign value for the locals variable to make it available across the project
	locals.KafkaKubernetes = stackInput.ApiResource

	kafkaKubernetes := stackInput.ApiResource

	//decide on the namespace
	locals.Namespace = kafkaKubernetes.Metadata.Id

	locals.BootstrapKubeServiceName = fmt.Sprintf("%s-kafka-%s-bootstrap", kafkaKubernetes.Metadata.Id, vars.ExternalPublicListenerName)

	//export kubernetes service name
	//ctx.Export(outputs.Service, pulumi.String(locals.KubeServiceName))

	locals.BootstrapKubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", locals.BootstrapKubeServiceName, locals.Namespace)

	//export kubernetes endpoint
	//ctx.Export(outputs.KubeEndpoint, pulumi.String(locals.KubeServiceFqdn))

	if kafkaKubernetes.Spec.Ingress == nil ||
		!kafkaKubernetes.Spec.Ingress.IsEnabled ||
		kafkaKubernetes.Spec.Ingress.EndpointDomainName == "" {
		return locals
	}

	locals.IngressExternalBootstrapHostname = fmt.Sprintf("%s-bootstrap.%s", kafkaKubernetes.Metadata.Id, kafkaKubernetes.Spec.Ingress.EndpointDomainName)

	locals.IngressInternalBootstrapHostname = fmt.Sprintf("%s-bootstrap-internal.%s", kafkaKubernetes.Metadata.Id, kafkaKubernetes.Spec.Ingress.EndpointDomainName)

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

	locals.IngressCertSecretName = kafkaKubernetes.Metadata.Id

	locals.IngressCertSecretName = kafkaKubernetes.Metadata.Id

	return locals
}
