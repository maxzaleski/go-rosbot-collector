package rosbotcollector

import (
	"os"
	"testing"
)

func Test_parseFormBuildID(t *testing.T) {
	file, err := os.Open("./samples/login.html")
	if err != nil {
		t.Errorf("could not open html file")
	}
	defer file.Close()

	test := struct {
		name    string
		want    string
		wantErr bool
	}{
		name:    "form_build_id is found",
		want:    "form-this-is-a-test",
		wantErr: false,
	}
	t.Run(test.name, func(t *testing.T) {
		got, err := parseFormBuildID(file)
		if (err != nil) != test.wantErr {
			t.Errorf("parseFormBuildID() error = %v, wantErr %v", err, test.wantErr)
			return
		}
		if got != test.want {
			t.Errorf("parseFormBuildID() = %v, want %v", got, test.want)
		}
	})
}

func Test_parseActivityEndpoint(t *testing.T) {
	file, err := os.Open("./samples/landing.html")
	if err != nil {
		t.Errorf("could not open html file")
	}
	defer file.Close()

	test := struct {
		name    string
		want    string
		wantErr bool
	}{
		name:    "activity path is found",
		want:    "/user/1234567/bot-activity",
		wantErr: false,
	}
	t.Run(test.name, func(t *testing.T) {
		got, err := parseActivityEndpoint(file)
		if (err != nil) != test.wantErr {
			t.Errorf("parseActivityEndpoint() error = %v, wantErr %v", err, test.wantErr)
			return
		}
		if got != test.want {
			t.Errorf("parseActivityEndpoint() = %v, want %v", got, test.want)
		}
	})
}
