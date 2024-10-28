package main

import (
	openshiftapiv1 "github.com/openshift/api/project/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func convertProjectToNamespace(project *openshiftapiv1.Project) corev1.Namespace {
	return corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   project.Name,
			Labels: project.Labels,
		},
	}
}
