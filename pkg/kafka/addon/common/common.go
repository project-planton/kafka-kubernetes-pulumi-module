package common

import corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"

func GetKafkaReadyCheckContainer() *corev1.ContainerArgs {
	return &corev1.ContainerArgs{
		Args:                     nil,
		Command:                  nil,
		Env:                      nil,
		EnvFrom:                  nil,
		Image:                    nil,
		ImagePullPolicy:          nil,
		Lifecycle:                nil,
		LivenessProbe:            nil,
		Name:                     nil,
		Ports:                    nil,
		ReadinessProbe:           nil,
		Resources:                nil,
		SecurityContext:          nil,
		StartupProbe:             nil,
		Stdin:                    nil,
		StdinOnce:                nil,
		TerminationMessagePath:   nil,
		TerminationMessagePolicy: nil,
		Tty:                      nil,
		VolumeDevices:            nil,
		VolumeMounts:             nil,
		WorkingDir:               nil,
	}
}
