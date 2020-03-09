package common

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	v1 "github.com/coreydaley/crawler/api/v1"
)

func PrintTree(root v1.Node, depth int) {
	for _, n := range *root.Children {
		fmt.Println(fmt.Sprintf("%s+--%s", strings.Repeat(" ", depth*4), n.Name))
		if n.Children != nil && len(*n.Children) != 0 {
			PrintTree(n, depth+1)
		}
	}
}

func EncodeToBytes(p interface{}) []byte {

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(p)
	if err != nil {
		log.Fatal(err)
	}

	return buf.Bytes()
}

func Compress(s []byte) []byte {
	fmt.Println("uncompressed size (bytes): ", len(s))
	zipbuf := bytes.Buffer{}
	zipped := gzip.NewWriter(&zipbuf)
	zipped.Write(s)
	zipped.Close()
	fmt.Println("compressed size (bytes): ", len(zipbuf.Bytes()))
	return zipbuf.Bytes()
}

func Decompress(s []byte) []byte {
	fmt.Println("compressed size (bytes): ", len(s))
	rdr, _ := gzip.NewReader(bytes.NewReader(s))
	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		log.Fatal(err)
	}
	rdr.Close()
	fmt.Println("uncompressed size (bytes): ", len(data))
	return data
}

func DecodeToNode(s []byte) v1.Node {

	p := v1.Node{}
	dec := gob.NewDecoder(bytes.NewReader(s))
	err := dec.Decode(&p)
	if err != nil {
		log.Fatal(err)
	}
	return p
}
