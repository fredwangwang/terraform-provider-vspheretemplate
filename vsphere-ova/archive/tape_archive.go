package archive

import (
	"archive/tar"
	"context"
	"errors"
	"github.com/vmware/govmomi/vim25/soap"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
)

type TapeArchive struct {
	path string
	Opener
}

type Opener struct {
	Downloader
}

type Downloader interface {
	Download(ctx context.Context, u *url.URL, param *soap.Download) (io.ReadCloser, int64, error)
}

func NewTapeArchive(path string, opener Opener) *TapeArchive {
	return &TapeArchive{
		path:   path,
		Opener: opener,
	}
}

type TapeArchiveEntry struct {
	io.Reader
	f io.Closer
}

func (e *TapeArchiveEntry) Close() error {
	return e.f.Close()
}

func (ta *TapeArchive) Open(name string) (io.ReadCloser, int64, error) {
	f, _, err := ta.OpenFile(ta.path)
	if err != nil {
		return nil, 0, err
	}

	r := tar.NewReader(f)

	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, err
		}

		matched, err := path.Match(name, path.Base(h.Name))
		if err != nil {
			return nil, 0, err
		}

		if matched {
			return &TapeArchiveEntry{r, f}, h.Size, nil
		}
	}

	_ = f.Close()

	return nil, 0, os.ErrNotExist
}

func (o *Opener) OpenFile(path string) (io.ReadCloser, int64, error) {
	if isRemotePath(path) {
		return o.OpenRemote(path)
	}
	return o.OpenLocal(path)
}

func (o Opener) OpenLocal(path string) (io.ReadCloser, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}

	return f, s.Size(), nil
}

func (o Opener) OpenRemote(link string) (io.ReadCloser, int64, error) {
	if o.Downloader == nil {
		return nil, 0, errors.New("remote path not supported")
	}

	u, err := url.Parse(link)
	if err != nil {
		return nil, 0, err
	}

	return o.Download(context.Background(), u, &soap.DefaultDownload)
}

func isRemotePath(path string) bool {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return true
	}
	return false
}
