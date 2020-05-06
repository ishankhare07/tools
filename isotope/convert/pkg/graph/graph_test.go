package graph

import (
	"istio.io/tools/isotope/convert/pkg/graph/svc"
	"reflect"
	"testing"
)

func TestServiceGraph_FindServiceByName(t *testing.T) {
	services := []svc.Service{
		{
			Name: "s0",
		},
		{
			Name: "s1",
		},
	}
	type fields struct {
		Global   ServiceDefaults
		Services []svc.Service
		Defaults Defaults
	}
	type args struct {
		serviceName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    svc.Service
		wantErr bool
	}{
		{"when service is not found", fields{Global: ServiceDefaults{}, Services: services, Defaults: Defaults{}}, args{serviceName: "s3"}, svc.Service{}, true},
		{"when service is found", fields{Global: ServiceDefaults{}, Services: services, Defaults: Defaults{}}, args{serviceName: "s1"}, svc.Service{Name: "s1"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceGraph := ServiceGraph{
				Global:   tt.fields.Global,
				Services: tt.fields.Services,
				Defaults: tt.fields.Defaults,
			}
			got, err := serviceGraph.FindServiceByName(tt.args.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindServiceByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindServiceByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}