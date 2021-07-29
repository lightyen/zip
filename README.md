# zip

Fork from: "github.com/alexmullins/zip"

```go
package main

import (
	"bytes"
	"log"
	"os"
	"github.com/lightyen/zip"
)

func main() {
	contents := []byte("Hello World")
	fzip, err := os.Create(`./test.zip`)
	if err != nil {
		log.Fatalln(err)
	}
	zipw := zip.NewWriter(fzip)
	defer zipw.Close()

	hdr := &zip.FileHeader{Name: "test.txt", Method: zip.Deflate}
	password := "golang"
	w, err := zipw.Encrypt(hdr, password)
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(w, bytes.NewReader(contents))
	if err != nil {
		log.Fatal(err)
	}

	zipw.Flush()
}
```
