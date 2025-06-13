// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package archive

import (
	"archive/tar"
	"bytes"
	"fmt"
	"os"
	"testing"

	"go.uber.org/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/runfinch/finch-daemon/mocks/mocks_ecc"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/ecc"
)

// TestContainerHandler function is the entry point of container handler package's unit test using ginkgo.
func TestContainerHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Archive Utils")
}

// Unit tests related to check RegisterHandlers() has configured the endpoint properly for containers related API.
var _ = Describe("TarExtractor's ", func() {
	Context("ExtractInTempDir method", func() {
		var (
			mockCtrl          *gomock.Controller
			cmdCreator        ecc.ExecCmdCreator
			logger            *mocks_logger.Logger
			tarExtractor      TarExtractor
			buf               bytes.Buffer
			dockerFileContent string
		)
		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()
			cmdCreator = ecc.NewExecCmdCreator()
			logger = mocks_logger.NewLogger(mockCtrl)
			tarExtractor = NewTarExtractor(cmdCreator, logger)
			tw := tar.NewWriter(&buf)
			dockerFileContent = "FROM alpine:latest"
			hdr := &tar.Header{
				Name: "Dockerfile",
				Mode: 0o600,
				Size: int64(len(dockerFileContent)),
			}
			_ = tw.WriteHeader(hdr)
			_, _ = tw.Write([]byte(dockerFileContent))
			_ = tw.Close()
		})
		It("should be able to extract a tar file in temp folder", func() {
			logger.EXPECT().Debugf("successfully cleaned up folder. path: %s", gomock.Any())

			cmd, err := tarExtractor.ExtractInTemp(bytes.NewReader(buf.Bytes()), "unit-test")
			Expect(err).ShouldNot(HaveOccurred())
			defer tarExtractor.Cleanup(cmd)
			err = cmd.Run()
			Expect(err).ShouldNot(HaveOccurred())
			dockerFile := fmt.Sprintf("%s/Dockerfile", cmd.GetDir())
			Expect(err).ShouldNot(HaveOccurred())
			_, err = os.Stat(dockerFile)
			Expect(err).ShouldNot(HaveOccurred())
			b, _ := os.ReadFile(dockerFile)
			Expect(string(b)).Should(Equal(dockerFileContent))
		})
	})
	Context("Cleanup method", func() {
		var (
			mockCtrl     *gomock.Controller
			logger       *mocks_logger.Logger
			mockCmd      *mocks_ecc.MockExecCmd
			tarExtractor TarExtractor
			dir          string
		)
		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			ecc := mocks_ecc.NewMockExecCmdCreator(mockCtrl)
			mockCmd = mocks_ecc.NewMockExecCmd(mockCtrl)
			logger = mocks_logger.NewLogger(mockCtrl)
			tarExtractor = NewTarExtractor(ecc, logger)
			dir, _ = os.MkdirTemp(os.TempDir(), "test")
			mockCmd.EXPECT().GetDir().Return(dir).AnyTimes()
		})
		AfterEach(func() {
			mockCtrl.Finish()
			// remove the folder if it is not deleted
			_ = os.RemoveAll(dir)
		})
		It("should be able be able to clean up files", func() {
			logger.EXPECT().Debugf("successfully cleaned up folder. path: %s", gomock.Any())
			tarExtractor.Cleanup(mockCmd)
			_, err := os.Stat(mockCmd.GetDir())
			Expect(errors.Is(err, os.ErrNotExist)).Should(BeTrue())
		})
		It("should skip cleaning as no path provided", func() {
			logger.EXPECT().Debugf("noting to clean up.")
			tarExtractor.Cleanup(nil)
		})
	})
})
