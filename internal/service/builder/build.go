// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"

	"github.com/runfinch/finch-daemon/api/events"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

const tagEventAction = "tag"
const shortLen = 12

// setting publishTagEventFunc as a variable to allow mocking this function for unit testing.
var publishTagEventFunc = (*service).publishTagEvent

// Build function builds an image using nerdctl function based on the BuilderBuildOptions.
func (s *service) Build(ctx context.Context, options *ncTypes.BuilderBuildOptions, tarBody io.ReadCloser) ([]types.BuildResult, error) {
	tarCmd, err := s.tarExtractor.ExtractInTemp(tarBody, "build-context")
	if err != nil {
		s.logger.Warnf("Failed to extract build context. Error: %v", err)
		return nil, err
	}
	dir := tarCmd.GetDir()

	// create an in-memory writer for the stderr
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	tarCmd.SetStderr(writer)
	// clean up the directory before exiting the method
	defer s.tarExtractor.Cleanup(tarCmd)
	// execute the extract command
	if err = tarCmd.Run(); err != nil {
		s.logger.Warnf("Failed to extract build context in temp folder. Dir: %s, Error: %s, Stderr: %s",
			dir, err.Error(), buf.String())
		return nil, fmt.Errorf("failed to extract build context in temp folder")
	}

	// update the build context and the docker file path with the temp dir.
	options.BuildContext = dir
	options.File = fmt.Sprintf("%s/%s", dir, options.File)
	if err = s.nctlBuilderSvc.Build(ctx, s.client, *options); err != nil {
		return nil, err
	}

	// publish tag event for the built image
	result := []types.BuildResult{}
	if options.Tag != nil {
		for _, tag := range options.Tag {
			tagEvent, err := publishTagEventFunc(s, ctx, tag)
			if err != nil {
				return nil, err
			}
			result = append(result, types.BuildResult{ID: tagEvent.ID})
			options.Stdout.Write([]byte(fmt.Sprintf("Successfully built %s\n", tagEvent.ID)))
		}
	}

	return result, nil
}

func (s *service) publishTagEvent(ctx context.Context, tag string) (*events.Event, error) {
	_, uniqueCount, images, err := s.nctlBuilderSvc.SearchImage(ctx, tag)
	if err != nil {
		return nil, err
	}
	if uniqueCount == 0 || len(images) == 0 {
		return nil, errdefs.NewNotFound(fmt.Errorf("no such image: %s", tag))
	}
	if uniqueCount != 1 {
		return nil, fmt.Errorf("multiple images exist with tag %s", tag)
	}

	tagEvent := getTagEvent(images[0].Target.Digest.String(), tag)
	if err = s.client.PublishEvent(ctx, tagTopic(), tagEvent); err != nil {
		return nil, fmt.Errorf("failed to publish tag event for image %s: %s", tag, err)
	}
	return tagEvent, nil
}

func tagTopic() string {
	return fmt.Sprintf("/%s/%s/%s", events.CompatibleTopicPrefix, "image", tagEventAction)
}

func getTagEvent(digest, imgName string) *events.Event {
	return &events.Event{
		ID:     truncateID(digest), // for docker compatibility
		Status: tagEventAction,
		Type:   "image",
		Action: tagEventAction,
		Actor: events.EventActor{
			Id: digest,
			Attributes: map[string]string{
				"name": imgName,
			},
		},
	}
}

func truncateID(id string) string {
	if i := strings.IndexRune(id, ':'); i >= 0 {
		id = id[i+1:]
	}
	if len(id) > shortLen {
		id = id[:shortLen]
	}
	return id
}
