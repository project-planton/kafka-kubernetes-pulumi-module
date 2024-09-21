package main

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-module/pkg"
	"github.com/plantoncloud/planton/apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/kafkakubernetes"
	"github.com/plantoncloud/pulumi-module-golang-commons/pkg/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &kafkakubernetes.KafkaKubernetesStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return pkg.Resources(ctx, stackInput)
	})
}