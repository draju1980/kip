/*
Copyright 2020 Elotl Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloud

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/elotl/cloud-instance-provider/pkg/api"
	"github.com/elotl/cloud-instance-provider/pkg/util/sets"

	"k8s.io/klog"
)

const MilpaAPISGName = "CellSecurityGroup"
const PublicCIDR = "0.0.0.0/0"
const RestAPIPort = 6421

const ProviderAWS = "aws"
const ProviderGCE = "gce"
const ProviderAzure = "azure"

const ControllerTagKey = "KipControllerID"
const NameTagKey = "Name"
const NamespaceTagKey = "KipNamespace"
const NametagTagKey = "KipNametag"
const PodNameTagKey = "KipPodName"

type CloudClient interface {
	SetBootSecurityGroupIDs([]string)
	GetBootSecurityGroupIDs() []string
	StartNode(*api.Node, string) (*StartNodeResult, error)
	StartSpotNode(*api.Node, string) (*StartNodeResult, error)
	// This should always be called from a goroutine as it can take a while
	StopInstance(instanceID string) error
	WaitForRunning(node *api.Node) ([]api.NetworkAddress, error)
	EnsureMilpaSecurityGroups([]string, []string) error
	AttachSecurityGroups(node *api.Node, groups []string) error
	AssignInstanceProfile(node *api.Node, instanceProfile string) error
	ListInstancesFilterID([]string) ([]CloudInstance, error)
	ListInstances() ([]CloudInstance, error)
	ResizeVolume(node *api.Node, size int64) (error, bool)
	GetRegistryAuth() (string, string, error)
	// Todo, correct capitalization on this one
	GetImageId(tags BootImageTags) (string, error)
	SetSustainedCPU(*api.Node, bool) error
	AddInstanceTags(string, map[string]string) error
	ControllerInsideVPC() bool
	ModifySourceDestinationCheck(string, bool) error
	RemoveRoute(string) error
	AddRoute(string, string) error
	GetVPCCIDRs() []string
	// Address spaces used by cloud-internal services that might initiate
	// connections to instances in the VPC.
	CloudStatusKeeper() StatusKeeper
	GetSubnets() ([]SubnetAttributes, error)
	GetAvailabilityZones() ([]string, error)
	GetAttributes() CloudAttributes
}

type CloudAttributes struct {
	DiskProductName           api.StorageType
	FixedSizeVolume           bool
	MaxInstanceSecurityGroups int
	Provider                  string
	Region                    string
}

type StartNodeResult struct {
	InstanceID       string
	AvailabilityZone string
}

type SubnetAddressAffinity string

const (
	PublicAddress  SubnetAddressAffinity = "Public"
	PrivateAddress SubnetAddressAffinity = "Private"
	AnyAddress     SubnetAddressAffinity = "Any"
)

type SubnetAttributes struct {
	Name string
	ID   string
	CIDR string
	AZ   string
	// In AWS subnets we use the subnets private/public address by default
	// flag to decide where to launch public and private nodes.  We store
	// that info in AddressAffinity.  In Azure, it's likely we don't have
	// that type of affinity (there's some interesting NAT options in azure)
	// so we don't really care what subnet our public and private addresses
	// go in.  Also, this is half baked so if you have an idea of a better
	// way to specify placement, knock yourself out.
	AddressAffinity SubnetAddressAffinity
	// In AWS and Azure (pretty sure...), we can get availability
	// stats However, they're harder to come by in GCP.  That said, in
	// GCP you can resize your subnets and we can always query
	// instances and bucket them.
	AvailableAddresses int
	//Capacity            int
}

type Image struct {
	Id   string
	Name string
}

type BootImageTags struct {
	Company string `json:"company"`
	Product string `json:"product"`
	Version string `json:"version"`
	Date    string `json:"date"`
	Time    string `json:"time"`
}

func (bit *BootImageTags) Timestamp() (time.Time, error) {
	s := fmt.Sprintf("%s %s", bit.Date, bit.Time)
	return time.Parse("20060102 150405", s)
}

func FilterImages(images []Image, tags BootImageTags) []Image {
	result := make([]Image, 0)
	for _, img := range images {
		t := BootImageTags{}
		t.Set(img.Name)
		if t.Matches(tags) {
			klog.V(4).Infof("Found image %s matching filter %+v", img.Name, tags)
			result = append(result, img)
		}
	}
	return result
}

func SortImages(images []Image) {
	sort.Slice(images, func(i, j int) bool {
		// For really old images, the creation timestamp might be empty. Use
		// epoch zero in that case.
		bitI := BootImageTags{}
		bitI.Set(images[i].Name)
		versionI, err := strconv.ParseUint(bitI.Version, 10, 32)
		if err != nil {
			klog.Warningf("Getting version for image %+v: %v", bitI, err)
		}
		dateI, err := bitI.Timestamp()
		if err != nil {
			klog.Warningf("Getting timestamp for image %+v: %v", bitI, err)
			dateI = time.Unix(0, 0)
		}
		bitJ := BootImageTags{}
		bitJ.Set(images[j].Name)
		versionJ, err := strconv.ParseUint(bitJ.Version, 10, 32)
		if err != nil {
			klog.Warningf("Getting version for image %+v: %v", bitI, err)
		}
		dateJ, err := bitJ.Timestamp()
		if err != nil {
			klog.Warningf("Getting timestamp for image %+v: %v", bitI, err)
			dateJ = time.Unix(0, 0)
		}
		if versionI != versionJ {
			return versionI < versionJ
		}
		return dateI.Before(dateJ)
	})
}

func GetBestImage(images []Image, tags BootImageTags) (string, error) {
	images = FilterImages(images, tags)
	SortImages(images)
	if len(images) == 0 {
		err := fmt.Errorf("No image matching tags %+v found", tags)
		return "", err
	}
	latest := images[len(images)-1].Id
	klog.V(2).Infof("Found image %s for tags %v", latest, tags)
	return latest, nil
}

func (bit *BootImageTags) String() string {
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		bit.Company, bit.Product, bit.Version, bit.Date, bit.Time)
}

func (bit *BootImageTags) Set(s string) {
	tags := strings.Split(s, "-")
	if len(tags) > 0 {
		bit.Company = tags[0]
	}
	if len(tags) > 1 {
		bit.Product = tags[1]
	}
	if len(tags) > 2 {
		bit.Version = tags[2]
	}
	if len(tags) > 3 {
		bit.Date = tags[3]
	}
	if len(tags) > 4 {
		bit.Time = tags[4]
	}
}

func (bit *BootImageTags) Matches(input BootImageTags) bool {
	if input.Company != "" && bit.Company != input.Company {
		return false
	}
	if input.Product != "" && bit.Product != input.Product {
		return false
	}
	if input.Version != "" && bit.Version != input.Version {
		return false
	}
	if input.Date != "" && bit.Date != input.Date {
		return false
	}
	if input.Time != "" && bit.Time != input.Time {
		return false
	}
	return true
}

type CloudInstance struct {
	ID       string
	NodeName string
}

type ContainerInstance struct {
	ID string
}

// List instances only gives us security identifier
type SecurityGroupIdentifier struct {
	ID   string
	Name string
}

type SecurityGroup struct {
	ID           string
	Name         string
	Ports        []InstancePort
	SourceRanges []string
}

func NewSecurityGroup(id, name string, ports []InstancePort, sources []string) SecurityGroup {
	sort.Sort(SortableSliceOfPorts(ports))
	sort.Strings(sources)
	return SecurityGroup{
		ID:           id,
		Name:         name,
		Ports:        ports,
		SourceRanges: sources,
	}
}

type LoadBalancer struct {
	Type             string
	ServiceName      string
	LoadBalancerName string
	Instances        sets.String
	Ports            []InstancePort
	SecurityGroupID  string
	Internal         bool
	Annotations      map[string]string
	DNSName          string
	IPAddress        string
}
