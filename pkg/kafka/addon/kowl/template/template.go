package template

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/util/file"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context) ([]byte, error) {
	i := extractInput(ctx)

	kowlConfig, err := file.RenderTemplate(i, kowlConfigFileTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render kowl config file")
	}
	return kowlConfig, nil
}
