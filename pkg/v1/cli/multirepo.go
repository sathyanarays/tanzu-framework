// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import "fmt"

// KnownRepositories is a list of known repositories.
var KnownRepositories = []Repository{
	CommunityGCPBucketRepository,
	TMCGCPBucketRepository,
}

// DefaultMultiRepo is the default multirepo with the known repositories.
var DefaultMultiRepo = NewMultiRepo(KnownRepositories...)

// MultiRepo is a multiple
type MultiRepo struct {
	repositories []Repository
}

// NewMultiRepo returns a new multirepo.
func NewMultiRepo(repositories ...Repository) *MultiRepo {
	return &MultiRepo{
		repositories: repositories,
	}
}

// AddRepository to known.
func (m *MultiRepo) AddRepository(repo Repository) {
	m.repositories = append(m.repositories, repo)
}

// RemoveRepository removes a repo.
func (m *MultiRepo) RemoveRepository(name string) {
	newRepos := []Repository{}
	for _, repo := range m.repositories {
		if name != repo.Name() {
			newRepos = append(newRepos, repo)
		}
	}
	m.repositories = newRepos
}

// GetRepository returns a repository.
func (m *MultiRepo) GetRepository(name string) (Repository, error) {
	for _, repo := range m.repositories {
		if name == repo.Name() {
			return repo, nil
		}
	}
	return nil, fmt.Errorf("could not find repository %q", name)
}

// ListPlugins across the repositories.
func (m *MultiRepo) ListPlugins() (mp map[string][]PluginDescriptor, err error) {
	mp = map[string][]PluginDescriptor{}
	for _, repo := range m.repositories {
		descriptors, err := repo.List()
		if err != nil {
			return mp, err
		}
		mp[repo.Name()] = descriptors
	}
	return
}

// Find a repository for the given plugin name.
func (m *MultiRepo) Find(name string) (r Repository, err error) {
	matches := []Repository{}
	for _, repo := range m.repositories {
		descriptors, err := repo.List()
		if err != nil {
			return r, err
		}
		for _, desc := range descriptors {
			if desc.Name == name {
				matches = append(matches, repo)
			} else if fmt.Sprintf("%s.%s", repo.Name(), desc.Name) == name {
				matches = append(matches, repo)
			}
		}
	}

	switch i := len(matches); i {
	case 0:
		return nil, fmt.Errorf("could not find plugin %q", name)
	case 1:
		return matches[0], nil
	default:
		return nil, fmt.Errorf("found plugin %q in %#v repositories, use the fully qualified <repository-name>.<plugin-name> to select", name, matches)
	}
}
