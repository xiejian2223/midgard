// Copyright 2020 The golang.design Initiative authors.
// All rights reserved. Use of this source code is governed by
// a GNU GPL-3.0 license that can be found in the LICENSE file.

package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"golang.design/x/midgard/pkg/clipboard"
	"golang.design/x/midgard/pkg/config"
	"golang.design/x/midgard/pkg/types"
	"golang.design/x/midgard/pkg/types/proto"
	"golang.design/x/midgard/pkg/utils"
	"golang.design/x/midgard/pkg/version"
)

// Ping response a pong
func (m *Daemon) Ping(ctx context.Context, in *proto.PingInput) (*proto.PingOutput, error) {
	return &proto.PingOutput{
		Version:   version.GitVersion,
		GoVersion: version.GoVersion,
		BuildTime: version.BuildTime,
	}, nil
}

// AllocateURL request the midgard server to allocate a given URL for
// a given resource, or the content from the midgard universal clipboard.
func (m *Daemon) AllocateURL(ctx context.Context, in *proto.AllocateURLInput) (*proto.AllocateURLOutput, error) {
	var (
		source = types.SourceUniversalClipboard
		data   string
		uri    string
	)

	// TODO: allocate urls using websocket action

	if in.SourcePath != "" {
		source = types.SourceAttachment
		b, err := ioutil.ReadFile(in.SourcePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %v, err: %w", in.SourcePath, err)
		}
		data = utils.BytesToString(b)

	}
	if in.DesiredPath != "" {
		// we want to make sure the extension of the file is correct
		dext := filepath.Ext(in.DesiredPath)
		sext := filepath.Ext(in.SourcePath)
		uri = strings.TrimSuffix(in.DesiredPath, dext) + sext
	}

	res, err := utils.Request(
		http.MethodPut,
		types.EndpointAllocateURL,
		&types.AllocateURLInput{
			Source: source,
			URI:    uri,
			Data:   data,
		})
	if err != nil {
		return nil, fmt.Errorf("cannot perform allocate request, err %w", err)
	}
	var out types.AllocateURLOutput
	err = json.Unmarshal(res, &out)
	if err != nil {
		return nil, fmt.Errorf("cannot parse requested URL, err: %w", err)
	}
	if out.URL == "" {
		return nil, fmt.Errorf("%s", out.Message)
	}

	url := config.Get().Domain + out.URL
	clipboard.Write(utils.StringToBytes(url))
	return &proto.AllocateURLOutput{URL: url, Message: "Done."}, nil
}

// CreateNews creates a news
func (m *Daemon) CreateNews(ctx context.Context, in *proto.CreateNewsInput) (out *proto.CreateNewsOutput, err error) {

	s := &types.ActionCreateNewsData{
		in.Date, in.Title, in.Body,
	}
	b, _ := json.Marshal(s)

	m.writeCh <- &types.WebsocketMessage{
		Action:  types.ActionCreateNews,
		UserID:  m.ID,
		Message: "create news",
		Data:    b,
	}

	return &proto.CreateNewsOutput{Message: "DONE."}, nil
}
