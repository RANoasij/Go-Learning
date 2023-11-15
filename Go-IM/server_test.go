package main

import "testing"

// Write a test for the NewServer function
func TestNewServer(t *testing.T) {
	server := NewServer("127.0.0.1, 8888")
	if server == nil {
		t.Error("NewServer() failed")
	}
}
