package main

import "testing"

func Test_buildUserdata(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := buildUserdata(); (err != nil) != tt.wantErr {
				t.Errorf("buildUserdata() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setup(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "setup1",
			args: args{
				[]string{"test"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(tt.args.args)
			if options.Name != tt.args.args[0] {
				t.Errorf("setup() expect options.Name %s, got %s", tt.args.args[0], options.Name)
			}
		})
	}
}
