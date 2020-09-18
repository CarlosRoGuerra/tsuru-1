// Copyright 2017 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provision_test

import (
	"context"

	"github.com/tsuru/config"
	"github.com/tsuru/tsuru/app/bind"
	"github.com/tsuru/tsuru/app/version"
	"github.com/tsuru/tsuru/db"
	"github.com/tsuru/tsuru/db/dbtest"
	"github.com/tsuru/tsuru/provision"
	"github.com/tsuru/tsuru/provision/provisiontest"
	_ "github.com/tsuru/tsuru/storage/mongodb"
	appTypes "github.com/tsuru/tsuru/types/app"
	check "gopkg.in/check.v1"
)

type S struct{}

var _ = check.Suite(&S{})

func (s *S) SetUpTest(c *check.C) {
	config.Set("database:url", "127.0.0.1:27017?maxPoolSize=100")
	config.Set("database:name", "provision_tests_2_s")
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	err = dbtest.ClearAllCollections(conn.Apps().Database)
	c.Assert(err, check.IsNil)
}

func (s *S) TestWebProcessDefaultPort(c *check.C) {
	port := provision.WebProcessDefaultPort()
	c.Assert(port, check.Equals, "8888")
}

func (s *S) TestWebProcessDefaultPortWithConfig(c *check.C) {
	config.Set("docker:run-cmd:port", "9191")
	defer config.Unset("docker:run-cmd:port")
	port := provision.WebProcessDefaultPort()
	c.Assert(port, check.Equals, "9191")
}

func (s *S) TestEnvsForApp(c *check.C) {
	a := provisiontest.NewFakeApp("myapp", "crystal", 1)
	a.SetEnv(bind.EnvVar{Name: "e1", Value: "v1"})
	envs := provision.EnvsForApp(a, "p1", false, nil)
	c.Assert(envs, check.DeepEquals, []bind.EnvVar{
		{Name: "e1", Value: "v1"},
		{Name: "TSURU_PROCESSNAME", Value: "p1"},
		{Name: "TSURU_HOST", Value: ""},
		{Name: "port", Value: "8888"},
		{Name: "PORT", Value: "8888"},
	})
	envs = provision.EnvsForApp(a, "p1", true, nil)
	c.Assert(envs, check.DeepEquals, []bind.EnvVar{
		{Name: "TSURU_HOST", Value: ""},
	})
}

func (s *S) TestEnvsForAppWithVersion(c *check.C) {
	a := provisiontest.NewFakeApp("myapp", "crystal", 1)
	a.SetEnv(bind.EnvVar{Name: "e1", Value: "v1"})

	svc, err := version.AppVersionService()
	c.Assert(err, check.IsNil)
	version, err := svc.NewAppVersion(context.TODO(), appTypes.NewVersionArgs{App: a})
	c.Assert(err, check.IsNil)

	envs := provision.EnvsForApp(a, "p1", false, version)
	c.Assert(envs, check.DeepEquals, []bind.EnvVar{
		{Name: "e1", Value: "v1"},
		{Name: "TSURU_PROCESSNAME", Value: "p1"},
		{Name: "TSURU_APPVERSION", Value: "1"},
		{Name: "TSURU_HOST", Value: ""},
		{Name: "port", Value: "8888"},
		{Name: "PORT", Value: "8888"},
	})
	envs = provision.EnvsForApp(a, "p1", true, version)
	c.Assert(envs, check.DeepEquals, []bind.EnvVar{
		{Name: "TSURU_HOST", Value: ""},
	})
}

func (s *S) TestEnvsForAppCustomConfig(c *check.C) {
	config.Set("host", "cloud.tsuru.io")
	config.Set("docker:run-cmd:port", "8989")
	defer config.Unset("host")
	defer config.Unset("docker:run-cmd:port")
	a := provisiontest.NewFakeApp("myapp", "crystal", 1)
	a.SetEnv(bind.EnvVar{Name: "e1", Value: "v1"})
	envs := provision.EnvsForApp(a, "p1", false, nil)
	c.Assert(envs, check.DeepEquals, []bind.EnvVar{
		{Name: "e1", Value: "v1"},
		{Name: "TSURU_PROCESSNAME", Value: "p1"},
		{Name: "TSURU_HOST", Value: "cloud.tsuru.io"},
		{Name: "port", Value: "8989"},
		{Name: "PORT", Value: "8989"},
	})
	envs = provision.EnvsForApp(a, "p1", true, nil)
	c.Assert(envs, check.DeepEquals, []bind.EnvVar{
		{Name: "TSURU_HOST", Value: "cloud.tsuru.io"},
	})
}
