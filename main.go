package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/ethersphere/bee/pkg/bmtpool"
)

func main() {
	prefix := os.Args[1]
	fmt.Println(prefix)
	s, err := hex.DecodeString(prefix)
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
	prefixLen := len(s)
	runs := 0
	defer func() { fmt.Println("ran", runs, "times") }()
	i := 64
	b := make([]byte, i)
	n, err := rand.Read(b)
	if n != i {
		panic("short read")
	}
	if err != nil {
		panic(err)
	}

	sem := make(chan struct{}, 8)
	done := make(chan struct{})
	for i := 1; ; i++ {
		sem <- struct{}{}
		runs++
		select {
		case <-done:
			return
		default:
		}
		go func(i int) {
			bmt := bmtpool.Get()
			defer func() {
				bmtpool.Put(bmt)
				<-sem
			}()
			nonce := make([]byte, 8)
			binary.LittleEndian.PutUint64(nonce, uint64(i))
			bbb := append(nonce, b...)
			span := make([]byte, 8)
			binary.LittleEndian.PutUint64(span, uint64(len(bbb)))
			bbb = append(span, bbb...)
			bmt.SetSpanBytes(span)
			bmt.Write(bbb[8:])
			hash := bmt.Sum(nil)
			if bytes.Equal(s, hash[:prefixLen]) {
				fmt.Printf("caddr := %s\n", hex.EncodeToString(hash))
				data := fmt.Sprintf("%v", bbb)
				data = strings.Replace(data, " ", ",", -1)
				data = strings.Replace(data, "[", "{", -1)
				data = strings.Replace(data, "]", "}", -1)

				fmt.Println(fmt.Sprintf("cdata := []byte %s", data))
				close(done)
			}
		}(i)
	}
}
