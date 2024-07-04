package istio

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/network/istio/cert"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/network/istio/virtualservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context) error {
	if err := cert.Resources(ctx); err != nil {
		return errors.Wrap(err, "failed to add cert resources")
	}
	if err := virtualservice.Resources(ctx); err != nil {
		return errors.Wrap(err, "failed to add virtual service resources")
	}
	return nil
}
