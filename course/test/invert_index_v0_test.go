package main

import (
	"fmt"
	"testing"
	"zcygo/course"
)

func TestBuildInvertIndex(t *testing.T) {
	docs := []*course.Doc{&course.Doc{1, []string{"go", "数据结构"}}, &course.Doc{2, []string{"go", "数据库"}}}
	index := course.BuildInvertIndex(docs)
	for k, v := range index {
		fmt.Println(k, v)
	}
}
