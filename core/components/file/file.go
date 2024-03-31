package file

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"jasonzhu.com/coin_labor/core/setting"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"jasonzhu.com/coin_labor/core/components/log"
)

/**

@author Jason
@version 2021-05-06 21:27
*/

var (
	RotatedLayout = "150405.2006-01-02"
	separator     = "/"
	lg            = log.New("file")
)

// RotateWriter writes and rotates files
type RotateWriter struct {
	lock            sync.Mutex
	filename        string
	rotatedFilename string
	fp              *os.File
	maxsize         int64
}

// New makes a new RotateWriter. Return nil if error occurs during setup.
func New(filename string, maxsize int64) (w *RotateWriter, err error) {
	w = &RotateWriter{
		filename: filename,
		maxsize:  maxsize,
	}
	w.fp, err = os.OpenFile(getPath(filename), syscall.O_RDWR|syscall.O_CREAT, 0666)
	if err != nil {
		return nil, err
	}
	return w, nil
}

// Write writes a  new line string to file
func (w *RotateWriter) Write(line string) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.fp.WriteString(fmt.Sprintf("%s\n", line))
}

// Rotateable checks if the file is ready for rotation. Meaning that it reached the size.
func (w *RotateWriter) Rotateable() (bool, error) {
	fi, err := os.Stat(getPath(w.filename))
	if err != nil {
		return false, errors.Wrap(err, "OS File stat")
	}
	size := fi.Size()
	if size > w.maxsize {
		return true, nil
	}
	return false, nil
}

// Rotate performs the actual act of rotating and reopening file. Returns the rotated filename
func (w *RotateWriter) Rotate() (string, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	lg.Debug("File rotation started")

	// Close existing file if open
	if w.fp != nil {
		err := w.fp.Close()
		w.fp = nil
		if err != nil {
			return w.rotatedFilename, errors.Wrap(err, "OS File close")
		}
	}
	// Rename the dest file if it already exists
	_, err := os.Stat(getPath(w.filename))
	if err == nil {
		// avoid hot spot data
		w.rotatedFilename = time.Now().Format(RotatedLayout) + "." + w.filename
		err = os.Rename(getPath(w.filename), getPath(w.rotatedFilename))
		if err != nil {
			return w.rotatedFilename, errors.Wrap(err, "OS File rename")
		}
	}

	// Create the file.
	w.fp, err = os.Create(getPath(w.filename))
	if err != nil {
		return w.rotatedFilename, errors.Wrap(err, "OS File create")
	}
	return w.rotatedFilename, nil
}

// Compress compresses plain files
func Compress(filename string) (string, error) {
	compressed := filename + ".gz"
	lg.Debug("Compressing", "filename", filename)

	// Open file on disk.
	f, err := os.Open(getPath(filename))
	if err != nil {
		return "", errors.Wrap(err, "File Compress")
	}

	// Create a Reader and use ReadAll to get all the bytes from the file.
	reader := bufio.NewReader(f)
	content, _ := ioutil.ReadAll(reader)

	// Open file for writing.
	f, err = os.Create(getPath(compressed))
	if err != nil {
		return "", errors.Wrap(err, "File Compress")
	}

	// Write compressed data.
	w := gzip.NewWriter(f)
	w.Write(content)
	w.Close()

	// Remove old
	os.Remove(getPath(filename))

	// Done.
	lg.Info("File compressed", "filename", filename)
	return compressed, nil
}

func getPath(filename string) string {
	return setting.AppConfig.GeneratedFile.Home + separator + filename
}
