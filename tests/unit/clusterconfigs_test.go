package unittest

import (
	"errors"
	"strconv"
	"gopkg.in/check.v1"
	"github.com/radanalyticsio/oshinko-core/clusterconfigs"
	//"github.com/radanalyticsio/oshinko-rest/models"
	"fmt"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/watch"
)

var tiny clusterconfigs.ClusterConfig = clusterconfigs.ClusterConfig{MasterCount: 1, WorkerCount: 0, Name: "tiny"}
var small clusterconfigs.ClusterConfig = clusterconfigs.ClusterConfig{MasterCount: 1, WorkerCount: 3,
	SparkMasterConfig: "master-config", SparkWorkerConfig: "worker-config", Name: "small"}
var large clusterconfigs.ClusterConfig = clusterconfigs.ClusterConfig{MasterCount: 0, WorkerCount: 10, Name: "large"}
var brokenMaster clusterconfigs.ClusterConfig = clusterconfigs.ClusterConfig{MasterCount: 2, WorkerCount: 0, Name: "brokenmaster"}
var brokenWorker clusterconfigs.ClusterConfig = clusterconfigs.ClusterConfig{MasterCount: 1, WorkerCount: -1, Name: "brokenworker"}
var nonIntMaster clusterconfigs.ClusterConfig = clusterconfigs.ClusterConfig{Name: "cow"}
var nonIntWorker clusterconfigs.ClusterConfig = clusterconfigs.ClusterConfig{Name: "pig"}
var userDefault = clusterconfigs.ClusterConfig{MasterCount: 3, WorkerCount: 3,
	SparkMasterConfig: "master-default", SparkWorkerConfig: "worker-default", Name: "default"}

func makeConfigMap(cfg clusterconfigs.ClusterConfig) *api.ConfigMap {
	var res api.ConfigMap = api.ConfigMap{Data: map[string]string{}}
	if cfg.SparkMasterConfig != "" {
		res.Data["sparkmasterconfig"] = cfg.SparkMasterConfig
	}
	if cfg.SparkWorkerConfig != "" {
		res.Data["sparkworkerconfig"] = cfg.SparkWorkerConfig
	}
	if cfg.MasterCount != 0 {
		res.Data["mastercount"] = strconv.Itoa(cfg.MasterCount)
	}
	if cfg.WorkerCount != 0 {
		res.Data["workercount"] = strconv.Itoa(cfg.WorkerCount)
	}
	res.Name = cfg.Name
	return &res
}

func addLineFeeds(cmap *api.ConfigMap) {
	for _, v := range cmap.Data {
		v = "\n" + v + "\n"
	}
}

// We need something that implements the kube ConfigMapsInterface since we
// are not conntected to a real client
type FakeConfigMapsClient struct {
	Configs api.ConfigMapList
}

func (f *FakeConfigMapsClient) Get(name string) (*api.ConfigMap, error) {
	for c := range f.Configs.Items {
		if f.Configs.Items[c].Name == name {
			return &f.Configs.Items[c], nil
		}
	}
	return nil, errors.New(fmt.Sprintf(clusterconfigs.NamedConfigDoesNotExist, name))
}

func (f *FakeConfigMapsClient) List(opts api.ListOptions) (*api.ConfigMapList, error) {
	return nil, nil
}

func (c *FakeConfigMapsClient) Create(cfg *api.ConfigMap) (*api.ConfigMap, error) {
	c.Configs.Items = append(c.Configs.Items, *cfg)
	return cfg, nil
}

func (f *FakeConfigMapsClient)Delete(string) error {
	return nil
}

func (f *FakeConfigMapsClient) Update(*api.ConfigMap) (*api.ConfigMap, error) {
	return nil, nil
}

func (f *FakeConfigMapsClient) Watch(api.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (s *OshinkoUnitTestSuite) TestNoLocalDefault(c *check.C) {
	// Test that if we ask for a named config "default" we do not
	// get an error if there is no local override of default.
	// For all other named configs, an error is returned if the local
	// definition is not found.
        var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	defconfig := clusterconfigs.GetDefaultConfig()
	configarg := clusterconfigs.ClusterConfig{Name: clusterconfigs.Defaultname}
	myconfig, err := clusterconfigs.GetClusterConfig(&configarg, cm)

	c.Assert(err, check.IsNil)
	c.Assert(myconfig.MasterCount, check.Equals, defconfig.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
}

func (s *OshinkoUnitTestSuite) TestDefaultConfig(c *check.C) {
	// Test that with no config object passed in, we get the default config
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}
	defconfig := clusterconfigs.GetDefaultConfig()
	myconfig, err := clusterconfigs.GetClusterConfig(nil, cm)
	c.Assert(myconfig.MasterCount, check.Equals, defconfig.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, defconfig.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, defconfig.SparkWorkerConfig)
	c.Assert(myconfig.Name, check.Equals, "default")
	c.Assert(err, check.IsNil)

	// Test that with a config object containing zeroes, we get the default config
	configarg := clusterconfigs.ClusterConfig{WorkerCount: 0, MasterCount: 0}
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MasterCount, check.Equals, defconfig.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, defconfig.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, defconfig.SparkWorkerConfig)
	c.Assert(myconfig.Name, check.Equals, "default")
	c.Assert(err, check.IsNil)
}

