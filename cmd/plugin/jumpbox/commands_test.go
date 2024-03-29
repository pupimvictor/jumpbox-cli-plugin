package main

import (
	"context"
	"github.com/vmware-tanzu/vm-operator-api/api/v1alpha1"
	"github.com/vmware-tanzu/vm-operator-api/api/v1alpha1/install"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	simpleFake "k8s.io/client-go/kubernetes/fake"
	"strings"
	"testing"
)

func Test_createJumpBox(t *testing.T) {
	ctx := context.Background()

	type args struct {
		options *VMOptions
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "test",
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateJumpBox(ctx); (err != nil) != tt.wantErr {
				t.Errorf("CreateJumpBox() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createPVC(t *testing.T) {
	type args struct {
		ctx     context.Context
		options *VMOptions
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "pvc-1",
			args: args{
				ctx: context.Background(),
				options: &VMOptions{
					Name:             "test-1",
					Namespace:        "test",
					UserData:         "",
					StorageClassName: "test",
				},
			},
			wantErr: false,
		},
	}
	c = simpleFake.NewSimpleClientset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createPVC(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("createPVC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createVM(t *testing.T) {
	ctx := context.Background()

	type args struct {
		ctx     context.Context
		options VMOptions
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "vm-1",
			args: args{
				ctx: ctx,
				options: VMOptions{
					Name:             "jumpbox-1",
					Namespace:        "test",
					UserData:         "test",
					StorageClassName: "test",
				},
			},
		}, {
			name: "vm-already-exists",
			args: args{
				ctx: ctx,
				options: VMOptions{
					Name:             "jumpbox-1",
					Namespace:        "test",
					UserData:         "test",
					StorageClassName: "test",
				},
			},
			wantErr: true,
		}, {
			name: "vm-missing-sc",
			args: args{
				ctx: ctx,
				options: VMOptions{
					Name:      "jumpbox-3",
					Namespace: "test",
					UserData:  "test",
				},
			},
			wantErr: true,
		},
	}
	scheme := runtime.NewScheme()
	install.Install(scheme)
	dynamicClient = fake.NewSimpleDynamicClient(scheme, &v1alpha1.VirtualMachine{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options = &tt.args.options
			err := createVM(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("createVM() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				t.Logf("Ok - err %s", err)
			}
		})
	}
}

func TestPowerOn(t *testing.T) {
	ctx := context.Background()
	type args struct {
		ctx    context.Context
		vmName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "poweron-1",
			args: args{
				ctx:    ctx,
				vmName: "jumpbox-1",
			},
		},
		{
			name: "poweron-not-found",
			args: args{
				ctx:    ctx,
				vmName: "jumpbox-2",
			},
			wantErr: true,
		},
	}
	scheme := runtime.NewScheme()
	install.Install(scheme)
	dynamicClient = fake.NewSimpleDynamicClient(scheme, &v1alpha1.VirtualMachine{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup([]string{tt.args.vmName})
			if !strings.Contains(tt.name, "not-found") {
				err := createVM(ctx)
				if err != nil {
					t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if err := PowerOn(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("PowerOn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateJumpBox(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "test",
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateJumpBox(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("CreateJumpBox() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDestroy(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "test",
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Destroy(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Destroy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPowerOff(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "test",
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PowerOff(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("PowerOff() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSsh(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "teste",
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SSH(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("SSH() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createConfigMap(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "teste",
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createConfigMap(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("createConfigMap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createSvc(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "",
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createSvc(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("createSvc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_waitCreate(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "",
			args:    args{},
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := waitCreate(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("waitCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
