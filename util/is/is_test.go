package is

import (
	"errors"
	"testing"
)

func TestIn(t *testing.T) {
	type args struct {
		arr  interface{}
		item interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test_in_1", args: args{arr: []int{1, 2, 3}, item: 1}, want: true},
		{name: "test_in_2", args: args{arr: []int{1, 2, 3}, item: 4}, want: false},
		{name: "test_in_3", args: args{arr: []string{"a", "b", "c"}, item: 2}, want: false},
		{name: "test_in_4", args: args{arr: []float32{1.1, 2.3, 1.3}, item: 2}, want: false},
		{name: "test_in_5", args: args{arr: []int64{1, 2, 3}, item: int64(1)}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := In(tt.args.arr, tt.args.item); got != tt.want {
				t.Errorf("In() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotIn(t *testing.T) {
	type args struct {
		arr  interface{}
		item interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test_not_in_1", args: args{arr: []int{1, 2, 3}, item: 1}, want: false},
		{name: "test_not_in_2", args: args{arr: []int{1, 2, 3}, item: 4}, want: true},
		{name: "test_not_in_3", args: args{arr: []string{"a", "b", "c"}, item: 2}, want: true},
		{name: "test_not_in_4", args: args{arr: []float32{1.1, 2.3, 1.3}, item: 2}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NotIn(tt.args.arr, tt.args.item); got != tt.want {
				t.Errorf("NotIn() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestZero(t *testing.T) {
	type ZeroStruct struct {
	}
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test_iszero_1", args: args{v: 0}, want: true},
		{name: "test_iszero_2", args: args{v: 1}, want: false},
		{name: "test_iszero_3", args: args{v: true}, want: false},
		{name: "test_iszero_4", args: args{v: struct {
			Name string
		}{Name: "xx"}}, want: false},
		{name: "test_iszero_5", args: args{v: 0.0}, want: true},
		{name: "test_iszero_6", args: args{v: &ZeroStruct{}}, want: false},
		{name: "test_iszero_7", args: args{v: (*ZeroStruct)(nil)}, want: true},
		{name: "test_iszero_8", args: args{v: ""}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Zero(tt.args.v); got != tt.want {
				t.Errorf("Zero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmail(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test_email_1", args: args{s: "12@qq.com"}, want: true},
		{name: "test_email_2", args: args{s: "ww@example"}, want: false},
		{name: "test_email_3", args: args{s: "@exmaple.org"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Email(tt.args.s); got != tt.want {
				t.Errorf("Email() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNumber(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test_number_1", args: args{s: "123"}, want: true},
		{name: "test_number_2", args: args{s: "1dfsf"}, want: false},
		{name: "test_number_3", args: args{s: "asdf"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Number(tt.args.s); got != tt.want {
				t.Errorf("Number() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPv4(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test_ipv4_1", args: args{s: "123.1.23.1"}, want: true},
		{name: "test_ipv4_2", args: args{s: "11131"}, want: false},
		{name: "test_ipv4_3", args: args{s: "255.255.255"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IPv4(tt.args.s); got != tt.want {
				t.Errorf("IPv4() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPv6(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test_ipv6_1", args: args{s: "2001:DB8:2de::e13"}, want: true},
		{name: "test_ipv6_2", args: args{s: "2001:DB8:2de:000:000:000:000:e13"}, want: true},
		{name: "test_ipv6_3", args: args{s: "255.255.255"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IPv6(tt.args.s); got != tt.want {
				t.Errorf("IPv6() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIP(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test_ip_1", args: args{s: "2001:DB8:2de::e13"}, want: true},
		{name: "test_ip_2", args: args{s: "2001:DB8:2de:000:000:000:000:e13"}, want: true},
		{name: "test_ip_3", args: args{s: "255.255.255.255"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IP(tt.args.s); got != tt.want {
				t.Errorf("IP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDomain(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test_domain_1", args: args{s: "baidu.cn"}, want: true},
		{name: "test_domain_2", args: args{s: "a.org"}, want: true},
		{name: "test_domain_3", args: args{s: "bbb"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Domain(tt.args.s); got != tt.want {
				t.Errorf("Domain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURL(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test_url_1", args: args{s: "http://baidu.cn"}, want: true},
		{name: "test_url_2", args: args{s: "http://a.org?name=1"}, want: true},
		{name: "test_url_3", args: args{s: "bbb"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := URL(tt.args.s); got != tt.want {
				t.Errorf("URL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUuid(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test_uuid_1", args: args{s: "0ff7af2b-33b3-40ec-8e2a-8dac5b340e19"}, want: true},
		{name: "test_uuid_2", args: args{s: "0ff7af2b-33b3-40ec-8e2a8dac5b3402e19"}, want: false},
		{name: "test_uuid_3", args: args{s: "bbb"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Uuid(tt.args.s); got != tt.want {
				t.Errorf("Uuid() = %v, want %v", got, tt.want)
			}
		})
	}
}

type S struct {
	Name  string
	Email string
	IP    string
	Age   int32
}

func (m *S) Validate() error {
	errs := make([]error, 0)
	if Zero(m.Name) {
		errs = append(errs, errors.New("field 'name' is required"))
	}

	return MargeErr(errs...)
}

func TestValidate(t *testing.T) {
	s := &S{}

	err := s.Validate()
	if err == nil {
		t.Fatal("S.Validate() = nil, want error")
	}

	t.Log(err)
}

func TestCrontab(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test_cron_1", args: args{s: "@every 3m"}, want: true},
		{name: "test_cron_2", args: args{s: "*/4 * * * * ?"}, want: true},
		{name: "test_cron_3", args: args{s: "5"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Crontab(tt.args.s); got != tt.want {
				t.Errorf("Crontab() = %v, want %v", got, tt.want)
			}
		})
	}
}
