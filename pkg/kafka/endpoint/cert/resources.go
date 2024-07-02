package cert

import (
	"fmt"
	"path/filepath"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	"github.com/plantoncloud/environment-pulumi-blueprint/pkg/gcpgke/endpointdomains"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	k8sapimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Name = "kafka-ingress"
)

type Input struct {
	Namespace          *pulumikubernetescorev1.Namespace
	Labels             map[string]string
	NamespaceName      string
	Hostnames          []string
	WorkspaceDir       string
	EnvironmentName    string
	EndpointDomainName string
}

func Resources(ctx *pulumi.Context, input *Input) error {
	certObj := buildCertObject(Name, input)
	resourceName := fmt.Sprintf("cert-%s", certObj.Name)
	manifestPath := filepath.Join(input.WorkspaceDir, fmt.Sprintf("%s.yaml", resourceName))
	if err := manifest.Create(manifestPath, certObj); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", manifestPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, resourceName, &pulumik8syaml.ConfigFileArgs{File: manifestPath}, pulumi.Parent(input.Namespace))
	if err != nil {
		return errors.Wrap(err, "failed to add cert manifest")
	}
	return nil
}

/*
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:

	name: kafka-ingress
	namespace: rbitex-rbx-data-kafka-dev-main

spec:

	dnsNames:
	  - 'kc-planton-pcs-prod-mon-bootstrap.dev.planton.cloud'
	  - 'kafka-main-broker-b0.dev.planton.cloud'
	  - 'kafka-main-broker-b1.dev.planton.cloud'
	  - 'kafka-main-broker-b2.dev.planton.cloud'

#    - 'kc-planton-pcs-prod-mon-bootstrap.dev.planton.live'
#    - 'kc-planton-pcs-prod-mon-broker-b0.dev.planton.live'
#    - 'kc-planton-pcs-prod-mon-broker-b1.dev.planton.live'
#    - 'kc-planton-pcs-prod-mon-broker-b2.dev.planton.live'

	issuerRef:
	  kind: ClusterIssuer
	  name: letsencrypt-production
	secretName: cert-kafka-ingress
*/
func buildCertObject(certName string, input *Input) *certmanagerv1.Certificate {
	return &certmanagerv1.Certificate{
		TypeMeta: k8sapimachineryv1.TypeMeta{
			APIVersion: "cert-manager.io/v1",
			Kind:       "Certificate",
		},
		ObjectMeta: k8sapimachineryv1.ObjectMeta{
			Name:      certName,
			Namespace: input.NamespaceName,
			Labels:    input.Labels,
		},
		Spec: certmanagerv1.CertificateSpec{
			SecretName: GetCertSecretName(certName),
			DNSNames:   input.Hostnames,
			IssuerRef: cmmeta.ObjectReference{Kind: "ClusterIssuer",
				Name: endpointdomains.GetClusterIssuerName(input.EnvironmentName, input.EndpointDomainName)},
		},
	}
}

func GetCertSecretName(certName string) string {
	return fmt.Sprintf("cert-%s", certName)
}
