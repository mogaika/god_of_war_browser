package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/mogaika/god_of_war_browser/vfs"
)

func recurs(curdir vfs.Directory, tabsCount int) {
	tabs := "|=" + strings.Repeat(" -", tabsCount)
	fileList, err := curdir.List()
	if err != nil {
		fmt.Printf("%s dir error: %v\n", tabs, err)
	} else {
		for _, name := range fileList {
			if e, err := curdir.GetElement(name); err != nil {
				fmt.Printf("%s error: %v\n", tabs, err)
			} else {
				switch e.(type) {
				case vfs.Directory:
					fmt.Printf("%s DIR: %s\n", tabs, e.Name())
					recurs(e.(*vfs.DirectoryDriver), tabsCount+1)
				case vfs.File:
					fmt.Printf("%s FILE: %s\n", tabs, e.Name())
				}
			}
		}
	}
}

var teststring = []byte(" <++ = This is test string = ++> \r\n TESTING ! ! ! \r\n ....END....")

func main() {
	d := vfs.NewDirectoryDriver(path.Dir(os.Args[0]))
	recurs(d, 0)

	df := vfs.NewDirectoryDriverFile("_test-dir-driver_.bin")
	d.Add(df)

	e, err := d.GetElement(df.Name())
	if err != nil {
		log.Panicf("Cannot open just created file: %v", err)
	}

	f, ok := e.(vfs.File)
	if !ok {
		log.Panicf("Cannot convert element to file: %+#v", e)
	}

	var buf bytes.Buffer
	buf.Write(teststring)

	if err := f.Copy(&buf); err != nil {
		log.Panic(err)
	}
	log.Printf("File created")
	if err := f.Open(false); err != nil {
		log.Panic(err)
	}

	if r, err := f.Reader(); err != nil {
		log.Panic(err)
	} else {
		if newdata, err := ioutil.ReadAll(r); err != nil {
			log.Panic(err)
		} else {
			if bytes.Compare(newdata, teststring) != 0 {
				log.Panicf("%v != %v", newdata, teststring)
			} else {
				if err := f.Close(); err != nil {
					log.Panic(err)
				}

				if err := d.Remove(f.Name()); err != nil {
					log.Panic(err)
				}

				if err := f.Open(true); err == nil {
					log.Panic("Can open removed file")
				} else {
					log.Printf("File removed: %v", err)
					f.Close()
				}
			}
		}
	}
	log.Printf("SUCCESS")
}
