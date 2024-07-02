package contextstate

import (
	code2cloudenvironmentmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/environment/model"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubecluster/enums/kubernetesworkloadingresstype"
	plantoncloudpostgresdbmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/postgreskubernetes/model"
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
	KubeProvider       *kubernetes.Provider
	ResourceId         string
	ResourceName       string
	Labels             map[string]string
	WorkspaceDir       string
	NamespaceName      string
	EnvironmentInfo    *code2cloudenvironmentmodel.ApiResourceEnvironmentInfo
	IsIngressEnabled   bool
	IngressType        kubernetesworkloadingresstype.KubernetesWorkloadIngressType
	EndpointDomainName string
	EnvDomainName      string
	ContainerSpec      *plantoncloudpostgresdbmodel.PostgresKubernetesSpecContainerSpec
	CustomHelmValues   map[string]string
	InternalHostname   string
	ExternalHostname   string
	KubeServiceName    string
	KubeLocalEndpoint  string
	CertSecretName     string
}

type Status struct {
	AddedResources *AddedResources
}

type AddedResources struct {
	Namespace                   *kubernetescorev1.Namespace
	LoadBalancerExternalService *kubernetescorev1.Service
	LoadBalancerInternalService *kubernetescorev1.Service
}
