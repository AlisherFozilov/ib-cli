package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestManager(t *testing.T) {
	file, err := os.OpenFile("testData/commands_1.txt", os.O_RDONLY, 0666)
	if err != nil {
		t.Fatal(err)
	}

	gotFile, err := os.OpenFile("got.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		t.Fatal(err)
	}

	instream = file
	outstream = gotFile
	main()

	wantData, err := ioutil.ReadFile("testData/test_1.txt")
	//if err != nil {
	//	t.Fatal(err)
	//}
	gotData, err := ioutil.ReadFile("got.txt")
	if err != nil {
		t.Fatal(err)
	}

	if string(wantData) != string(gotData) {
		t.Error("EPIC FAIL!")
	}
	fmt.Println(len(string(wantData)))
	fmt.Println(len(string(gotData)))
}