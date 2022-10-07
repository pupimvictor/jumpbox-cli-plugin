package main

import (
	"context"
	"encoding/json"
	"fmt"
	errors "github.com/pkg/errors"
	"github.com/vmware-tanzu/vm-operator-api/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"os/exec"
)

type (
	VMOptions struct {
		Name             string `json:"name,omitempty"`
		Namespace        string `json:"namespace,omitempty"`
		UserData         string `json:"userData,omitempty"`
		PVCName          string `json:"pvcName"`
		ConfigName       string `json:"configName"`
		StorageClassName string `json:"storageClassName"`
	}
	sshOptions struct {
		vmName     string
		namespace  string
		sshKeyPath string
	}
)

var gvrVM schema.GroupVersionResource
var gvrSvc schema.GroupVersionResource

func init() {
	gvrVM = schema.GroupVersionResource{
		Group:    "vmoperator.vmware.com",
		Version:  "v1alpha1",
		Resource: "virtualmachines",
	}
	gvrSvc = schema.GroupVersionResource{
		Group:    "vmoperator.vmware.com",
		Version:  "v1alpha1",
		Resource: "virtualmachineservices",
	}

}

func createJumpBox(ctx context.Context, options *VMOptions) error {

	err := createPVC(ctx, options)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("Skip Creating PVC. %s\n", err)
		} else {
			return err
		}
	}
	err = createConfigMap(ctx, options)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("Skip Creating Config. %s\n", err)
		} else {
			return err
		}
	}
	err = createVM(ctx, options)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("Skip Creating VMs. %s\n", err)
		} else {
			return err
		}
	}
	err = createSvc(ctx, options)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("Skip Creating Service. %s\n", err)
		} else {
			return err
		}
	}

	return nil
}

func createSvc(ctx context.Context, options *VMOptions) error {
	svc := v1alpha1.VirtualMachineService{
		TypeMeta: v1.TypeMeta{
			Kind:       "VirtualMachineService",
			APIVersion: "vmoperator.vmware.com/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      options.Name + "-svc",
			Namespace: options.Namespace,
		},
		Spec: v1alpha1.VirtualMachineServiceSpec{
			Type: "LoadBalancer",
			Ports: []v1alpha1.VirtualMachineServicePort{{
				Name:       "ssh",
				Protocol:   "TCP",
				Port:       22,
				TargetPort: 22,
			}},
			Selector: map[string]string{
				"jumpbox": options.Name,
			},
		},
	}

	data, err := json.Marshal(svc) // Convert to a json string
	if err != nil {
		return errors.Wrap(err, "err json marshal")
	}

	var svcMap map[string]interface{}
	err = json.Unmarshal(data, &svcMap) // Convert to a map
	if err != nil {
		return errors.Wrap(err, "err unmarshal to map")
	}

	svcData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&svcMap)
	if err != nil {
		return errors.WithMessage(err, "err converting to unstructured")
	}

	dataUnstructured := &unstructured.Unstructured{Object: svcData}
	_, err = dynamicClient.Resource(gvrSvc).Namespace(options.Namespace).Create(ctx, dataUnstructured, v1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "err creating service")
	}
	return nil
}

func createPVC(ctx context.Context, options *VMOptions) error {
	filesystem := corev1.PersistentVolumeFilesystem

	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      options.PVCName,
			Namespace: options.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: resource.MustParse("12Gi"),
				},
			},
			StorageClassName: &options.StorageClassName,
			VolumeMode:       &filesystem,
		},
	}

	_, err := c.CoreV1().PersistentVolumeClaims(options.Namespace).Create(ctx, &pvc, v1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "err creating pvc")
	}
	return nil
}

