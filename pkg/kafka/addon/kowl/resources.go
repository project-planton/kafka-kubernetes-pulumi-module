package kowl

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/addon/kowl/configmap"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/addon/kowl/deployment"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/addon/kowl/ingress"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context) error {
	ctx, err := configmap.Resources(ctx)
	if err != nil {
		return err
	}
	if err := deployment.Resources(ctx); err != nil {
		return errors.Wrap(err, "failed to add deployment resources")
	}
	if err := ingress.Resources(ctx); err != nil {
		return errors.Wrap(err, "failed to add ingress resources")
	}
	return nil
}
