package main

import (
	"net/http"
	"reflect"
	"testing"
)

func Test_getHandler(t *testing.T) {
	type args struct {
		g Getter
	}
	tests := []struct {
		name string
		args args
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHandler(tt.args.g); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