func createConfigMap(ctx context.Context, options *VMOptions) error {
	cm := corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      options.ConfigName,
			Namespace: options.Namespace,
		},
		Data: map[string]string{
			"user-data": "I2Nsb3VkLWNvbmZpZwojIyBSZXF1aXJlZCBzeW50YXggYXQgdGhlIHN0YXJ0IG9mIHVzZXItZGF0YSBmaWxlCnVzZXJzOgojIyBDcmVhdGUgdGhlIGRlZmF1bHQgdXNlciBmb3IgdGhlIE9TCiAgLSBkZWZhdWx0CgogIC0gbmFtZTogdnAKICAgIGdlY29zOiB2aWN0b3IgcHVwaW0KICAgIHN1ZG86IEFMTD0oQUxMKSBOT1BBU1NXRDpBTEwKICAgIGdyb3VwczogdXNlcnMsIGFkbWlucwogICAgbG9ja19wYXNzd2Q6IGZhbHNlCiAgICBzaGVsbDogL2Jpbi9iYXNoCiAgICBwYXNzd2Q6IDczYjQwOTJjZjI4N2RhYmNhNTI5ZTdlZGExNmE0NGU4Yzk4MWVjNDM1NTVkMjUxOTE2YmNmZmJlZmE1OThlZmUKICAgIHNzaF9hdXRob3JpemVkX2tleXM6CiAgICAgIC0gc3NoLXJzYSBBQUFBQjNOemFDMXljMkVBQUFBREFRQUJBQUFDQVFDcHA4UVRMNDF5bVA4dmRmVFdPblE3eXhaek5FY1U5TWlWaEg2S3B6WWhzdkRsRjdSKzVrMTRycVNqajBlRlBrM2VWWGV5L0dTS2FNUjBCL29SMlhjdjk2Y1RuczNzMnZGa29rWUppeldiZEphU1gwVXEvcUVaUjFmVFdrcm9vTEN0c1JoS1hYQUtkaitKOTk5a3pNQ28xaE4wODF4SHlQdXB5YXErWjJVdDZKc0FWdElUSHE4aiswZzNTUndQQzlRVDJtVUJ6L3Q4TTh1THpRYWh6NlZ2UUs4ZGJEWkpCaU42TXdRME9PdVZrNFRxZnFldkV6SDFta2lZRVlLTGdrVU8vd0FGQmZjZUZ3K0NwSWUwUnZBaHdib25oVjg1TE5yRWczMU1FQ055Nm0xM1MxTVY2bUVFcHhqUHNEVjJEQVFqTi93dDV4M1VYenpQNGlLS2V3ZXhueCtWMklTQWxYTE96b3BPSTJLdTUwWGtiSFFGRUFNM3dHNlFKcWRCelFWUTA0dGRJdmpTdTNWdHVNcjhmYVVvT1ZKV0V1aXAxc2Mvc1ptNTlPd0JBMWxEbitvYjlIK2R0VUw4QlFkVDlTNFd2ZEhSbkpHclp6TXBBTEZKRGtJNUhCTnNGcGFpWmNId1RaWEk2RzR2aVpKVWFHVWNhRWl0eE9rV0QxZVV2TzZYRjFSdktNTk04a2Y3bmRQZzhaeXRMQi9tdllKcGErU0VyT1cwUXcxcDJNWlBIUXhtZnE5ODdiNVhKMk1WNGtjSXRTQmxXRHo4ZkJObjRaMUpXZ0V5d0hUNjJmOUt1RHR4cnkvYS8yNS82ZjM0VHVXQmp4RFBVUDhscThaRStrSU4ySjFRaTA1QVNBc2R3WWZ3YW1vTm9ZNHlqaGY5REVoTmVmY20zaEdsaHc9PSBwdXBpbXZpY3RvckBnbWFpbC5jb20KICAgICAgCiAgICAKIyMgRW5hYmxlIERIQ1Agb24gdGhlIGRlZmF1bHQgbmV0d29yayBpbnRlcmZhY2UgcHJvdmlzaW9uZWQgaW4gdGhlIFZNCm5ldHdvcms6CiAgdmVyc2lvbjogMgogIGV0aGVybmV0czoKICAgICAgZW5zMTkyOgogICAgICAgICAgZGhjcDQ6IHRydWUKCiMjIFNldHVwIEZpbGVzeXN0ZW0gYW5kIE1vdW50IFBWIGRpc2sKZnNfc2V0dXA6CiAgLSBsYWJlbDogd29ya3NwYWNlCiAgICBmaWxlc3lzdGVtOiAnZXh0NCcKICAgIGRldmljZTogJy9kZXYvc2RiJwogICAgcGFydGl0aW9uOiAnYXV0bycKCm1vdW50czoKIC0gWyBzZGIsIC93b3Jrc3BhY2UgXQoKYXB0X3VwZ3JhZGU6IHRydWUKcGFja2FnZXM6CiAgICAtIHRyYWNlcm91dGUKICAgIC0gdW56aXAKICAgIC0gdHJlZQogICAgLSBqcQoKcnVuY21kOgogIC0gY2htb2QgNzc0IC93b3Jrc3BhY2UKICAtIGNob3duIC1SIHJvb3Q6YWRtaW5zIC93b3Jrc3BhY2UKICA=",
			"hostname":  options.Name,
		},
	}

	_, err := c.CoreV1().ConfigMaps(options.Namespace).Create(ctx, &cm, v1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "err creating cm")
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
			Namespace: options.Namespace,
			Labels: map[string]string{
				"jumpbox": options.Name,
			},
		},
		Spec: v1alpha1.VirtualMachineSpec{
			ImageName:  "ubuntu-20-1633387172196",
			ClassName:  "best-effort-large",
			PowerState: "poweredOff",
			VmMetadata: &v1alpha1.VirtualMachineMetadata{
				ConfigMapName: options.ConfigName,
				Transport:     "OvfEnv",
			},
			StorageClass: options.StorageClassName,
			NetworkInterfaces: []v1alpha1.VirtualMachineNetworkInterface{{
				NetworkType: "nsx-t",
			}},
			Volumes: []v1alpha1.VirtualMachineVolume{{
				Name: "workspace",
				PersistentVolumeClaim: &v1alpha1.PersistentVolumeClaimVolumeSource{
					PersistentVolumeClaimVolumeSource: corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: options.PVCName,
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
	_, err = dynamicClient.Resource(gvrVM).Namespace(options.Namespace).Create(ctx, dataUnstructured, v1.CreateOptions{
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

func powerOn(ctx context.Context, options *VMOptions) error {
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
	_, err = dynamicClient.Resource(gvrVM).Namespace(options.Namespace).Patch(ctx, options.Name, types.JSONPatchType, payload, v1.PatchOptions{})
	if err != nil {
		return errors.Wrap(err, "err patching")
	}
	//fmt.Printf("VM Powered ON - %s\n", res.Object["metadata"].(map[string]interface{})["name"])
	return nil
}

func sshJumpbox(ctx context.Context, options *sshOptions) error {
	svc, err := c.CoreV1().Services(options.namespace).Get(ctx, options.vmName+"-svc", v1.GetOptions{})

	if err != nil {
		return errors.Wrap(err, "error getting svc")
	}

	ip := svc.Status.LoadBalancer.Ingress[0].IP

	cmd := exec.Command("ssh", "-i", options.sshKeyPath, "-o", "StrictHostKeyChecking=no", "vp@"+ip)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "error ssh into vm")
	}
	return nil
}
