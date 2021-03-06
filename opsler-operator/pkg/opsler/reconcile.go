package opsler

import (
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	api "github.com/opsler/opsler/opsler-operator/pkg/apis/opsler/v1alpha1"
	"github.com/opsler/opsler/opsler-operator/pkg/istio"
	"github.com/opsler/opsler/opsler-operator/pkg/models"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Reconcile() (err error) {
	listOptions := sdk.WithListOptions(&metav1.ListOptions{
		IncludeUninitialized: true,
	})
	namespace := "default"

	entrypointList := api.EntrypointList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Entrypoint",
			APIVersion: "opsler.com/v1alpha1",
		},
	}

	if err := sdk.List(namespace, &entrypointList, listOptions); err != nil {
		logrus.Errorf("Query failed: %v", err)
		return err
	}

	logrus.Infof("Received entrypoints: %+v", entrypointList)

	virtualEnvironmentList := api.VirtualEnvironmentList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualEnvironment",
			APIVersion: "opsler.com/v1alpha1",
		},
	}
	if err := sdk.List(namespace, &virtualEnvironmentList, listOptions); err != nil {
		logrus.Errorf("Query failed: %v", err)
		return err
	}

	logrus.Infof("Received virtualEnvironments: %+v", virtualEnvironmentList)

	targetingList := api.TargetingList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Targeting",
			APIVersion: "opsler.com/v1alpha1",
		},
	}
	if err := sdk.List(namespace, &targetingList, listOptions); err != nil {
		logrus.Errorf("Query failed: %+v", err)
		return err
	}

	logrus.Infof("Received targetings: %v", targetingList)

	entrypointFlows := combine(virtualEnvironmentList, targetingList, entrypointList)
	istio.Apply(entrypointFlows, namespace)
	//sdk.Create()

	return nil
}

func combine(virtualEnvironmentList api.VirtualEnvironmentList, targetingList api.TargetingList, entrypointList api.EntrypointList) []models.EntrypointFlow {
	entrypointFlows := make([]models.EntrypointFlow, 0)
	for _, entrypoint := range entrypointList.Items {
		defaultVirtualEnvironment, ok := findVirtualEnvironment(entrypoint.Spec.DefaultVirtualEnvironment, virtualEnvironmentList.Items)
		if ok {
			targetings := getAllTargetingsByEntrypoint(entrypoint.ObjectMeta.Name, targetingList.Items)
			entrypointFlows = append(entrypointFlows, models.EntrypointFlow{
				Entrypoint:                entrypoint,
				DefaultVirtualEnvironment: defaultVirtualEnvironment,
				Targetings:                combineTargetingToVirtualEnvironments(targetings, virtualEnvironmentList.Items)})
		} else {
			// TODO: Notify that we are waiting for virtual env to be created
		}
	}
	return entrypointFlows
}

func combineTargetingToVirtualEnvironments(targetings []api.Targeting, virtualEnvironments []api.VirtualEnvironment) []models.TargetingFlow {
	targetingFlows := make([]models.TargetingFlow, 0)
	for _, targeting := range targetings {
		virtualEnvironment, ok := findVirtualEnvironment(targeting.Spec.VirtualEnvironment, virtualEnvironments)
		if ok {
			targetingFlows = append(targetingFlows, models.TargetingFlow{
				Targeting:          targeting,
				VirtualEnvironment: virtualEnvironment})
		}
	}
	return targetingFlows
}

func getAllTargetingsByEntrypoint(entrypointName string, targetings []api.Targeting) []api.Targeting {
	targetingsOfEntrypoint := make([]api.Targeting, 0)
	for _, targeting := range targetings {
		if targeting.Spec.Entrypoint == entrypointName {
			targetingsOfEntrypoint = append(targetingsOfEntrypoint, targeting)
		}
	}
	return targetingsOfEntrypoint
}

func findVirtualEnvironment(name string, virtualEnvironments []api.VirtualEnvironment) (api.VirtualEnvironment, bool) {
	for _, virtualEnvironment := range virtualEnvironments {
		if virtualEnvironment.ObjectMeta.Name == name {
			return virtualEnvironment, true
		}
	}
	return api.VirtualEnvironment{}, false
}
