package crypto

import (
	"reflect"
	"testing"
)

func TestGenerateBlsKey(t *testing.T) {
	type args struct {
		message string
	}
	tests := []struct {
		name    string
		args    args
		want    BLSKey
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "",
			args: args{
				message: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateBlsKey(tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateBlsKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateBlsKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}