func (s *OshinkoUnitTestSuite) TestGetClusterConfigNamed(c *check.C) {
	// Test that named configs can inherit and override parts of the default config
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}
	defconfig := clusterconfigs.GetDefaultConfig()

	// configarg will represent a config object passed in a REST
	// request which specifies a named config but leaves counts unset
	configarg := clusterconfigs.ClusterConfig{WorkerCount: 0, MasterCount: 0}

	// tiny should inherit the default worker count
	cm.Create(makeConfigMap(tiny))
	configarg.Name = tiny.Name
	myconfig, err := clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MasterCount, check.Equals, tiny.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, defconfig.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, defconfig.SparkWorkerConfig)
	c.Assert(err, check.IsNil)

	// small supplies values for everything
	cm.Create(makeConfigMap(small))
	configarg.Name = small.Name
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MasterCount, check.Equals, small.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, small.WorkerCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, small.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, small.SparkWorkerConfig)
	c.Assert(err, check.IsNil)

	// large should inherit everything but the workercount
	cm.Create(makeConfigMap(large))
	configarg.Name = large.Name
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MasterCount, check.Equals, defconfig.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, large.WorkerCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, large.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, large.SparkWorkerConfig)
	c.Assert(err, check.IsNil)
}

func (s *OshinkoUnitTestSuite) TestGetClusterConfigNamedLinefeed(c *check.C) {
	// Depending on how a configmap is created in openshift, values may
	// have trailing linefeeds. Add linefeeds to the config values and
	// verify that ints are read correctly and strings do not contain
	// linefeeds. For our configs, string values should never legitimately
	// contain a linefeed.
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	// configarg will represent a config object passed in a REST
	// request which specifies a named config but leaves counts unset
	sm := makeConfigMap(small)
	addLineFeeds(sm)
	cm.Create(sm)
	configarg := clusterconfigs.ClusterConfig{WorkerCount: 0, MasterCount: 0}
	configarg.Name = small.Name

	myconfig, err := clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MasterCount, check.Equals, small.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, small.WorkerCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, small.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, small.SparkWorkerConfig)
	c.Assert(err, check.IsNil)

}

func (s *OshinkoUnitTestSuite) TestGetClusterConfigArgs(c *check.C) {
	// Test that a config object with no name but with args will
	// inherit and override defaults
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	defconfig := clusterconfigs.GetDefaultConfig()

	configarg := clusterconfigs.ClusterConfig{WorkerCount: 7, MasterCount: 0,
		SparkMasterConfig: "test-master-config", SparkWorkerConfig: "test-worker-config"}

	myconfig, err := clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MasterCount, check.Equals, defconfig.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, 7)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, "test-master-config")
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, "test-worker-config")
	c.Assert(err, check.IsNil)

	configarg = clusterconfigs.ClusterConfig{WorkerCount: 0, MasterCount: 7}
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MasterCount, check.Equals, 7)
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
	c.Assert(err, check.NotNil) // master count is illegal ...

	configarg = clusterconfigs.ClusterConfig{WorkerCount: 7, MasterCount: 7}
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MasterCount, check.Equals, 7)
	c.Assert(myconfig.WorkerCount, check.Equals, 7)
	c.Assert(err, check.NotNil) // master count is illegal ...
}

func (s *OshinkoUnitTestSuite) TestGetClusterConfigNamedArgs(c *check.C) {
	// Test that a named config with args will override and inherit
	// defaults, and that the args will take precedence
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	defconfig := clusterconfigs.GetDefaultConfig()

	cm.Create(makeConfigMap(brokenMaster))
	configarg := clusterconfigs.ClusterConfig{Name: brokenMaster.Name, WorkerCount: 7, MasterCount: 1}
	myconfig, err := clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MasterCount, check.Equals, 1)
	c.Assert(myconfig.WorkerCount, check.Equals, 7)
	c.Assert(brokenMaster.MasterCount, check.Not(check.Equals), int64(1))
	c.Assert(err, check.IsNil)

	configarg = clusterconfigs.ClusterConfig{Name: brokenMaster.Name, WorkerCount: 0, MasterCount: 5}
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MasterCount, check.Equals, 5)
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
	c.Assert(brokenMaster.MasterCount, check.Not(check.Equals), defconfig.WorkerCount)
	c.Assert(err, check.NotNil) // master count is wrong

	cm.Create(makeConfigMap(small))
	configarg = clusterconfigs.ClusterConfig{Name: small.Name, SparkMasterConfig: "test-master-config", SparkWorkerConfig: "test-worker-config"}
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, "test-master-config")
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, "test-worker-config")
	c.Assert(small.SparkMasterConfig, check.Not(check.Equals), "test-master-config")
	c.Assert(small.SparkWorkerConfig, check.Not(check.Equals), "test-worker-config")
	c.Assert(myconfig.MasterCount, check.Equals, small.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, small.WorkerCount)
	c.Assert(err, check.IsNil)
}

