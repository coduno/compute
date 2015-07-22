package util

import (
	"io"
	"os"
	"path"
)

// CopyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func CopyFileContents(dst, src, fileName string) (err error) {
	os.MkdirAll(dst, 0777)
	in, err := os.Open(src)
	if err != nil {
		return
	}
	dst = dst + "/" + fileName
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	return
}

// CreateFile creates a file in a specific destination and with a specific
// content
func CreateFile(tmpDir, fileName, content string) (err error) {
	f, err := os.Create(path.Join(tmpDir, fileName))
	if err != nil {
		return
	}
	f.WriteString(content)
	f.Close()
	return
}
