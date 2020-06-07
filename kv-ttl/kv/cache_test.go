package kv

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"testing"
	"time"
)

func TestCacheBackup(t *testing.T) {
	// clean up
	rx, _ := regexp.Compile(`.*\.json`)
	files, _ := ioutil.ReadDir(".")
	for _, f := range files {
		if rx.MatchString(f.Name()) {
			_ = os.Remove(f.Name())
		}
	}

	cache := NewCache(Configuration{
		BackupInterval: 1 * time.Second,
	})
	values := []T{
		{V: "one"},
		{V: "two"},
		{V: "three"},
	}
	for i, v := range values {
		cache.Add(fmt.Sprintf("%d", i), v)
	}
	go func() {
		time.Sleep(5 * time.Second)
		cache.AddWithTtl("4", T{V: "four"}, 3*time.Second)
	}()

	time.Sleep(10 * time.Second)

	// find backup file
	files, _ = ioutil.ReadDir(".")
	var backupFile string
	for _, f := range files {
		if rx.MatchString(f.Name()) {
			backupFile = f.Name()
			return
		}
	}

	newCache := NewCache(Configuration{
		FileName: backupFile,
	})
	storedValues := newCache.GetAll()
	if !reflect.DeepEqual(values, storedValues) {
		t.Errorf("%v\n!=\n%v", values, storedValues)
	}
}
