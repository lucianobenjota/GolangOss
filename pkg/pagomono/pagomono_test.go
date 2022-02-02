package pagomono

import (
	"testing"
)

func TestFormatoCuit(t *testing.T) {
	type args struct {
		cuit string
	}

	testCuit := args{
		cuit: "23332227459",
	}

	tests := []struct {
		name        string
		args        args
		wantResCuit string
	}{
		// TODO: Add test cases.
		{name: "Generar CUIT con formato dd-DNI-d", args: testCuit, wantResCuit: "23-33222745-9"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResCuit, _ := FormatoCuit(tt.args.cuit); gotResCuit != tt.wantResCuit {
				t.Errorf("FormatoCuit() = %v, want %v", gotResCuit, tt.wantResCuit)
			}
		})
	}
}

func TestPrettyPrint(t *testing.T) {
	type args struct {
		v interface{}
	}

	good := args{
		v: &args{},
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "No error",
			args:    good,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PrettyPrint(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("PrettyPrint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_checkErr(t *testing.T) {
	type args struct {
		err error
	}

	errorTest := args{
		err: nil,
	}

	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "Probar un error", 
			args: errorTest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkErr(tt.args.err)
		})
	}
}