func (s *OshinkoUnitTestSuite) TestGetClusterBadConfig(c *check.C) {
	// Test that master count != 1 and worker count < 1 raises an error
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	defconfig := clusterconfigs.GetDefaultConfig()

	// configarg will represent a config object passed in a REST
	// request which specifies a named config but leaves counts unset
	configarg := clusterconfigs.ClusterConfig{WorkerCount: 0, MasterCount: 0}

	// brokenmaster should result in an error because the mastercount is != 1
	cm.Create(makeConfigMap(brokenMaster))
	configarg.Name = brokenMaster.Name
	myconfig, err := clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MasterCount, check.Equals, brokenMaster.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
	c.Assert(brokenMaster.MasterCount, check.Not(check.Equals), 1)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, clusterconfigs.MasterCountMustBeOne)

	// brokenworker should result in an error because the workercount is < 0
	cm.Create(makeConfigMap(brokenWorker))
	configarg.Name = brokenWorker.Name
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MasterCount, check.Equals, defconfig.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, brokenWorker.WorkerCount)
	w := brokenWorker.WorkerCount < 1
	c.Assert(w, check.Equals, true)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, clusterconfigs.WorkerCountMustBeAtLeastOne)
}

func (s *OshinkoUnitTestSuite) TestGetClusterNoConfig(c *check.C) {
	// Test that referencing a named config that doesn't exist fails
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	defconfig := clusterconfigs.GetDefaultConfig()
	configarg := clusterconfigs.ClusterConfig{WorkerCount: 0, MasterCount: 0, Name: "notthere"}

	myconfig, err := clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MasterCount, check.Equals, defconfig.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, fmt.Sprintf(clusterconfigs.NamedConfigDoesNotExist, "notthere"))
}

func (s *OshinkoUnitTestSuite) TestGetClusterNonInts(c *check.C) {
	// Test that master count and worker count must be ints
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	// configarg will represent a config object passed in a REST
	// request which specifies a named config but leaves counts unset
	configarg := clusterconfigs.ClusterConfig{WorkerCount: 0, MasterCount: 0}

	m := makeConfigMap(nonIntMaster)
	m.Data["mastercount"] = "fish"
	cm.Create(m)
	configarg.Name = nonIntMaster.Name
	_, err := clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals,
		fmt.Sprintf(clusterconfigs.ErrorWhileProcessing,
			configarg.Name + ".mastercount", "expected integer"))

	w := makeConfigMap(nonIntWorker)
	w.Data["workercount"] = "dog"
	cm.Create(w)
	configarg.Name = nonIntWorker.Name
	_, err = clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals,
		fmt.Sprintf(clusterconfigs.ErrorWhileProcessing,
			configarg.Name + ".workercount", "expected integer"))
}

func (s *OshinkoUnitTestSuite) TestGetClusterUserDefault(c *check.C) {
	// Test that defaults can be overridden optionally with a named
	// "default" config
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	defaultconfig := clusterconfigs.GetDefaultConfig()
	olddefault, err := clusterconfigs.GetClusterConfig(nil, cm)
	c.Assert(err, check.IsNil)
	c.Assert(defaultconfig, check.Equals, olddefault)

	cm.Create(makeConfigMap(userDefault))
	newdefault, err := clusterconfigs.GetClusterConfig(nil, cm)
	c.Assert(newdefault, check.Equals, userDefault)
}

func (s *OshinkoUnitTestSuite) TestGetClusterBadElements(c *check.C) {
	// Test that bogus config elements don't break anything
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	configarg := clusterconfigs.ClusterConfig{Name: small.Name, WorkerCount: 0, MasterCount: 0}

	sm := makeConfigMap(small)
	sm.Data["somethingelse"] = "chicken"
	cm.Create(sm)

	myconfig, err := clusterconfigs.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MasterCount, check.Equals, small.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, small.WorkerCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, small.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, small.SparkWorkerConfig)
	c.Assert(err, check.IsNil)
}
