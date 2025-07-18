/*
Copyright 2024 The Kubeflow Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package core

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	batchv1ac "k8s.io/client-go/applyconfigurations/batch/v1"
	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/klog/v2/ktesting"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
	jobsetv1alpha2ac "sigs.k8s.io/jobset/client-go/applyconfiguration/jobset/v1alpha2"
	jobsetconsts "sigs.k8s.io/jobset/pkg/constants"
	schedulerpluginsv1alpha1 "sigs.k8s.io/scheduler-plugins/apis/scheduling/v1alpha1"

	trainer "github.com/kubeflow/trainer/v2/pkg/apis/trainer/v1alpha1"
	"github.com/kubeflow/trainer/v2/pkg/apply"
	"github.com/kubeflow/trainer/v2/pkg/constants"
	"github.com/kubeflow/trainer/v2/pkg/runtime"
	"github.com/kubeflow/trainer/v2/pkg/runtime/framework"
	fwkplugins "github.com/kubeflow/trainer/v2/pkg/runtime/framework/plugins"
	"github.com/kubeflow/trainer/v2/pkg/runtime/framework/plugins/coscheduling"
	"github.com/kubeflow/trainer/v2/pkg/runtime/framework/plugins/jobset"
	jobsetplgconsts "github.com/kubeflow/trainer/v2/pkg/runtime/framework/plugins/jobset/constants"
	"github.com/kubeflow/trainer/v2/pkg/runtime/framework/plugins/mpi"
	"github.com/kubeflow/trainer/v2/pkg/runtime/framework/plugins/plainml"
	"github.com/kubeflow/trainer/v2/pkg/runtime/framework/plugins/torch"
	testingutil "github.com/kubeflow/trainer/v2/pkg/util/testing"
)

// TODO: We should introduce mock plugins and use plugins in this framework testing.
// After we migrate the actual plugins to mock one for testing data,
// we can delegate the actual plugin testing to each plugin directories, and implement detailed unit testing.

func TestNew(t *testing.T) {
	cases := map[string]struct {
		registry                                                               fwkplugins.Registry
		emptyCoSchedulingIndexerTrainingRuntimeContainerRuntimeClassKey        bool
		emptyCoSchedulingIndexerClusterTrainingRuntimeContainerRuntimeClassKey bool
		wantFramework                                                          *Framework
		wantError                                                              error
	}{
		"positive case": {
			registry: fwkplugins.NewRegistry(),
			wantFramework: &Framework{
				registry: fwkplugins.NewRegistry(),
				plugins: map[string]framework.Plugin{
					coscheduling.Name: &coscheduling.CoScheduling{},
					mpi.Name:          &mpi.MPI{},
					plainml.Name:      &plainml.PlainML{},
					torch.Name:        &torch.Torch{},
					jobset.Name:       &jobset.JobSet{},
				},
				enforceMLPlugins: []framework.EnforceMLPolicyPlugin{
					&mpi.MPI{},
					&plainml.PlainML{},
					&torch.Torch{},
				},
				enforcePodGroupPolicyPlugins: []framework.EnforcePodGroupPolicyPlugin{
					&coscheduling.CoScheduling{},
				},
				customValidationPlugins: []framework.CustomValidationPlugin{
					&mpi.MPI{},
					&torch.Torch{},
					&jobset.JobSet{},
				},
				watchExtensionPlugins: []framework.WatchExtensionPlugin{
					&coscheduling.CoScheduling{},
					&jobset.JobSet{},
					&mpi.MPI{},
				},
				podNetworkPlugins: []framework.PodNetworkPlugin{
					&jobset.JobSet{},
				},
				componentBuilderPlugins: []framework.ComponentBuilderPlugin{
					&coscheduling.CoScheduling{},
					&jobset.JobSet{},
					&mpi.MPI{},
				},
				terminalConditionPlugins: []framework.TerminalConditionPlugin{
					&jobset.JobSet{},
				},
			},
		},
		"indexer key for trainingRuntime and runtimeClass is an empty": {
			registry: fwkplugins.Registry{
				coscheduling.Name: coscheduling.New,
			},
			emptyCoSchedulingIndexerTrainingRuntimeContainerRuntimeClassKey: true,
			wantError: coscheduling.ErrorCanNotSetupTrainingRuntimeRuntimeClassIndexer,
		},
		"indexer key for clusterTrainingRuntime and runtimeClass is an empty": {
			registry: fwkplugins.Registry{
				coscheduling.Name: coscheduling.New,
			},
			emptyCoSchedulingIndexerClusterTrainingRuntimeContainerRuntimeClassKey: true,
			wantError: coscheduling.ErrorCanNotSetupClusterTrainingRuntimeRuntimeClassIndexer,
		},
	}
	cmpOpts := []cmp.Option{
		cmp.AllowUnexported(Framework{}),
		cmpopts.IgnoreUnexported(coscheduling.CoScheduling{}, mpi.MPI{}, plainml.PlainML{}, torch.Torch{}, jobset.JobSet{}),
		cmpopts.IgnoreFields(coscheduling.CoScheduling{}, "client"),
		cmpopts.IgnoreFields(jobset.JobSet{}, "client"),
		cmpopts.IgnoreTypes(apiruntime.Scheme{}, meta.DefaultRESTMapper{}, fwkplugins.Registry{}),
		cmpopts.SortMaps(func(a, b string) bool { return a < b }),
		cmpopts.SortSlices(func(a, b framework.Plugin) bool { return a.Name() < b.Name() }),
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			if tc.emptyCoSchedulingIndexerTrainingRuntimeContainerRuntimeClassKey {
				originTrainingRuntimeRuntimeKey := coscheduling.TrainingRuntimeContainerRuntimeClassKey
				coscheduling.TrainingRuntimeContainerRuntimeClassKey = ""
				t.Cleanup(func() {
					coscheduling.TrainingRuntimeContainerRuntimeClassKey = originTrainingRuntimeRuntimeKey
				})
			}
			if tc.emptyCoSchedulingIndexerClusterTrainingRuntimeContainerRuntimeClassKey {
				originClusterTrainingRuntimeKey := coscheduling.ClusterTrainingRuntimeContainerRuntimeClassKey
				coscheduling.ClusterTrainingRuntimeContainerRuntimeClassKey = ""
				t.Cleanup(func() {
					coscheduling.ClusterTrainingRuntimeContainerRuntimeClassKey = originClusterTrainingRuntimeKey
				})
			}
			clientBuilder := testingutil.NewClientBuilder()
			fwk, err := New(ctx, clientBuilder.Build(), tc.registry, testingutil.AsIndex(clientBuilder))
			if diff := cmp.Diff(tc.wantError, err, cmpopts.EquateErrors()); len(diff) != 0 {
				t.Errorf("Unexpected errors (-want,+got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantFramework, fwk, cmpOpts...); len(diff) != 0 {
				t.Errorf("Unexpected framework (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestRunEnforceMLPolicyPlugins(t *testing.T) {
	cases := map[string]struct {
		registry        fwkplugins.Registry
		runtimeInfo     *runtime.Info
		trainJob        *trainer.TrainJob
		wantRuntimeInfo *runtime.Info
		wantError       error
	}{
		"plainml MLPolicy is applied to runtime.Info, TrainJob doesn't have numNodes": {
			registry: fwkplugins.NewRegistry(),
			runtimeInfo: runtime.NewInfo(
				runtime.WithMLPolicySource(testingutil.MakeMLPolicyWrapper().Obj()),
				runtime.WithPodSet(constants.DatasetInitializer, ptr.To(constants.DatasetInitializer), 1, corev1.PodSpec{}, corev1ac.PodSpec().
					WithContainers(
						corev1ac.Container().WithName(constants.DatasetInitializer),
					),
				),
				runtime.WithPodSet(constants.ModelInitializer, ptr.To(constants.ModelInitializer), 1, corev1.PodSpec{}, corev1ac.PodSpec().
					WithContainers(
						corev1ac.Container().WithName(constants.ModelInitializer),
					),
				),
				runtime.WithPodSet(constants.Node, ptr.To(constants.AncestorTrainer), 10, corev1.PodSpec{}, corev1ac.PodSpec().
					WithContainers(corev1ac.Container().WithName(constants.Node)),
				),
			),
			trainJob: &trainer.TrainJob{
				Spec: trainer.TrainJobSpec{},
			},
			wantRuntimeInfo: &runtime.Info{
				RuntimePolicy: runtime.RuntimePolicy{
					MLPolicySource: testingutil.MakeMLPolicySourceWrapper().Obj(),
				},
				TemplateSpec: runtime.TemplateSpec{
					PodSets: []runtime.PodSet{
						{
							Name:     constants.DatasetInitializer,
							Ancestor: ptr.To(constants.DatasetInitializer),
							Count:    ptr.To[int32](1),
							Containers: []runtime.Container{
								{
									Name: constants.DatasetInitializer,
								},
							},
						},
						{
							Name:     constants.ModelInitializer,
							Ancestor: ptr.To(constants.ModelInitializer),
							Count:    ptr.To[int32](1),
							Containers: []runtime.Container{
								{
									Name: constants.ModelInitializer,
								},
							},
						},
						{
							Name:     constants.Node,
							Ancestor: ptr.To(constants.AncestorTrainer),
							Count:    ptr.To[int32](10),
							Containers: []runtime.Container{{
								Name: constants.Node,
							}},
						},
					},
				},
				Scheduler: &runtime.Scheduler{PodLabels: make(map[string]string)},
			},
		},
		"plainml MLPolicy is applied to runtime.Info, TrainJob has numNodes": {
			registry: fwkplugins.NewRegistry(),
			runtimeInfo: runtime.NewInfo(
				runtime.WithMLPolicySource(
					testingutil.MakeMLPolicyWrapper().Obj(),
				),
				runtime.WithPodSet(constants.DatasetInitializer, ptr.To(constants.DatasetInitializer), 1, corev1.PodSpec{}, corev1ac.PodSpec().
					WithContainers(
						corev1ac.Container().WithName(constants.DatasetInitializer),
					),
				),
				runtime.WithPodSet(constants.ModelInitializer, ptr.To(constants.ModelInitializer), 1, corev1.PodSpec{}, corev1ac.PodSpec().
					WithContainers(
						corev1ac.Container().WithName(constants.ModelInitializer),
					),
				),
				runtime.WithPodSet(constants.Node, ptr.To(constants.AncestorTrainer), 10, corev1.PodSpec{}, corev1ac.PodSpec().
					WithContainers(corev1ac.Container().WithName(constants.Node)),
				),
			),
			trainJob: &trainer.TrainJob{
				Spec: trainer.TrainJobSpec{
					Trainer: &trainer.Trainer{
						NumNodes: ptr.To[int32](30),
					},
				},
			},
			wantRuntimeInfo: &runtime.Info{
				RuntimePolicy: runtime.RuntimePolicy{
					MLPolicySource: testingutil.MakeMLPolicySourceWrapper().Obj(),
				},
				TemplateSpec: runtime.TemplateSpec{
					PodSets: []runtime.PodSet{
						{
							Name:     constants.DatasetInitializer,
							Ancestor: ptr.To(constants.DatasetInitializer),
							Count:    ptr.To[int32](1),
							Containers: []runtime.Container{
								{
									Name: constants.DatasetInitializer,
								},
							},
						},
						{
							Name:     constants.ModelInitializer,
							Ancestor: ptr.To(constants.ModelInitializer),
							Count:    ptr.To[int32](1),
							Containers: []runtime.Container{
								{
									Name: constants.ModelInitializer,
								},
							},
						},
						{
							Name:     constants.Node,
							Ancestor: ptr.To(constants.AncestorTrainer),
							Count:    ptr.To[int32](30),
							Containers: []runtime.Container{{
								Name: constants.Node,
							}},
						},
					},
				},
				Scheduler: &runtime.Scheduler{PodLabels: make(map[string]string)},
			},
		},
		"registry is empty": {
			runtimeInfo: &runtime.Info{
				Scheduler: &runtime.Scheduler{PodLabels: make(map[string]string)},
			},
			wantRuntimeInfo: &runtime.Info{
				Scheduler: &runtime.Scheduler{PodLabels: make(map[string]string)},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			clientBuilder := testingutil.NewClientBuilder()

			fwk, err := New(ctx, clientBuilder.Build(), tc.registry, testingutil.AsIndex(clientBuilder))
			if err != nil {
				t.Fatal(err)
			}
			err = fwk.RunEnforceMLPolicyPlugins(tc.runtimeInfo, tc.trainJob)
			if diff := cmp.Diff(tc.wantError, err, cmpopts.EquateErrors()); len(diff) != 0 {
				t.Errorf("Unexpected error (-want,+got): %s", diff)
			}
			if diff := cmp.Diff(tc.wantRuntimeInfo, tc.runtimeInfo, cmpopts.EquateEmpty()); len(diff) != 0 {
				t.Errorf("Unexpected runtime.Info (-want,+got): %s", diff)
			}
		})
	}
}

func TestRunEnforcePodGroupPolicyPlugins(t *testing.T) {
	cases := map[string]struct {
		registry        fwkplugins.Registry
		runtimeInfo     *runtime.Info
		trainJob        *trainer.TrainJob
		wantRuntimeInfo *runtime.Info
		wantError       error
	}{
		"coscheduling plugin is applied to runtime.Info": {
			registry: fwkplugins.NewRegistry(),
			runtimeInfo: &runtime.Info{
				RuntimePolicy: runtime.RuntimePolicy{
					PodGroupPolicy: &trainer.PodGroupPolicy{
						PodGroupPolicySource: trainer.PodGroupPolicySource{
							Coscheduling: &trainer.CoschedulingPodGroupPolicySource{
								ScheduleTimeoutSeconds: ptr.To[int32](99),
							},
						},
					},
				},
				Scheduler: &runtime.Scheduler{},
			},
			trainJob: &trainer.TrainJob{ObjectMeta: metav1.ObjectMeta{Name: "test-job", Namespace: metav1.NamespaceDefault}},
			wantRuntimeInfo: &runtime.Info{
				RuntimePolicy: runtime.RuntimePolicy{
					PodGroupPolicy: &trainer.PodGroupPolicy{
						PodGroupPolicySource: trainer.PodGroupPolicySource{
							Coscheduling: &trainer.CoschedulingPodGroupPolicySource{
								ScheduleTimeoutSeconds: ptr.To[int32](99),
							},
						},
					},
				},
				Scheduler: &runtime.Scheduler{
					PodLabels: map[string]string{
						schedulerpluginsv1alpha1.PodGroupLabel: "test-job",
					},
				},
			},
		},
		"an empty registry": {
			trainJob:        &trainer.TrainJob{ObjectMeta: metav1.ObjectMeta{Name: "test-job", Namespace: metav1.NamespaceDefault}},
			runtimeInfo:     &runtime.Info{},
			wantRuntimeInfo: &runtime.Info{},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			clientBuilder := testingutil.NewClientBuilder()

			fwk, err := New(ctx, clientBuilder.Build(), tc.registry, testingutil.AsIndex(clientBuilder))
			if err != nil {
				t.Fatal(err)
			}
			err = fwk.RunEnforcePodGroupPolicyPlugins(tc.runtimeInfo, tc.trainJob)
			if diff := cmp.Diff(tc.wantError, err, cmpopts.EquateErrors()); len(diff) != 0 {
				t.Errorf("Unexpected error (-want,+got): %s", diff)
			}
			if diff := cmp.Diff(tc.wantRuntimeInfo, tc.runtimeInfo); len(diff) != 0 {
				t.Errorf("Unexpected runtime.Info (-want,+got): %s", diff)
			}
		})
	}
}

func TestRunCustomValidationPlugins(t *testing.T) {
	cases := map[string]struct {
		registry     fwkplugins.Registry
		oldObj       *trainer.TrainJob
		newObj       *trainer.TrainJob
		wantWarnings admission.Warnings
		wantError    field.ErrorList
	}{
		// Need to implement more detail testing after we implement custom validator in any plugins.
		"there are not any custom validations": {
			registry: fwkplugins.NewRegistry(),
			oldObj:   testingutil.MakeTrainJobWrapper(metav1.NamespaceDefault, "test").Obj(),
			newObj:   testingutil.MakeTrainJobWrapper(metav1.NamespaceDefault, "test").Obj(),
		},
		"an empty registry": {
			oldObj: testingutil.MakeTrainJobWrapper(metav1.NamespaceDefault, "test").Obj(),
			newObj: testingutil.MakeTrainJobWrapper(metav1.NamespaceDefault, "test").Obj(),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			clientBuildr := testingutil.NewClientBuilder()

			fwk, err := New(ctx, clientBuildr.Build(), tc.registry, testingutil.AsIndex(clientBuildr))
			if err != nil {
				t.Fatal(err)
			}
			jobSetSpecApply, err := apply.FromTypedObjWithFields[jobsetv1alpha2ac.JobSetSpecApplyConfiguration](
				testingutil.MakeJobSetWrapper(metav1.NamespaceDefault, "test").Obj(),
				"spec",
			)
			if err != nil {
				t.Fatalf("Failed to convert typed JobSet to ApplyConfigurations: %v", err)
			}
			runtimeInfo := runtime.NewInfo(
				runtime.WithTemplateSpecObjApply(jobSetSpecApply),
			)
			warnings, errs := fwk.RunCustomValidationPlugins(ctx, runtimeInfo, tc.oldObj, tc.newObj)
			if diff := cmp.Diff(tc.wantWarnings, warnings, cmpopts.SortSlices(func(a, b string) bool { return a < b })); len(diff) != 0 {
				t.Errorf("Unexpected warninigs (-want,+got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantError, errs, cmpopts.IgnoreFields(field.Error{}, "Detail", "BadValue")); len(diff) != 0 {
				t.Errorf("Unexpected error (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestRunComponentBuilderPlugins(t *testing.T) {
	cases := map[string]struct {
		registry        fwkplugins.Registry
		runtimeInfo     *runtime.Info
		trainingRuntime *trainer.TrainingRuntime
		trainJob        *trainer.TrainJob
		wantRuntimeInfo *runtime.Info
		wantObjs        []apiruntime.Object
		wantError       error
	}{
		"succeeded to build PodGroup and JobSet with NumNodes from TrainJob": {
			registry: fwkplugins.NewRegistry(),
			trainingRuntime: testingutil.MakeTrainingRuntimeWrapper(metav1.NamespaceDefault, "test-runtime").
				Obj(),
			runtimeInfo: &runtime.Info{
				RuntimePolicy: runtime.RuntimePolicy{
					PodGroupPolicy: &trainer.PodGroupPolicy{
						PodGroupPolicySource: trainer.PodGroupPolicySource{
							Coscheduling: &trainer.CoschedulingPodGroupPolicySource{
								ScheduleTimeoutSeconds: ptr.To[int32](300),
							},
						},
					},
				},
				TemplateSpec: runtime.TemplateSpec{
					PodSets: []runtime.PodSet{
						{
							Name:     constants.DatasetInitializer,
							Ancestor: ptr.To(constants.DatasetInitializer),
							Count:    ptr.To[int32](1),
							SinglePodRequests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("4Gi"),
							},
							Containers: []runtime.Container{
								{
									VolumeMounts: []corev1ac.VolumeMountApplyConfiguration{
										*corev1ac.VolumeMount().
											WithName(jobsetplgconsts.VolumeNameInitializer).
											WithMountPath(constants.DatasetMountPath),
									},
								},
							},
							Volumes: []corev1ac.VolumeApplyConfiguration{
								*corev1ac.Volume().
									WithName(jobsetplgconsts.VolumeNameInitializer).
									WithPersistentVolumeClaim(corev1ac.PersistentVolumeClaimVolumeSource().
										WithClaimName(jobsetplgconsts.VolumeNameInitializer),
									),
							},
						},
						{
							Name:     constants.ModelInitializer,
							Ancestor: ptr.To(constants.ModelInitializer),
							Count:    ptr.To[int32](1),
							SinglePodRequests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("4Gi"),
							},
							Containers: []runtime.Container{
								{
									VolumeMounts: []corev1ac.VolumeMountApplyConfiguration{
										*corev1ac.VolumeMount().
											WithName(jobsetplgconsts.VolumeNameInitializer).
											WithMountPath(constants.ModelMountPath),
									},
								},
							},
							Volumes: []corev1ac.VolumeApplyConfiguration{
								*corev1ac.Volume().
									WithName(jobsetplgconsts.VolumeNameInitializer).
									WithPersistentVolumeClaim(corev1ac.PersistentVolumeClaimVolumeSource().
										WithClaimName(jobsetplgconsts.VolumeNameInitializer),
									),
							},
						},
						{
							Name:     constants.Node,
							Ancestor: ptr.To(constants.AncestorTrainer),
							Count:    ptr.To[int32](1),
							SinglePodRequests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("4Gi"),
							},
							Containers: []runtime.Container{{
								VolumeMounts: []corev1ac.VolumeMountApplyConfiguration{
									*corev1ac.VolumeMount().
										WithName(jobsetplgconsts.VolumeNameInitializer).
										WithMountPath(constants.DatasetMountPath),
									*corev1ac.VolumeMount().
										WithName(jobsetplgconsts.VolumeNameInitializer).
										WithMountPath(constants.ModelMountPath),
								},
							}},
						},
					},
					ObjApply: jobsetv1alpha2ac.JobSetSpec().
						WithReplicatedJobs(
							jobsetv1alpha2ac.ReplicatedJob().
								WithName(constants.DatasetInitializer).
								WithTemplate(batchv1ac.JobTemplateSpec().
									WithLabels(map[string]string{
										constants.LabelTrainJobAncestor: constants.DatasetInitializer,
									}).
									WithSpec(batchv1ac.JobSpec().
										WithTemplate(corev1ac.PodTemplateSpec().
											WithSpec(corev1ac.PodSpec().
												WithContainers(
													corev1ac.Container().
														WithName(constants.DatasetInitializer).
														WithVolumeMounts(
															corev1ac.VolumeMount().
																WithName(jobsetplgconsts.VolumeNameInitializer).
																WithMountPath(constants.DatasetMountPath),
														),
												).
												WithVolumes(
													corev1ac.Volume().
														WithName(jobsetplgconsts.VolumeNameInitializer).
														WithPersistentVolumeClaim(corev1ac.PersistentVolumeClaimVolumeSource().
															WithClaimName(jobsetplgconsts.VolumeNameInitializer)),
												),
											),
										),
									),
								),
							jobsetv1alpha2ac.ReplicatedJob().
								WithName(constants.ModelInitializer).
								WithTemplate(batchv1ac.JobTemplateSpec().
									WithLabels(map[string]string{
										constants.LabelTrainJobAncestor: constants.ModelInitializer,
									}).
									WithSpec(batchv1ac.JobSpec().
										WithTemplate(corev1ac.PodTemplateSpec().
											WithSpec(corev1ac.PodSpec().
												WithContainers(
													corev1ac.Container().
														WithName(constants.ModelInitializer).
														WithVolumeMounts(
															corev1ac.VolumeMount().
																WithName(jobsetplgconsts.VolumeNameInitializer).
																WithMountPath(constants.ModelMountPath),
														),
												).
												WithVolumes(
													corev1ac.Volume().
														WithName(jobsetplgconsts.VolumeNameInitializer).
														WithPersistentVolumeClaim(corev1ac.PersistentVolumeClaimVolumeSource().
															WithClaimName(jobsetplgconsts.VolumeNameInitializer)),
												),
											),
										),
									),
								),
							jobsetv1alpha2ac.ReplicatedJob().
								WithName(constants.Node).
								WithTemplate(batchv1ac.JobTemplateSpec().
									WithLabels(map[string]string{
										constants.LabelTrainJobAncestor: constants.AncestorTrainer,
									}).
									WithSpec(batchv1ac.JobSpec().
										WithTemplate(corev1ac.PodTemplateSpec().
											WithSpec(corev1ac.PodSpec().
												WithContainers(
													corev1ac.Container().
														WithName(constants.Node).
														WithVolumeMounts(
															corev1ac.VolumeMount().
																WithName(jobsetplgconsts.VolumeNameInitializer).
																WithMountPath(constants.DatasetMountPath),
															corev1ac.VolumeMount().
																WithName(jobsetplgconsts.VolumeNameInitializer).
																WithMountPath(constants.ModelMountPath),
														),
												).
												WithVolumes(
													corev1ac.Volume().
														WithName(jobsetplgconsts.VolumeNameInitializer).
														WithPersistentVolumeClaim(corev1ac.PersistentVolumeClaimVolumeSource().
															WithClaimName(jobsetplgconsts.VolumeNameInitializer)),
												),
											),
										),
									),
								),
						),
				},
				Scheduler: &runtime.Scheduler{PodLabels: make(map[string]string)},
			},
			trainJob: testingutil.MakeTrainJobWrapper(metav1.NamespaceDefault, "test-job").
				UID("uid").
				RuntimeRef(trainer.SchemeGroupVersion.WithKind(trainer.TrainingRuntimeKind), "test-runtime").
				Trainer(
					testingutil.MakeTrainJobTrainerWrapper().
						NumNodes(100).
						Container("test:trainjob", []string{"trainjob"}, []string{"trainjob"}, corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("4Gi"),
						}).
						Obj(),
				).
				Obj(),
			wantRuntimeInfo: &runtime.Info{
				RuntimePolicy: runtime.RuntimePolicy{
					PodGroupPolicy: &trainer.PodGroupPolicy{
						PodGroupPolicySource: trainer.PodGroupPolicySource{
							Coscheduling: &trainer.CoschedulingPodGroupPolicySource{
								ScheduleTimeoutSeconds: ptr.To[int32](300),
							},
						},
					},
				},
				TemplateSpec: runtime.TemplateSpec{
					ObjApply: jobsetv1alpha2ac.JobSetSpec().
						WithReplicatedJobs(
							jobsetv1alpha2ac.ReplicatedJob().
								WithReplicas(1).
								WithName(constants.DatasetInitializer).
								WithTemplate(batchv1ac.JobTemplateSpec().
									WithLabels(map[string]string{
										constants.LabelTrainJobAncestor: constants.DatasetInitializer,
									}).
									WithSpec(batchv1ac.JobSpec().
										WithTemplate(corev1ac.PodTemplateSpec().
											WithLabels(map[string]string{
												schedulerpluginsv1alpha1.PodGroupLabel: "test-job",
											}).
											WithSpec(corev1ac.PodSpec().
												WithContainers(
													corev1ac.Container().
														WithName(constants.DatasetInitializer).
														WithVolumeMounts(
															corev1ac.VolumeMount().
																WithName(jobsetplgconsts.VolumeNameInitializer).
																WithMountPath(constants.DatasetMountPath),
														),
												).
												WithVolumes(
													corev1ac.Volume().
														WithName(jobsetplgconsts.VolumeNameInitializer).
														WithPersistentVolumeClaim(corev1ac.PersistentVolumeClaimVolumeSource().
															WithClaimName(jobsetplgconsts.VolumeNameInitializer)),
												),
											),
										),
									),
								),
							jobsetv1alpha2ac.ReplicatedJob().
								WithReplicas(1).
								WithName(constants.ModelInitializer).
								WithTemplate(batchv1ac.JobTemplateSpec().
									WithLabels(map[string]string{
										constants.LabelTrainJobAncestor: constants.ModelInitializer,
									}).
									WithSpec(batchv1ac.JobSpec().
										WithTemplate(corev1ac.PodTemplateSpec().
											WithLabels(map[string]string{
												schedulerpluginsv1alpha1.PodGroupLabel: "test-job",
											}).
											WithSpec(corev1ac.PodSpec().
												WithContainers(
													corev1ac.Container().
														WithName(constants.ModelInitializer).
														WithVolumeMounts(
															corev1ac.VolumeMount().
																WithName(jobsetplgconsts.VolumeNameInitializer).
																WithMountPath(constants.ModelMountPath),
														),
												).
												WithVolumes(
													corev1ac.Volume().
														WithName(jobsetplgconsts.VolumeNameInitializer).
														WithPersistentVolumeClaim(corev1ac.PersistentVolumeClaimVolumeSource().
															WithClaimName(jobsetplgconsts.VolumeNameInitializer)),
												),
											),
										),
									),
								),
							jobsetv1alpha2ac.ReplicatedJob().
								WithName(constants.Node).
								WithReplicas(1).
								WithTemplate(batchv1ac.JobTemplateSpec().
									WithLabels(map[string]string{
										constants.LabelTrainJobAncestor: constants.AncestorTrainer,
									}).
									WithSpec(batchv1ac.JobSpec().
										WithParallelism(100).
										WithCompletions(100).
										WithTemplate(corev1ac.PodTemplateSpec().
											WithLabels(map[string]string{
												schedulerpluginsv1alpha1.PodGroupLabel: "test-job",
											}).
											WithSpec(corev1ac.PodSpec().
												WithContainers(
													corev1ac.Container().
														WithName(constants.Node).
														WithImage("test:trainjob").
														WithCommand("trainjob").
														WithArgs("trainjob").
														WithResources(corev1ac.ResourceRequirements().
															WithRequests(corev1.ResourceList{
																corev1.ResourceCPU:    resource.MustParse("1"),
																corev1.ResourceMemory: resource.MustParse("4Gi"),
															})).
														WithVolumeMounts(
															corev1ac.VolumeMount().
																WithName(jobsetplgconsts.VolumeNameInitializer).
																WithMountPath(constants.DatasetMountPath),
															corev1ac.VolumeMount().
																WithName(jobsetplgconsts.VolumeNameInitializer).
																WithMountPath(constants.ModelMountPath),
														),
												).
												WithVolumes(
													corev1ac.Volume().
														WithName(jobsetplgconsts.VolumeNameInitializer).
														WithPersistentVolumeClaim(corev1ac.PersistentVolumeClaimVolumeSource().
															WithClaimName(jobsetplgconsts.VolumeNameInitializer)),
												),
											),
										),
									),
								),
						),
					PodSets: []runtime.PodSet{
						{
							Name:     constants.DatasetInitializer,
							Ancestor: ptr.To(constants.DatasetInitializer),
							Count:    ptr.To[int32](1),
							SinglePodRequests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("4Gi"),
							},
							Containers: []runtime.Container{
								{
									VolumeMounts: []corev1ac.VolumeMountApplyConfiguration{
										*corev1ac.VolumeMount().
											WithName(jobsetplgconsts.VolumeNameInitializer).
											WithMountPath(constants.DatasetMountPath),
									},
								},
							},
							Volumes: []corev1ac.VolumeApplyConfiguration{
								*corev1ac.Volume().
									WithName(jobsetplgconsts.VolumeNameInitializer).
									WithPersistentVolumeClaim(corev1ac.PersistentVolumeClaimVolumeSource().
										WithClaimName(jobsetplgconsts.VolumeNameInitializer),
									),
							},
						},
						{
							Name:     constants.ModelInitializer,
							Ancestor: ptr.To(constants.ModelInitializer),
							Count:    ptr.To[int32](1),
							SinglePodRequests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("4Gi"),
							},
							Containers: []runtime.Container{
								{
									VolumeMounts: []corev1ac.VolumeMountApplyConfiguration{
										*corev1ac.VolumeMount().
											WithName(jobsetplgconsts.VolumeNameInitializer).
											WithMountPath(constants.ModelMountPath),
									},
								},
							},
							Volumes: []corev1ac.VolumeApplyConfiguration{
								*corev1ac.Volume().
									WithName(jobsetplgconsts.VolumeNameInitializer).
									WithPersistentVolumeClaim(corev1ac.PersistentVolumeClaimVolumeSource().
										WithClaimName(jobsetplgconsts.VolumeNameInitializer),
									),
							},
						},
						{
							Name:     constants.Node,
							Ancestor: ptr.To(constants.AncestorTrainer),
							Count:    ptr.To[int32](100),
							SinglePodRequests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("4Gi"),
							},
							Containers: []runtime.Container{{
								VolumeMounts: []corev1ac.VolumeMountApplyConfiguration{
									*corev1ac.VolumeMount().
										WithName(jobsetplgconsts.VolumeNameInitializer).
										WithMountPath(constants.DatasetMountPath),
									*corev1ac.VolumeMount().
										WithName(jobsetplgconsts.VolumeNameInitializer).
										WithMountPath(constants.ModelMountPath),
								},
							}},
						},
					},
				},
				Scheduler: &runtime.Scheduler{
					PodLabels: map[string]string{schedulerpluginsv1alpha1.PodGroupLabel: "test-job"},
				},
			},
			wantObjs: []apiruntime.Object{
				testingutil.MakeSchedulerPluginsPodGroup(metav1.NamespaceDefault, "test-job").
					SchedulingTimeout(300).
					MinMember(102). // 102 replicas = 100 Trainer nodes + 2 Initializer.
					MinResources(corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("102"), // 1 CPU and 4Gi per replica.
						corev1.ResourceMemory: resource.MustParse("408Gi"),
					}).
					ControllerReference(trainer.SchemeGroupVersion.WithKind("TrainJob"), "test-job", "uid").
					Obj(),
				testingutil.MakeJobSetWrapper(metav1.NamespaceDefault, "test-job").
					ControllerReference(trainer.SchemeGroupVersion.WithKind("TrainJob"), "test-job", "uid").
					PodLabel(schedulerpluginsv1alpha1.PodGroupLabel, "test-job").
					Replicas(1, constants.DatasetInitializer, constants.ModelInitializer, constants.Node).
					NumNodes(100).
					Container(constants.Node, constants.Node, "test:trainjob", []string{"trainjob"}, []string{"trainjob"}, corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					}).
					Obj(),
			},
		},
		"an empty registry": {},
	}
	cmpOpts := []cmp.Option{
		cmpopts.SortSlices(func(a, b apiruntime.Object) bool {
			return a.GetObjectKind().GroupVersionKind().String() < b.GetObjectKind().GroupVersionKind().String()
		}),
		cmpopts.EquateEmpty(),
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			clientBuilder := testingutil.NewClientBuilder()
			c := clientBuilder.Build()

			fwk, err := New(ctx, c, tc.registry, testingutil.AsIndex(clientBuilder))
			if err != nil {
				t.Fatal(err)
			}

			if err = fwk.RunEnforcePodGroupPolicyPlugins(tc.runtimeInfo, tc.trainJob); err != nil {
				t.Fatal(err)
			}
			if err = fwk.RunEnforceMLPolicyPlugins(tc.runtimeInfo, tc.trainJob); err != nil {
				t.Fatal(err)
			}
			objs, err := fwk.RunComponentBuilderPlugins(ctx, tc.runtimeInfo, tc.trainJob)

			if diff := cmp.Diff(tc.wantError, err, cmpopts.EquateErrors()); len(diff) != 0 {
				t.Errorf("Unexpected errors (-want,+got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.wantRuntimeInfo, tc.runtimeInfo); len(diff) != 0 {
				t.Errorf("Unexpected runtime.Info (-want,+got)\n%s", diff)
			}

			resultObjs, err := testingutil.ToObject(c.Scheme(), objs...)
			if err != nil {
				t.Errorf("Pipeline built unrecognizable objects: %v", err)
			}

			if diff := cmp.Diff(tc.wantObjs, resultObjs, cmpOpts...); len(diff) != 0 {
				t.Errorf("Unexpected objects (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestWatchExtensionPlugins(t *testing.T) {
	cases := map[string]struct {
		registry    fwkplugins.Registry
		wantPlugins []framework.WatchExtensionPlugin
	}{
		"coscheduling, jobset, and mpi are performed": {
			registry: fwkplugins.NewRegistry(),
			wantPlugins: []framework.WatchExtensionPlugin{
				&coscheduling.CoScheduling{},
				&jobset.JobSet{},
				&mpi.MPI{},
			},
		},
		"an empty registry": {
			wantPlugins: nil,
		},
	}
	cmpOpts := []cmp.Option{
		cmpopts.SortSlices(func(a, b framework.Plugin) bool { return a.Name() < b.Name() }),
		cmpopts.IgnoreUnexported(coscheduling.CoScheduling{}, jobset.JobSet{}, mpi.MPI{}),
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			clientBuilder := testingutil.NewClientBuilder()

			fwk, err := New(ctx, clientBuilder.Build(), tc.registry, testingutil.AsIndex(clientBuilder))
			if err != nil {
				t.Fatal(err)
			}
			plugins := fwk.WatchExtensionPlugins()
			if diff := cmp.Diff(tc.wantPlugins, plugins, cmpOpts...); len(diff) != 0 {
				t.Errorf("Unexpected plugins (-want,+got):\n%s", diff)
			}
		})
	}
}

type fakeTerminalConditionPlugin struct{}

var _ framework.TerminalConditionPlugin = (*fakeTerminalConditionPlugin)(nil)

func newFakeTerminalConditionPlugin(context.Context, client.Client, client.FieldIndexer) (framework.Plugin, error) {
	return &fakeTerminalConditionPlugin{}, nil
}

const fakeTerminalConditionPluginName = "fake"

func (f fakeTerminalConditionPlugin) Name() string { return fakeTerminalConditionPluginName }
func (f fakeTerminalConditionPlugin) TerminalCondition(context.Context, *trainer.TrainJob) (*metav1.Condition, error) {
	return nil, nil
}

func TestTerminalConditionPlugins(t *testing.T) {
	cases := map[string]struct {
		registry      fwkplugins.Registry
		trainJob      *trainer.TrainJob
		jobSet        *jobsetv1alpha2.JobSet
		wantCondition *metav1.Condition
		wantError     error
	}{
		"jobSet has not been finalized, yet": {
			registry: fwkplugins.NewRegistry(),
			trainJob: testingutil.MakeTrainJobWrapper(metav1.NamespaceDefault, "testing").
				Obj(),
			jobSet: testingutil.MakeJobSetWrapper(metav1.NamespaceDefault, "testing").
				Conditions(metav1.Condition{
					Type:    string(jobsetv1alpha2.JobSetSuspended),
					Reason:  jobsetconsts.JobSetSuspendedReason,
					Message: jobsetconsts.JobSetSuspendedMessage,
					Status:  metav1.ConditionFalse,
				}).
				Obj(),
		},
		"succeeded to obtain completed terminal condition": {
			registry: fwkplugins.NewRegistry(),
			trainJob: testingutil.MakeTrainJobWrapper(metav1.NamespaceDefault, "testing").
				Obj(),
			jobSet: testingutil.MakeJobSetWrapper(metav1.NamespaceDefault, "testing").
				Conditions(metav1.Condition{
					Type:    string(jobsetv1alpha2.JobSetCompleted),
					Reason:  jobsetconsts.AllJobsCompletedReason,
					Message: jobsetconsts.AllJobsCompletedMessage,
					Status:  metav1.ConditionTrue,
				}).
				Obj(),
			wantCondition: &metav1.Condition{
				Type:    trainer.TrainJobComplete,
				Reason:  jobsetconsts.AllJobsCompletedReason,
				Message: jobsetconsts.AllJobsCompletedMessage,
				Status:  metav1.ConditionTrue,
			},
		},
		"succeeded to obtain failed terminal condition": {
			registry: fwkplugins.NewRegistry(),
			trainJob: testingutil.MakeTrainJobWrapper(metav1.NamespaceDefault, "testing").
				Obj(),
			jobSet: testingutil.MakeJobSetWrapper(metav1.NamespaceDefault, "testing").
				Conditions(metav1.Condition{
					Type:    string(jobsetv1alpha2.JobSetFailed),
					Reason:  jobsetconsts.FailedJobsReason,
					Message: jobsetconsts.FailedJobsMessage,
					Status:  metav1.ConditionTrue,
				}).
				Obj(),
			wantCondition: &metav1.Condition{
				Type:    trainer.TrainJobFailed,
				Reason:  jobsetconsts.FailedJobsReason,
				Message: jobsetconsts.FailedJobsMessage,
				Status:  metav1.ConditionTrue,
			},
		},
		"failed to obtain any terminal condition due to multiple terminalCondition plugin": {
			registry: fwkplugins.Registry{
				jobset.Name:                     jobset.New,
				fakeTerminalConditionPluginName: newFakeTerminalConditionPlugin,
			},
			wantError: errorTooManyTerminalConditionPlugin,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			clientBuilder := testingutil.NewClientBuilder()
			if tc.jobSet != nil {
				clientBuilder = clientBuilder.WithObjects(tc.jobSet)
			}
			c := clientBuilder.Build()

			fwk, err := New(ctx, c, tc.registry, testingutil.AsIndex(clientBuilder))
			if err != nil {
				t.Fatal(err)
			}

			gotCond, gotErr := fwk.RunTerminalConditionPlugins(ctx, tc.trainJob)
			if diff := cmp.Diff(tc.wantError, gotErr, cmpopts.EquateErrors()); len(diff) != 0 {
				t.Errorf("Unexpected error (-want,+got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.wantCondition, gotCond); len(diff) != 0 {
				t.Errorf("Unexpected terminal condition (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestPodNetworkPlugins(t *testing.T) {
	cases := map[string]struct {
		registry        fwkplugins.Registry
		runtimeInfo     *runtime.Info
		trainJob        *trainer.TrainJob
		wantError       error
		wantRuntimeInfo *runtime.Info
	}{
		"Pod network is calculated by jobset plugin": {
			trainJob: testingutil.MakeTrainJobWrapper(metav1.NamespaceDefault, "test-job").
				Obj(),
			registry: fwkplugins.NewRegistry(),
			runtimeInfo: &runtime.Info{
				RuntimePolicy: runtime.RuntimePolicy{
					MLPolicySource: testingutil.MakeMLPolicySourceWrapper().Obj(),
				},
				TemplateSpec: runtime.TemplateSpec{
					PodSets: []runtime.PodSet{
						{
							Name:       constants.Node,
							Count:      ptr.To[int32](2),
							Containers: make([]runtime.Container, 1),
						},
					},
					ObjApply: jobsetv1alpha2ac.JobSetSpec().
						WithReplicatedJobs(
							jobsetv1alpha2ac.ReplicatedJob().
								WithName(constants.Node).
								WithTemplate(batchv1ac.JobTemplateSpec().
									WithSpec(batchv1ac.JobSpec().
										WithParallelism(1).
										WithCompletions(1).
										WithTemplate(corev1ac.PodTemplateSpec().
											WithSpec(corev1ac.PodSpec().
												WithContainers(
													corev1ac.Container().
														WithName(constants.Node),
												),
											),
										),
									),
								),
						),
				},
			},
			wantRuntimeInfo: &runtime.Info{
				RuntimePolicy: runtime.RuntimePolicy{
					MLPolicySource: testingutil.MakeMLPolicySourceWrapper().Obj(),
				},
				TemplateSpec: runtime.TemplateSpec{
					PodSets: []runtime.PodSet{
						{
							Name:       constants.Node,
							Containers: make([]runtime.Container, 1),
							Count:      ptr.To[int32](2),
							Endpoints: func(yield func(string) bool) {
								yield("test-job-node-0-0.test-job")
								yield("test-job-node-0-1.test-job")
							},
						},
					},
					ObjApply: jobsetv1alpha2ac.JobSetSpec().
						WithReplicatedJobs(
							jobsetv1alpha2ac.ReplicatedJob().
								WithName(constants.Node).
								WithTemplate(batchv1ac.JobTemplateSpec().
									WithSpec(batchv1ac.JobSpec().
										WithParallelism(1).
										WithCompletions(1).
										WithTemplate(corev1ac.PodTemplateSpec().
											WithSpec(corev1ac.PodSpec().
												WithContainers(
													corev1ac.Container().
														WithName(constants.Node),
												),
											),
										),
									),
								),
						),
				},
			},
		},
		"am empty registry": {
			trainJob:        testingutil.MakeTrainJobWrapper(metav1.NamespaceDefault, "test-job").Obj(),
			runtimeInfo:     &runtime.Info{},
			wantRuntimeInfo: &runtime.Info{},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, ctx := ktesting.NewTestContext(t)
			var cancel func()
			ctx, cancel = context.WithCancel(ctx)
			t.Cleanup(cancel)
			cliBuilder := testingutil.NewClientBuilder()
			fwk, err := New(ctx, cliBuilder.Build(), tc.registry, testingutil.AsIndex(cliBuilder))
			if err != nil {
				t.Fatal(err)
			}
			err = fwk.RunPodNetworkPlugins(tc.runtimeInfo, tc.trainJob)
			if diff := cmp.Diff(tc.wantError, err); len(diff) != 0 {
				t.Errorf("Unexpected error (-want,+got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantRuntimeInfo, tc.runtimeInfo, testingutil.PodSetEndpointsCmpOpts); len(diff) != 0 {
				t.Errorf("Unexpected runtimeInfo (-want,+got):\n%s", diff)
			}
		})
	}
}
