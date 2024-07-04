package contextstate

import (
	code2cloudenvironmentmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/environment/model"
	kafkakubernetesstatemodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/model"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubecluster/enums/kubernetesworkloadingresstype"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
)

const (
	Key = "ctx-state"
)

type ContextState struct {
	Spec   *Spec
	Status *Status
}

type Spec struct {
	KubeProvider                   *kubernetes.Provider
	ResourceId                     string
	ResourceName                   string
	Labels                         map[string]string
	WorkspaceDir                   string
	NamespaceName                  string
	EnvironmentInfo                *code2cloudenvironmentmodel.ApiResourceEnvironmentInfo
	BrokerContainerSpec            *kafkakubernetesstatemodel.KafkaKubernetesSpecBrokerContainerSpec
	ZookeeperContainerSpec         *kafkakubernetesstatemodel.KafkaKubernetesSpecZookeeperContainerSpec
	SchemaRegistryContainerSpec    *kafkakubernetesstatemodel.KafkaKubernetesSpecSchemaRegistryContainerSpec
	Topics                         []*kafkakubernetesstatemodel.KafkaTopic
	IsSchemaRegistryEnabled        bool
	IsKowlDashboardEnabled         bool
	IsIngressEnabled               bool
	IngressType                    kubernetesworkloadingresstype.KubernetesWorkloadIngressType
	EndpointDomainName             string
	EnvDomainName                  string
	ExternalBootstrapHostname      string
	InternalBootstrapHostname      string
	ExternalSchemaRegistryHostname string
	InternalSchemaRegistryHostname string
	ExternalKowlDashboardHostname  string
	InternalKowlDashboardHostname  string
}

type Status struct {
	AddedResources *AddedResources
}

type AddedResources struct {
	Namespace     *kubernetescorev1.Namespace
	KowlConfigMap *kubernetescorev1.ConfigMap
}
