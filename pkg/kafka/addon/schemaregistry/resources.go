package schemaregistry

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/addon/schemaregistry/deployment"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/addon/schemaregistry/ingress"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context) error {
	if err := deployment.Resources(ctx); err != nil {
		return errors.Wrap(err, "failed to add deployment resources")
	}
	if err := ingress.Resources(ctx); err != nil {
		return errors.Wrap(err, "failed to add ingress resources")
	}
	return nil
}
