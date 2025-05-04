package md4

import (
	"testing"
)

func TestMD4(t *testing.T) {
	tests := []struct {
		source string
		hash   string
	}{
		{"", "31d6cfe0d16ae931b73c59d7e0c089c0"},
		{"abc", "a448017aaf21d8525fc10ae87aa6729d"},
		{"message digest", "d9130a8164549fe818874806e1c7014b"},
		{"abcdefghijklmnopqrstuvwxyz", "d79e1c308aa5bbcdeea8ed63df412da9"},
		{"Podalirius", "13c5197a7245b1cb75829e38e72002ad"},
		{"TheManticoreProject", "9802c1ca09c41e32abd53fdd764f793b"},
		{"00000000000000000000000000000000", "04054272316bf2b21e8039902c980ab7"},
		{"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Cras risus ante, mollis sed ultricies eget, bibendum at leo. Ut gravida tristique convallis. Vivamus eu pulvinar leo, eu pharetra enim. Nullam in eros mattis, finibus nibh id, eleifend lacus. Sed ac scelerisque nulla. Aenean ut risus id massa maximus suscipit vitae in nisl. Morbi sagittis turpis eu ullamcorper dapibus. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Donec et faucibus risus. Quisque non lobortis purus, id dapibus elit. Donec non ex porttitor, vestibulum eros ut, vulputate ante. Cras rutrum mauris sit amet feugiat iaculis.", "f59ab2ecb4e3ee399b9e37778fe1fb3b"},
	}

	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			md4 := New()
			md4.Write([]byte(test.source))
			result := md4.HexSum()

			if result != test.hash {
				t.Errorf("MD4(\"%s\") = %v; want %v", test.source, result, test.hash)
			}
		})
	}
}
