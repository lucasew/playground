package main

import (
    "io"
    "compress/gzip"
    "archive/tar"
    "path"
    "os"
    "log"
)

func main() {
    if len(os.Args) != 3 {
        log.Fatalf("expected 2 arguments: [tar file] [folder]. got %d arguments", len(os.Args))
    }
    var tarFile = os.Args[1]
    var folder = os.Args[2]
    tarf, err := os.Open(tarFile)
    if err != nil {
        log.Fatal(err)
    }
    defer tarf.Close()
    Decompress(folder, tarf)
}

// https://stackoverflow.com/questions/23629080/how-to-implement-tar-cvfz-xxx-tar-gz-in-golang
// Why reinvent the wheel?
func Decompress(targetdir string, reader io.ReadCloser) error {
    gzReader, err := gzip.NewReader(reader)
    if err != nil {
        return err
    }
    defer gzReader.Close()

    tarReader := tar.NewReader(gzReader)
    for {
        header, err := tarReader.Next()
        if err == io.EOF {
            break
        } else if err != nil {
            return err
        }

        target := path.Join(targetdir, header.Name)
        switch header.Typeflag {
        case tar.TypeDir:
            err = os.MkdirAll(target, os.FileMode(header.Mode))
            if err != nil {
                return err
            }

            setAttrs(target, header)
            break

        case tar.TypeReg:
            w, err := os.Create(target)
            if err != nil {
                return err
            }
            _, err = io.Copy(w, tarReader)
            if err != nil {
                return err
            }
            w.Close()

            setAttrs(target, header)
            break

        default:
            log.Printf("unsupported type: %v", header.Typeflag)
            break
        }
    }

    return nil
}

func setAttrs(target string, header *tar.Header) {
    os.Chmod(target, os.FileMode(header.Mode))
    os.Chtimes(target, header.AccessTime, header.ModTime)
}
