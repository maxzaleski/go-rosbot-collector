package rosbotcollector

import (
	"reflect"
	"testing"
)

// Hard to test simply because it strongly relies on the person running the test
// to have access to a populated bot activity page.

func TestNewClient(t *testing.T) {
	type args struct {
		usernameOrEmail string
		password        string
	}
	tests := []struct {
		name    string
		args    args
		want    Client
		wantErr bool
	}{
		{
			name: "Invalid credentials",
			args: args{
				usernameOrEmail: "test",
				password:        "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.args.usernameOrEmail, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
