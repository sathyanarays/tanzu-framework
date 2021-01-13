// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultiRepo(t *testing.T) {
	repoNew := newTestRepo(t, "artifacts-new")
	repoAlt := newTestRepo(t, "artifacts-alt")

	m := NewMultiRepo(repoNew)
	m.AddRepository(repoAlt)

	r, err := m.Find("foo")
	require.NoError(t, err)

	require.Equal(t, r.Name(), "artifacts-new")

	r, err = m.Find("qux")
	require.NoError(t, err)

	require.Equal(t, r.Name(), "artifacts-alt")

	_, err = m.GetRepository("artifacts-new")
	require.NoError(t, err)

	mp, err := m.ListPlugins()
	require.NoError(t, err)
	require.Len(t, mp, 2)

	m.RemoveRepository("artifacts-alt")
	mp, err = m.ListPlugins()
	require.NoError(t, err)
	require.Len(t, mp, 1)

	// test duplicates
	repoOld := newTestRepo(t, "artifacts-old")
	m.AddRepository(repoOld)

	r, err = m.Find("foo")
	require.Error(t, err)

	r, err = m.Find("artifacts-new.foo")
	require.NoError(t, err)
}
