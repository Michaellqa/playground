package main

import (
	"fmt"
	"testing"
)

func TestCollectResults(t *testing.T) {
	count := 100
	ts := load(count)
	res := collectResults(ts)
	if len(res) != count {
		t.Errorf("%d", len(res))
		for _, v := range res {
			fmt.Println(v)
		}
	}
}
