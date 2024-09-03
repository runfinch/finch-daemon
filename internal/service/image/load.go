// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/containerd/fifo"
)

func (s *service) Load(ctx context.Context, inStream io.Reader, outStream io.Writer, quiet bool) error {
	if inStream == nil {
		return fmt.Errorf("import stream should not be nil")
	}
	root, err := s.nctlImageSvc.GetDataStore()
	if err != nil {
		s.logger.Errorf("failed to get data store dir: %s", err)
		return err
	}
	d := filepath.Join(root, "fifo")
	if err = os.Mkdir(d, 0700); err != nil && !os.IsExist(err) {
		fmt.Println(err)
		s.logger.Errorf("failed to create fifo dir %s: %s", d, err)
		return err
	}
	t := time.Now()
	var b [3]byte
	rand.Read(b[:])
	img := filepath.Join(d, fmt.Sprintf("%d%s", t.Nanosecond(), base64.URLEncoding.EncodeToString(b[:])))
	rw, err := fifo.OpenFifo(ctx, img, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_NONBLOCK, 0700)
	if err != nil {
		s.logger.Errorf("failed to open fifo %s: %s", img, err)
		return err
	}
	defer func() {
		rw.Close()
		os.Remove(img)
	}()
	go func() {
		io.Copy(rw, inStream)
	}()
	if err = s.nctlImageSvc.LoadImage(ctx, img, outStream, quiet); err != nil {
		s.logger.Errorf("failed to load image %s: %s", img, err)
		return err
	}
	return nil
}
