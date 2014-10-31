// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package tools

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/juju/errors"
	"github.com/juju/utils"

	"github.com/juju/juju/environs"
	conf "github.com/juju/juju/environs/config"
	"github.com/juju/juju/environs/simplestreams"
	"github.com/juju/juju/environs/storage"
)

type toolsDatasourceFuncId struct {
	id string
	f  ToolsDataSourceFunc
}

var (
	toolsDatasourceFuncsMu sync.RWMutex
	toolsDatasourceFuncs   []toolsDatasourceFuncId
)

// ToolsDataSourceFunc is a function type that takes an environment and
// returns a simplestreams datasource.
//
// ToolsDataSourceFunc will be used in GetMetadataSources.
// Any error satisfying errors.IsNotSupported will be ignored;
// any other error will be cause GetMetadataSources to fail.
type ToolsDataSourceFunc func(environs.Environ) (simplestreams.DataSource, error)

// RegisterToolsDataSourceFunc registers an ToolsDataSourceFunc
// with the specified id, overwriting any function previously registered
// with the same id.
func RegisterToolsDataSourceFunc(id string, f ToolsDataSourceFunc) {
	toolsDatasourceFuncsMu.Lock()
	defer toolsDatasourceFuncsMu.Unlock()
	for i := range toolsDatasourceFuncs {
		if toolsDatasourceFuncs[i].id == id {
			toolsDatasourceFuncs[i].f = f
			return
		}
	}
	toolsDatasourceFuncs = append(toolsDatasourceFuncs, toolsDatasourceFuncId{id, f})
}

// UnregisterToolsDataSourceFunc unregisters an ToolsDataSourceFunc
// with the specified id.
func UnregisterToolsDataSourceFunc(id string) {
	toolsDatasourceFuncsMu.Lock()
	defer toolsDatasourceFuncsMu.Unlock()
	for i, f := range toolsDatasourceFuncs {
		if f.id == id {
			head := toolsDatasourceFuncs[:i]
			tail := toolsDatasourceFuncs[i+1:]
			toolsDatasourceFuncs = append(head, tail...)
			return
		}
	}
}

// GetMetadataSources returns the sources to use when looking for
// simplestreams tools metadata for the given stream.
func GetMetadataSources(env environs.Environ) ([]simplestreams.DataSource, error) {
	config := env.Config()

	// Add configured and environment-specific datasources.
	var sources []simplestreams.DataSource
	if userURL, ok := config.AgentMetadataURL(); ok {
		verify := utils.VerifySSLHostnames
		if !config.SSLHostnameVerification() {
			verify = utils.NoVerifySSLHostnames
		}
		sources = append(sources, simplestreams.NewURLDataSource(conf.AgentMetadataURLKey, userURL, verify))
	}

	envDataSources, err := environmentDataSources(env)
	if err != nil {
		return nil, err
	}
	sources = append(sources, envDataSources...)

	// Add the default, public datasource.
	defaultURL, err := ToolsURL(DefaultBaseURL)
	if err != nil {
		return nil, err
	}
	if defaultURL != "" {
		sources = append(sources,
			simplestreams.NewURLDataSource("default simplestreams", defaultURL, utils.VerifySSLHostnames))
	}
	return sources, nil
}

// environmentDataSources returns simplestreams datasources for the environment
// by calling the functions registered in RegisterToolsDataSourceFunc.
// The datasources returned will be in the same order the functions were registered.
func environmentDataSources(env environs.Environ) ([]simplestreams.DataSource, error) {
	toolsDatasourceFuncsMu.RLock()
	defer toolsDatasourceFuncsMu.RUnlock()
	var datasources []simplestreams.DataSource
	for _, f := range toolsDatasourceFuncs {
		logger.Debugf("trying datasource %q", f.id)
		datasource, err := f.f(env)
		if err != nil {
			if errors.IsNotSupported(err) {
				continue
			}
			return nil, err
		}
		datasources = append(datasources, datasource)
	}
	return datasources, nil
}

// ToolsURL returns a valid tools URL constructed from source.
// source may be a directory, or a URL like file://foo or http://foo.
func ToolsURL(source string) (string, error) {
	if source == "" {
		return "", nil
	}
	// If source is a raw directory, we need to append the file:// prefix
	// so it can be used as a URL.
	defaultURL := source
	u, err := url.Parse(defaultURL)
	if err != nil {
		return "", fmt.Errorf("invalid default tools URL %s: %v", defaultURL, err)
	}

	_, err = os.Stat(defaultURL)
	if u.Scheme == "" || err == nil {
		defaultURL = utils.MakeFileURL(defaultURL)
		if !strings.HasSuffix(defaultURL, "/"+storage.BaseToolsPath) {
			defaultURL = fmt.Sprintf("%s/%s", defaultURL, storage.BaseToolsPath)
		}
	}
	return defaultURL, nil
}
