package main

import "testing"

func TestLogin(t *testing.T) {
	tests := []map[string]string{
		{
			"user": "ivan1",
			"pwd":  "xyz",
		},
		{
			"user": "ivan",
			"pwd":  "xyz",
		},
		{
			"pwd": "xyz",
		},
		{
			"user": "ivan1",
			"pwd":  "xy",
		},
	}

	for _, tcase := range tests {
		t.Log(tcase)
		repl, err := login(tcase)
		t.Log(repl, err)
	}
}
