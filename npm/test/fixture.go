//go:build ignore

// This executable doubles as the archive generator and the native shim fixture.
package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	if len(os.Args) == 4 && os.Args[1] == "archive" {
		must(archive(os.Args[2], os.Args[3]))
		return
	}
	fmt.Println(strings.Join(os.Args[1:], "|"))
	os.Exit(23)
}

func archive(filename, variant string) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	binary, err := os.ReadFile(self)
	if err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if strings.HasSuffix(filename, ".zip") {
		zw := zip.NewWriter(file)
		add := func(name string, data []byte, mode os.FileMode) error {
			header := &zip.FileHeader{Name: name, Method: zip.Deflate}
			header.SetMode(mode)
			entry, err := zw.CreateHeader(header)
			if err == nil {
				_, err = entry.Write(data)
			}
			return err
		}
		must(add("README.md", []byte("fixture"), 0o644))
		switch variant {
		case "success", "duplicate":
			must(add("gumlet.exe", binary, 0o755))
		case "symlink":
			must(add("gumlet.exe", []byte("README.md"), os.ModeSymlink|0o777))
		}
		if variant == "duplicate" {
			must(add("gumlet.exe", []byte("duplicate"), 0o755))
		}
		return zw.Close()
	}

	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	add := func(header *tar.Header, data []byte) error {
		header.Size = int64(len(data))
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		_, err := tw.Write(data)
		return err
	}
	must(add(&tar.Header{Name: "../unexpected", Mode: 0o644, Typeflag: tar.TypeReg}, []byte("fixture")))
	switch variant {
	case "success", "duplicate":
		must(add(&tar.Header{Name: "gumlet", Mode: 0o755, Typeflag: tar.TypeReg}, binary))
	case "symlink":
		must(add(&tar.Header{Name: "gumlet", Linkname: "../unexpected", Mode: 0o777, Typeflag: tar.TypeSymlink}, nil))
	}
	if variant == "duplicate" {
		must(add(&tar.Header{Name: "gumlet", Mode: 0o755, Typeflag: tar.TypeReg}, []byte("duplicate")))
	}
	must(tw.Close())
	return gz.Close()
}

func must(err error) {
	if err != nil {
		io.WriteString(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
}
