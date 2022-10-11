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
	"time"
)

var (
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
)

func init() {
}

func CreateJumpBox(ctx context.Context) error {

	err := createPVC(ctx)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("Skip Creating PVC. %s\n", err)
		} else {
			return err
		}
	}
	err = createConfigMap(ctx)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("Skip Creating Config. %s\n", err)
		} else {
			return err
		}
	}
	err = createVM(ctx)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("Skip Creating VMs. %s\n", err)
		} else {
			return err
		}
	}
	err = createSvc(ctx)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("Skip Creating Service. %s\n", err)
		} else {
			return err
		}
	}

	if options.WaitCreate {
		return waitCreate(ctx)
	}

	fmt.Printf("Creating VM. run `kubectl get vm %s -n %s` for progress\n", options.Name, options.Namespace)
	return nil
}

func createSvc(ctx context.Context) error {
	svc := v1alpha1.VirtualMachineService{
		TypeMeta: v1.TypeMeta{
			Kind:       "VirtualMachineService",
			APIVersion: "vmoperator.vmware.com/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      options.svcName,
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

func createPVC(ctx context.Context) error {
	filesystem := corev1.PersistentVolumeFilesystem

	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      options.pvcName,
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

func createConfigMap(ctx context.Context) error {
	cm := corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      options.configName,
			Namespace: options.Namespace,
		},
		Data: map[string]string{
			"user-data": options.UserData,
			"hostname":  options.Name,
		},
	}

	_, err := c.CoreV1().ConfigMaps(options.Namespace).Create(ctx, &cm, v1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "err creating cm")
	}
	return nil
}

func createVM(ctx context.Context) error {
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
			ImageName:  options.ImageName,
			ClassName:  options.ClassName,
			PowerState: "poweredOn",
			VmMetadata: &v1alpha1.VirtualMachineMetadata{
				ConfigMapName: options.configName,
				Transport:     "OvfEnv",
			},
			StorageClass: options.StorageClassName,
			NetworkInterfaces: []v1alpha1.VirtualMachineNetworkInterface{{
				NetworkType: options.NetworkType,
			}},
			Volumes: []v1alpha1.VirtualMachineVolume{{
				Name: "workspace",
				PersistentVolumeClaim: &v1alpha1.PersistentVolumeClaimVolumeSource{
					PersistentVolumeClaimVolumeSource: corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: options.pvcName,
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
	_, err = dynamicClient.Resource(gvrVM).Namespace(options.Namespace).Create(ctx, dataUnstructured, v1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "error creating vm")
	}
	//fmt.Printf("vm created %v\n", res)
	return nil
}

func waitCreate(ctx context.Context) error {
	fmt.Print("\nwaiting for VM to be ready ")
	for true {
		fmt.Print(".")
		time.Sleep(5 * time.Second)

		vm, err := dynamicClient.Resource(gvrVM).Namespace(options.Namespace).Get(ctx, options.Name, v1.GetOptions{})
		if err != nil {
			return err
		}
		if status, ok := vm.Object["status"]; ok && status != nil {
			if _, ok := status.(map[string]interface{})["vmIp"]; ok {
				svc, err := c.CoreV1().Services(options.Namespace).Get(ctx, options.svcName, v1.GetOptions{})
				if err != nil {
					return err
				}
				fmt.Printf("Jumpbox %s is ready\n", vm.Object["metadata"].(map[string]interface{})["name"])
				fmt.Printf("Load balancer IP: %s\n", svc.Status.LoadBalancer.Ingress[0].IP)
				fmt.Printf("\nAccess Jumpbox: `tanzu jumpbox ssh %s -i %s -n %s\n", vm.Object["metadata"].(map[string]interface{})["name"], options.SshKeyPath, options.Namespace)
				break
			}
		}
	}

	return nil
}

func Destroy(ctx context.Context) error {
	err := dynamicClient.Resource(gvrVM).Namespace(options.Namespace).Delete(ctx, options.Name, v1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "error deleting VM")
	}
	err = dynamicClient.Resource(gvrSvc).Namespace(options.Namespace).Delete(ctx, options.svcName, v1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "error deleting VMService")
	}
	fmt.Println("VM Service deleted")
	err = c.CoreV1().ConfigMaps(options.Namespace).Delete(ctx, options.configName, v1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "error deleting Config")
	}
	fmt.Println("VM Config deleted")
	err = c.CoreV1().PersistentVolumeClaims(options.Namespace).Delete(ctx, options.pvcName, v1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "error deleting PVC")
	}
	fmt.Println("VM Persistent Volume deleted")
	return nil

}

func PowerOn(ctx context.Context) error {
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
	res, err := dynamicClient.Resource(gvrVM).Namespace(options.Namespace).Patch(ctx, options.Name, types.JSONPatchType, payload, v1.PatchOptions{})
	if err != nil {
		return errors.Wrap(err, "err patching")
	}
	fmt.Printf("VM Powered ON - %s\n", res.Object["metadata"].(map[string]interface{})["name"])
	return nil
}

func PowerOff(ctx context.Context) error {
	patch := []interface{}{
		map[string]interface{}{
			"op":    "replace",
			"path":  "/spec/powerState",
			"value": "poweredOff",
		},
	}
	payload, err := json.Marshal(patch)
	if err != nil {
		return errors.Wrap(err, "err marshaling")
	}
	res, err := dynamicClient.Resource(gvrVM).Namespace(options.Namespace).Patch(ctx, options.Name, types.JSONPatchType, payload, v1.PatchOptions{})
	if err != nil {
		return errors.Wrap(err, "err patching")
	}
	fmt.Printf("VM Powered Off - %s\n", res.Object["metadata"].(map[string]interface{})["name"])
	return nil
}

func Ssh(ctx context.Context) error {
	svc, err := c.CoreV1().Services(options.Namespace).Get(ctx, options.svcName, v1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "error getting svc")
	}

	ip := svc.Status.LoadBalancer.Ingress[0].IP

	cmd := exec.Command("ssh", "-i", options.SshKeyPath, "-o", "StrictHostKeyChecking=no", options.User+"@"+ip)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "error ssh into vm")
	}
	return nil
}
