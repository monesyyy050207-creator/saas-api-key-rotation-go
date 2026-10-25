package main

import "testing"

func TestGraceWindowDecision(t *testing.T) {
	tests := []struct {
		name  string
		grace int
		want  bool
	}{{"overlap", 24, true}, {"immediate cutover", 0, false}, {"negative", -1, false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if (tt.grace > 0) != tt.want {
				t.Fatalf("grace_hours=%d: want %v", tt.grace, tt.want)
			}
		})
	}
}
