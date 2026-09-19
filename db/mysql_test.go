package db

import "testing"

func TestMySQLConnection(t *testing.T) {
	if err := InitMySQL(); err != nil {
		t.Fatalf("failed %v\n", err)
	}
}
