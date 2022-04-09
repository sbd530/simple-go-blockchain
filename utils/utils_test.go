package utils

import (
	"encoding/hex"
	"errors"
	"reflect"
	"testing"
)

func TestGetHash(t *testing.T) {
	hash := "e005c1d727f7776a57a661d61a182816d8953c0432780beeae35e337830b1746"
	s := struct{ Test string }{Test: "test"}

	//1. x 와 hash 가 같아야함
	t.Run("hash is always same", func(t *testing.T) {
		x := GetHash(s)
		if x != hash {
			t.Errorf("Expected hash : %s, got %s", hash, x)
		}

	})

	//2. hash는 hex encoded여야함
	t.Run("hash is hex encoded", func(t *testing.T) {
		x := GetHash(s)
		_, err := hex.DecodeString(x)
		if err != nil {
			t.Error("hash should be hex encoded")
		}
	})

}

func TestToBytes(t *testing.T) {
	s := "test"
	b := ToBytes(s)
	tp := reflect.TypeOf(b).Kind()
	if tp != reflect.Slice {
		t.Errorf("ToBytes should return a slice of bytes. now is %s", tp)
	}
}

func TestSplitter(t *testing.T) {
	type test struct {
		input  string
		sep    string
		index  int
		output string
	}

	tests := []test{
		{input: "0:6:0", sep: ":", index: 1, output: "6"},
		{input: "0:6:0", sep: ":", index: 10, output: ""},
		{input: "0:6:0", sep: "/", index: 0, output: "0:6:0"},
	}

	for _, tc := range tests {
		got := Splitter(tc.input, tc.sep, tc.index)
		if got != tc.output {
			t.Errorf("Expected : %s. now got : %s", tc.output, got)
		}
	}
}

func TestHandleErr(t *testing.T) {
	oldLogFn := logFn
	defer func() {
		logFn = oldLogFn
	}()
	called := false
	logFn = func(v ...interface{}) {
		called = true
	}
	err := errors.New("test")
	HandleErr(err)

	if !called {
		t.Error("HandleError should call false")
	}
}
