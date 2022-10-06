package main

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/vm-operator-api/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type (
	VMOptions struct {
		Name             string `json:"name,omitempty"`
		Namespace        string `json:"namespace,omitempty"`
		UserData         string `json:"userData,omitempty"`
		StorageClassName string
	}
)

var gvr schema.GroupVersionResource

func init() {
	gvr = schema.GroupVersionResource{
		Group:    "vmoperator.vmware.com",
		Version:  "v1alpha1",
		Resource: "virtualmachines",
	}
}

func createJumpBox(ctx context.Context, options *VMOptions) error {
	return nil
}

func createPVC(ctx context.Context, options *VMOptions) error {
	filesystem := corev1.PersistentVolumeFilesystem

	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      options.Name + "-pvc",
			Namespace: options.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			VolumeName:       "workspace",
			StorageClassName: &options.StorageClassName,
			VolumeMode:       &filesystem,
		},
	}

	_, err := c.CoreV1().PersistentVolumeClaims(options.Namespace).Create(ctx, &pvc, v1.CreateOptions{})
	if err != nil {
		errors.Wrap(err, "err creating pvc")
		return err
	}
	return nil
}

func createVM(ctx context.Context, options *VMOptions) error {
	vm := v1alpha1.VirtualMachine{
		TypeMeta: v1.TypeMeta{
			Kind:       "VirtualMachine",
			APIVersion: "vmoperator.vmware.com/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      options.Name,
			Namespace: "vms",
		},
		Spec: v1alpha1.VirtualMachineSpec{
			ImageName:  "ubuntu-20-1633387172196",
			ClassName:  "best-effort-large",
			PowerState: "poweredOff",
			VmMetadata: &v1alpha1.VirtualMachineMetadata{
				ConfigMapName: "jumpbox-os-config",
				Transport:     "OvfEnv",
			},
			StorageClass: "vc01cl01-t0compute",
			NetworkInterfaces: []v1alpha1.VirtualMachineNetworkInterface{{
				NetworkType: "nsx-t",
			}},
			Volumes: []v1alpha1.VirtualMachineVolume{{
				Name: "workspace",
				PersistentVolumeClaim: &v1alpha1.PersistentVolumeClaimVolumeSource{
					PersistentVolumeClaimVolumeSource: corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: "vm-pv",
						ReadOnly:  false,
					},
				},
			}},
		},
	}

	data, err := json.Marshal(vm) // Convert to a json string
	if err != nil {
		return errors.Wrap(err, "err json marshal")
	}

	var vmMap map[string]interface{}
	err = json.Unmarshal(data, &vmMap) // Convert to a map
	if err != nil {
		return errors.Wrap(err, "err unmarshal to map")
	}

	vmData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&vmMap)
	if err != nil {
		return errors.WithMessage(err, "err converting to unstructured")
	}

	dataUnstructured := &unstructured.Unstructured{Object: vmData}
	_, err = dynamicClient.Resource(gvr).Namespace("vms").Create(ctx, dataUnstructured, v1.CreateOptions{
		TypeMeta: v1.TypeMeta{
			Kind:       "VirtualMachine",
			APIVersion: "v1alpha1",
		},
	})
	if err != nil {
		return errors.Wrap(err, "error creating vm")
	}
	//fmt.Printf("vm created %v\n", res)
	return nil
}

func powerOnVM(ctx context.Context, vmName string) error {
	patch := []interface{}{
		map[string]interface{}{
			"op":    "replace",
			"path":  "/spec/powerState",
			"value": "poweredOn",
		},
	}
	payload, err := json.Marshal(patch)
	if err != nil {
		return errors.Wrap(err, "err marshaling")
	}
	_, err = dynamicClient.Resource(gvr).Namespace("vms").Patch(ctx, vmName, types.JSONPatchType, payload, v1.PatchOptions{})
	if err != nil {
		return errors.Wrap(err, "err patching")
	}
	//fmt.Printf("VM Powered ON - %s\n", res.Object["metadata"].(map[string]interface{})["name"])
	return nil
}
