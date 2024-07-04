package addon

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/addon/kowl"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/addon/schemaregistry"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context) error {
	i := extractInput(ctx)
	if i.isSchemaRegistryEnabled {
		if err := schemaregistry.Resources(ctx); err != nil {
			return errors.Wrap(err, "failed to add schema-registry resources")
		}
	}

	if i.isKowlDashboardEnabled {
		if err := kowl.Resources(ctx); err != nil {
			return errors.Wrap(err, "failed to add kowl resources")
		}
	}
	return nil
}
