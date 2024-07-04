package cert

import (
	"fmt"
	"path/filepath"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	"github.com/plantoncloud/environment-pulumi-blueprint/pkg/gcpgke/endpointdomains"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	k8sapimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Name = "kafka-ingress"
)

func Resources(ctx *pulumi.Context) error {
	i := extractInput(ctx)
	certObj := buildCertObject(i)
	resourceName := fmt.Sprintf("cert-%s", certObj.Name)
	manifestPath := filepath.Join(i.workspaceDir, fmt.Sprintf("%s.yaml", resourceName))
	if err := manifest.Create(manifestPath, certObj); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", manifestPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, resourceName, &pulumik8syaml.ConfigFileArgs{File: manifestPath}, pulumi.Parent(i.namespace))
	if err != nil {
		return errors.Wrap(err, "failed to add cert manifest")
	}
	return nil
}

func buildCertObject(i *input) *certmanagerv1.Certificate {
	return &certmanagerv1.Certificate{
		TypeMeta: k8sapimachineryv1.TypeMeta{
			APIVersion: "cert-manager.io/v1",
			Kind:       "Certificate",
		},
		ObjectMeta: k8sapimachineryv1.ObjectMeta{
			Name:      Name,
			Namespace: i.namespaceName,
			Labels:    i.labels,
		},
		Spec: certmanagerv1.CertificateSpec{
			SecretName: GetCertSecretName(Name),
			DNSNames:   i.hostnames,
			IssuerRef: cmmeta.ObjectReference{Kind: "ClusterIssuer",
				Name: endpointdomains.GetClusterIssuerName(i.environmentName, i.endpointDomainName)},
		},
	}
}

func GetCertSecretName(certName string) string {
	return fmt.Sprintf("cert-%s", certName)
}
