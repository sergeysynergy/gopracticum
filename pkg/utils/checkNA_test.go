package utils

import "testing"

func TestCheckNA(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Ok",
			args: args{
				str: "nice",
			},
			want: "nice",
		},
		{
			name: "NA",
			args: args{
				str: "",
			},
			want: "N/A",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckNA(tt.args.str); got != tt.want {
				t.Errorf("CheckNA() = %v, want %v", got, tt.want)
			}
		})
	}
}
