//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2023 Red Hat, Inc.
 *
 */

// Code generated by defaulter-gen. DO NOT EDIT.

package v1beta1

import (
	"encoding/json"

	runtime "k8s.io/apimachinery/pkg/runtime"
)

// RegisterDefaults adds defaulters functions to the given scheme.
// Public to allow building arbitrary schemes.
// All generated defaulters are covering - they call all nested defaulters.
func RegisterDefaults(scheme *runtime.Scheme) error {
	scheme.AddTypeDefaultingFunc(&HyperConverged{}, func(obj interface{}) { SetObjectDefaults_HyperConverged(obj.(*HyperConverged)) })
	scheme.AddTypeDefaultingFunc(&HyperConvergedList{}, func(obj interface{}) { SetObjectDefaults_HyperConvergedList(obj.(*HyperConvergedList)) })
	return nil
}

func SetObjectDefaults_HyperConverged(in *HyperConverged) {
	if in.Spec.FeatureGates.WithHostPassthroughCPU == nil {
		var ptrVar1 bool = false
		in.Spec.FeatureGates.WithHostPassthroughCPU = &ptrVar1
	}
	if in.Spec.FeatureGates.EnableCommonBootImageImport == nil {
		var ptrVar1 bool = true
		in.Spec.FeatureGates.EnableCommonBootImageImport = &ptrVar1
	}
	if in.Spec.FeatureGates.DeployTektonTaskResources == nil {
		var ptrVar1 bool = false
		in.Spec.FeatureGates.DeployTektonTaskResources = &ptrVar1
	}
	if in.Spec.FeatureGates.DeployKubeSecondaryDNS == nil {
		var ptrVar1 bool = false
		in.Spec.FeatureGates.DeployKubeSecondaryDNS = &ptrVar1
	}
	if in.Spec.FeatureGates.DisableMDevConfiguration == nil {
		var ptrVar1 bool = false
		in.Spec.FeatureGates.DisableMDevConfiguration = &ptrVar1
	}
	if in.Spec.LiveMigrationConfig.ParallelMigrationsPerCluster == nil {
		var ptrVar1 uint32 = 5
		in.Spec.LiveMigrationConfig.ParallelMigrationsPerCluster = &ptrVar1
	}
	if in.Spec.LiveMigrationConfig.ParallelOutboundMigrationsPerNode == nil {
		var ptrVar1 uint32 = 2
		in.Spec.LiveMigrationConfig.ParallelOutboundMigrationsPerNode = &ptrVar1
	}
	if in.Spec.LiveMigrationConfig.CompletionTimeoutPerGiB == nil {
		var ptrVar1 int64 = 800
		in.Spec.LiveMigrationConfig.CompletionTimeoutPerGiB = &ptrVar1
	}
	if in.Spec.LiveMigrationConfig.ProgressTimeout == nil {
		var ptrVar1 int64 = 150
		in.Spec.LiveMigrationConfig.ProgressTimeout = &ptrVar1
	}
	if in.Spec.LiveMigrationConfig.AllowAutoConverge == nil {
		var ptrVar1 bool = false
		in.Spec.LiveMigrationConfig.AllowAutoConverge = &ptrVar1
	}
	if in.Spec.LiveMigrationConfig.AllowPostCopy == nil {
		var ptrVar1 bool = false
		in.Spec.LiveMigrationConfig.AllowPostCopy = &ptrVar1
	}
	if in.Spec.CertConfig.CA.Duration == nil {
		if err := json.Unmarshal([]byte(`"48h0m0s"`), &in.Spec.CertConfig.CA.Duration); err != nil {
			panic(err)
		}
	}
	if in.Spec.CertConfig.CA.RenewBefore == nil {
		if err := json.Unmarshal([]byte(`"24h0m0s"`), &in.Spec.CertConfig.CA.RenewBefore); err != nil {
			panic(err)
		}
	}
	if in.Spec.CertConfig.Server.Duration == nil {
		if err := json.Unmarshal([]byte(`"24h0m0s"`), &in.Spec.CertConfig.Server.Duration); err != nil {
			panic(err)
		}
	}
	if in.Spec.CertConfig.Server.RenewBefore == nil {
		if err := json.Unmarshal([]byte(`"12h0m0s"`), &in.Spec.CertConfig.Server.RenewBefore); err != nil {
			panic(err)
		}
	}
	if in.Spec.WorkloadUpdateStrategy.WorkloadUpdateMethods == nil {
		if err := json.Unmarshal([]byte(`["LiveMigrate"]`), &in.Spec.WorkloadUpdateStrategy.WorkloadUpdateMethods); err != nil {
			panic(err)
		}
	}
	if in.Spec.WorkloadUpdateStrategy.BatchEvictionSize == nil {
		var ptrVar1 int = 10
		in.Spec.WorkloadUpdateStrategy.BatchEvictionSize = &ptrVar1
	}
	if in.Spec.WorkloadUpdateStrategy.BatchEvictionInterval == nil {
		if err := json.Unmarshal([]byte(`"1m0s"`), &in.Spec.WorkloadUpdateStrategy.BatchEvictionInterval); err != nil {
			panic(err)
		}
	}
	if in.Spec.UninstallStrategy == "" {
		in.Spec.UninstallStrategy = "BlockUninstallIfWorkloadsExist"
	}
}

func SetObjectDefaults_HyperConvergedList(in *HyperConvergedList) {
	for i := range in.Items {
		a := &in.Items[i]
		SetObjectDefaults_HyperConverged(a)
	}
}
