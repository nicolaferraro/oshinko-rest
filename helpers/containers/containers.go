package containers

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
)

type OContainer struct {
	kapi.Container
}

func Container(name, image string) *OContainer {
	c := OContainer{}
	c.Name = name
	c.Image = image
	c.ImagePullPolicy = kapi.PullIfNotPresent
	c.TerminationMessagePath = "/dev/termination-log"
	return &c
}

func (c *OContainer) Command(args ...string) *OContainer {
	c.Container.Command = args
	return c
}

func (c *OContainer) EnvVar(name, value string) *OContainer {
	c.Container.Env = append(c.Container.Env, kapi.EnvVar{Name: name, Value: value})
	return c
}

func (c *OContainer) EnvVars(envs ...kapi.EnvVar) *OContainer {
	c.Container.Env = envs
	return c
}

// TODO we might want to add some handling around building Quantities too
func (pt *OContainer) ResourceLimit(name kapi.ResourceName, q resource.Quantity) *OContainer {
	if pt.Resources.Limits == nil {
		pt.Resources.Limits = make(kapi.ResourceList, 1)
	}
	pt.Resources.Limits[name] = q
	return pt
}

func (pt *OContainer) ResourceRequest(name kapi.ResourceName, q resource.Quantity) *OContainer {
	if pt.Resources.Requests == nil {
		pt.Resources.Requests = make(kapi.ResourceList, 1)
	}
	pt.Resources.Requests[name] = q
	return pt
}

func (c *OContainer) Ports(ports ...*OContainerPort) *OContainer {
	kports := make([]kapi.ContainerPort, len(ports))
	for idx, p := range ports {
		kports[idx] = p.ContainerPort
	}
	c.Container.Ports = kports
	return c
}

type OContainerPort struct {
	kapi.ContainerPort
}


func ContainerPort(name string, port int) *OContainerPort {
	cp := OContainerPort{}
	cp.Name = name
	cp.ContainerPort.ContainerPort = port
	cp.ContainerPort.Protocol = kapi.ProtocolTCP
	return &cp
}

func (cp *OContainerPort) Protocol(proto kapi.Protocol) *OContainerPort {
	cp.ContainerPort.Protocol = proto
	return cp
}

func (cp *OContainerPort) SetName(name string) *OContainerPort {
	cp.Name = name
	return cp
}

func (cp *OContainerPort) HostPort(port int) *OContainerPort {
	cp.ContainerPort.HostPort = port
	return cp
}

func (cp *OContainerPort) HostIP(hostip string) *OContainerPort {
	cp.ContainerPort.HostIP = hostip
	return cp
}