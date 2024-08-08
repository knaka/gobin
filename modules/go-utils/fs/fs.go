package fs

import (
	. "github.com/knaka/go-utils"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CanonPath returns the canonical absolute path of the given value.
func CanonPath(s string) (ret string, err error) {
	ret, err = filepath.Abs(s)
	if err != nil {
		return
	}
	ret, err = filepath.EvalSymlinks(ret)
	if err != nil {
		return
	}
	ret = filepath.Clean(ret)
	return
}

func copyFile(src, dst string) (err error) {
	reader := V(os.Open(src))
	defer (func() { V0(reader.Close()) })()
	writer := V(os.Create(dst))
	defer (func() { Ignore(writer.Close()) })()
	V0(io.Copy(writer, reader))
	V0(writer.Close())
	statSrc := V(os.Stat(src))
	V0(os.Chmod(dst, statSrc.Mode()))
	V0(os.Chtimes(dst, statSrc.ModTime(), statSrc.ModTime()))
	return
}

func copyDir(srcDir string, dstDir string) (err error) {
	return filepath.Walk(srcDir, func(srcObjPath string, srcObjStat fs.FileInfo, errGiven error) (err error) {
		if errGiven != nil {
			return errGiven
		}
		dstObj := filepath.Join(dstDir, strings.TrimPrefix(srcObjPath, srcDir))
		if srcObjStat.IsDir() {
			err = os.MkdirAll(dstObj, srcObjStat.Mode())
			if err != nil {
				return
			}
		} else if !srcObjStat.Mode().IsRegular() {
			switch srcObjStat.Mode().Type() & os.ModeType {
			case os.ModeSymlink:
				linkDst, err_ := os.Readlink(srcObjPath)
				if err_ != nil {
					return err_
				}
				err = os.Symlink(linkDst, dstObj)
				if err != nil {
					return
				}
			default:
				err = nil
			}
		} else {
			return copyFile(srcObjPath, dstObj)
		}
		err = os.Chtimes(dstObj, srcObjStat.ModTime(), srcObjStat.ModTime())
		return err
	})
}

// Copy copies a file or a directory.
func Copy(src string, dst string) (err error) {
	if stat, err_ := os.Stat(src); err_ != nil {
		return err_
	} else if stat.IsDir() {
		if _, err_ := os.Stat(dst); err_ == nil {
			Ignore(os.RemoveAll(dst))
		}
		return copyDir(src, dst)
	} else {
		if _, err_ := os.Stat(dst); err_ == nil {
			Ignore(os.RemoveAll(dst))
		}
		return copyFile(src, dst)
	}
}

// Move moves a file or a directory.
func Move(src string, dst string) (err error) {
	err = Copy(src, dst)
	if err != nil {
		return
	}
	return os.RemoveAll(src)
}

func Touch(path string) (err error) {
	_, err = os.Stat(path)
	if err != nil {
		// If not exists, create an empty file.
		if os.IsNotExist(err) {
			file, err_ := os.Create(path)
			if err_ != nil {
				return err_
			}
			return file.Close()
		}
		return err
	}
	// If exists, update the timestamp.
	now := time.Now()
	return os.Chtimes(path, now, now)
}
