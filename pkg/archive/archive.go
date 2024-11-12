// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package archive

import (
	"io"
	"os"
	"path"

	"github.com/containerd/nerdctl/v2/pkg/tarutil"
	"github.com/docker/docker/pkg/archive"

	"github.com/runfinch/finch-daemon/pkg/ecc"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

//go:generate mockgen --destination=../../mocks/mocks_archive/tarcreator.go -package=mocks_archive github.com/runfinch/finch-daemon/pkg/archive TarCreator
type TarCreator interface {
	CreateTarCommand(srcPath string, slashDot bool) (ecc.ExecCmd, error)
}

type tarCreator struct {
	ecc    ecc.ExecCmdCreator
	logger flog.Logger
}

func NewTarCreator(ecc ecc.ExecCmdCreator, logger flog.Logger) TarCreator {
	return &tarCreator{
		ecc:    ecc,
		logger: logger,
	}
}

// CreateTarCommand creates an *exec.Cmd that will output a tar archive of the provided srcPath to stdout.
func (c *tarCreator) CreateTarCommand(srcPath string, slashDot bool) (ecc.ExecCmd, error) {
	tarBinary, _, err := tarutil.FindTarBinary()
	if err != nil {
		c.logger.Debugf("error getting tar binary: %s", err.Error())
		return nil, err
	}

	// "/." is a Docker thing that instructions the copy command to download contents of the folder only
	var tarDir, tarPath string
	if slashDot {
		tarDir = srcPath
		tarPath = "."
	} else {
		tarDir = path.Dir(srcPath)
		tarPath = path.Base(srcPath)
	}

	cmd := c.ecc.Command(tarBinary, []string{"-c", "-f", "-", tarPath}...)
	cmd.SetDir(tarDir)
	return cmd, nil
}

// TarExtractor interface to extract a tar file
//
//go:generate mockgen --destination=../../mocks/mocks_archive/tarextractor.go -package=mocks_archive github.com/runfinch/finch-daemon/pkg/archive TarExtractor
type TarExtractor interface {
	ExtractInTemp(reader io.Reader, dirPrefix string) (ecc.ExecCmd, error)
	CreateExtractCmd(reader io.Reader, destDir string) (ecc.ExecCmd, error)
	Cleanup(cmd ecc.ExecCmd)
	ExtractCompressed(tarArchive io.Reader, dest string, options *archive.TarOptions) error
}

// tarExtractor struct in an implementation of TarExtractor. It extracts uncompressed tar file.
type tarExtractor struct {
	ecc    ecc.ExecCmdCreator
	logger flog.Logger
}

// NewTarExtractor creates a new UncompressedTarExtractor.
func NewTarExtractor(ecc ecc.ExecCmdCreator, logger flog.Logger) TarExtractor {
	return &tarExtractor{
		ecc:    ecc,
		logger: logger,
	}
}

// ExtractInTemp is implementation of TarExtractor interface. This function extracts a tar file in a dest dir.
func (ext *tarExtractor) ExtractInTemp(reader io.Reader, dirPrefix string) (ecc.ExecCmd, error) {
	dir, err := os.MkdirTemp(os.TempDir(), dirPrefix)
	if err != nil {
		ext.logger.Errorf("Failed to extract in %s, error: %s", dir, err.Error())
		return nil, err
	}
	return ext.CreateExtractCmd(reader, dir)
}

// CreateExtractCmd is implementation of TarExtractor interface. This function extracts a tar file in a dest dir.
func (ext *tarExtractor) CreateExtractCmd(reader io.Reader, destDir string) (ecc.ExecCmd, error) {
	tarBinary, _, err := tarutil.FindTarBinary()
	if err != nil {
		ext.logger.Debugf("error getting tar binary: %s", err.Error())
		return nil, err
	}
	cmd := ext.ecc.Command(tarBinary, []string{"-x", "-f", "-"}...)
	cmd.SetStdin(reader)
	cmd.SetDir(destDir)
	return cmd, nil
}

// Cleanup function delete the extracted directory.
func (ext *tarExtractor) Cleanup(cmd ecc.ExecCmd) {
	// clean up not required
	if cmd == nil {
		ext.logger.Debugf("noting to clean up.")
		return
	}
	if err := os.RemoveAll(cmd.GetDir()); err != nil {
		ext.logger.Debugf("unable to cleanup folder. path: %s", cmd.GetDir())
	} else {
		ext.logger.Debugf("successfully cleaned up folder. path: %s", cmd.GetDir())
	}
}

// Wraps https://github.com/moby/moby/blob/master/pkg/archive/archive.go#L1233
// ExtractCompressed reads a stream of bytes from `archive`, parses it as a tar archive,
// and unpacks it into the directory at `dest`.
// The archive may be compressed with one of the following algorithms:
// identity (uncompressed), gzip, bzip2, xz.

func (ext *tarExtractor) ExtractCompressed(tarArchive io.Reader, dest string, options *archive.TarOptions) error {
	return archive.Untar(tarArchive, dest, options)
}
