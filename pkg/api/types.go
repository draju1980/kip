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

package api

import (
	"strings"

	uuid "github.com/satori/go.uuid"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	LatestAPIVersion = "v1"
)

// TypeMeta is metadata information for API objects.
type TypeMeta struct {
	// Kind is a string value for the resource this object represents.
	Kind string `json:"kind,omitempty"`
	// APIVersion defines the versioned schema of this representation of an
	// object.
	APIVersion string `json:"apiVersion,omitempty"`
}

func (meta *TypeMeta) Create() {
	meta.APIVersion = LatestAPIVersion
}

func (meta *TypeMeta) GetAPIVersion() string {
	return meta.APIVersion
}

type MilpaObject interface {
	// Implement this to be a MilpaObject
	IsMilpaObject()
}

// ObjectMeta is metadata that is maintained for all persisted resources, which
// includes all objects users create. This is added and kept up to date by
// Milpa.
type ObjectMeta struct {
	// Name of the resource.
	Name string `json:"name"`
	// A dictionary of labels applied to this resource..
	Labels map[string]string `json:"labels"`
	// Time of creation.
	CreationTimestamp Time `json:"creationTimestamp,omitempty"`
	// Time when the resource got deleted.
	DeletionTimestamp *Time `json:"deletionTimestamp,omitempty"`
	// Unused.
	Annotations map[string]string `json:"annotations,omitempty"`
	// Universal identifier in order to distinguish between different objects
	// that are named the same in differing timespans. E.g. if a user creates a
	// Pod named foo, then deletes and recreates the Pod, we need a way to tell
	// those two Pods apart.
	UID string `json:"uid,omitempty"`
	// Namespace placeholder. Currently Milpa does not support multiple
	// namespaces so this will always be set to "default".
	Namespace string `json:"namespace,omitempty"`
	// todo, other metadata parameters?
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	// see also vendor/k8s.io/api/core/v1 and
	// pkg/api/types.go

}

func (meta *ObjectMeta) Create() {
	meta.CreationTimestamp = Now()
	if meta.Name == "" {
		meta.Name = uuid.NewV4().String()
	}
	meta.UID = uuid.NewV4().String()
	meta.Namespace = "default"
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
}

func SetAPIVersion(version string) {

}

// Pod is a collection of Units that run on the same Node.
type Pod struct {
	// "squash" tag is used by mapstructure instead of inline
	TypeMeta `json:",inline,squash"`
	// Object metadata.
	ObjectMeta `json:"metadata"`
	// Spec is the desired behavior of the pod.
	Spec PodSpec `json:"spec,omitempty"`
	// Status is the observed status of the Pod. It is kept up to date by
	// Milpa.
	Status PodStatus `json:"status,omitempty"`
}

type PodSpec struct {
	// Desired condition of the Pod.
	Phase PodPhase `json:"phase"`
	// Restart policy for all Units in this Pod. It can be "always",
	// "onFailure" or "never". Default is "always". The restartPolicy
	// applies to all Units in the Pod. Exited Units are restarted
	// with an exponential back-off delay (10s, 20s, 40s …) capped at
	// five minutes, the delay is reset after 10 minutes.
	RestartPolicy RestartPolicy `json:"restartPolicy"`
	// List of Units that together compose this Pod.
	Units []Unit `json:"units"`
	// Init Units. They are run in order, one at a time before regular Units
	// are started.
	InitUnits []Unit `json:"initUnits"`
	// List of Secrets that will be used for authenticating when pulling
	// images.
	ImagePullSecrets []string `json:"imagePullSecrets,omitemtpy"`
	// Type of cloud instance type that will be used to run this Pod.
	InstanceType string `json:"instanceType,omitempty"`
	// PodSpot is the policy that determines if a spot instance may be used for
	// a Pod.
	Spot PodSpot `json:"spot,omitempty"`
	// Resource requirements for the Node that will run this Pod. If both
	// instanceType and resources are specified, instanceType will take
	// precedence.
	Resources ResourceSpec `json:"resources,omitempty"`
	// Placement is used to specify where a Pod will be place in the
	// infrastructure.
	Placement PlacementSpec `json:"placement,omitempty"`
	// List of volumes that will be made available to the Pod. Units can then
	// attach any of these mounts.
	Volumes []Volume `json:"volumes,omitempty"`
	// Pod security context.
	SecurityContext *PodSecurityContext `json:"securityContext,omitempty"`
	// Pod DNS policy.
	DNSPolicy DNSPolicy `json:"dnsPolicy,omitempty"`
	// Pod DNS config.
	DNSConfig *PodDNSConfig `json:"dnsConfig,omitempty"`
	// Specifies the hostname of the Pod
	// If not specified, the pod's hostname will be set to a system-defined value.
	// +optional
	Hostname string `json:"hostname,omitempty"`
	// If specified, the fully qualified Pod hostname will be "<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>".
	// If not specified, the pod will not have a domainname at all.
	// +optional
	Subdomain string `json:"subdomain,omitempty"`
	// HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts
	// file if specified. This is only valid for non-hostNetwork pods.
	// +optional
	// +patchMergeKey=ip
	// +patchStrategy=merge
	HostAliases []HostAlias `json:"hostAliases,omitempty"`
}

// HostAlias holds the mapping between IP and hostnames that will be injected as an entry in the
// pod's hosts file.
type HostAlias struct {
	// IP address of the host file entry.
	IP string `json:"ip,omitempty"`
	// Hostnames for the above IP address.
	Hostnames []string `json:"hostnames,omitempty"`
}

// DNSPolicy defines how a pod's DNS will be configured.
type DNSPolicy string

const (
	// DNSClusterFirstWithHostNet indicates that the pod should use cluster DNS
	// first, if it is available, then fall back on the default
	// (as determined by kubelet) DNS settings.
	DNSClusterFirstWithHostNet DNSPolicy = "ClusterFirstWithHostNet"

	// DNSClusterFirst indicates that the pod should use cluster DNS
	// first unless hostNetwork is true, if it is available, then
	// fall back on the default (as determined by kubelet) DNS settings.
	DNSClusterFirst DNSPolicy = "ClusterFirst"

	// DNSDefault indicates that the pod should use the default (as
	// determined by kubelet) DNS settings.
	DNSDefault DNSPolicy = "Default"

	// DNSNone indicates that the pod should use empty DNS settings. DNS
	// parameters such as nameservers and search paths should be defined via
	// DNSConfig.
	DNSNone DNSPolicy = "None"
)

// PodDNSConfig defines the DNS parameters of a pod in addition to
// those generated from DNSPolicy.
type PodDNSConfig struct {
	// A list of DNS name server IP addresses.
	// This will be appended to the base nameservers generated from DNSPolicy.
	// Duplicated nameservers will be removed.
	// +optional
	Nameservers []string `json:"nameservers,omitempty" protobuf:"bytes,1,rep,name=nameservers"`
	// A list of DNS search domains for host-name lookup.
	// This will be appended to the base search paths generated from DNSPolicy.
	// Duplicated search paths will be removed.
	// +optional
	Searches []string `json:"searches,omitempty" protobuf:"bytes,2,rep,name=searches"`
	// A list of DNS resolver options.
	// This will be merged with the base options generated from DNSPolicy.
	// Duplicated entries will be removed. Resolution options given in Options
	// will override those that appear in the base DNSPolicy.
	// +optional
	Options []PodDNSConfigOption `json:"options,omitempty" protobuf:"bytes,3,rep,name=options"`
}

// PodDNSConfigOption defines DNS resolver options of a pod.
type PodDNSConfigOption struct {
	// Required.
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// +optional
	Value *string `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
}

type PodSecurityContext struct {
	// PID, IPC and network namespace sharing options.
	NamespaceOptions *NamespaceOption `json:"namespaceOptions,omitempty"`
	// UID to run pod processes as.
	RunAsUser *int64 `json:"runAsUser,omitempty"`
	// GID to run pod processes as.
	RunAsGroup *int64 `json:"runAsGroup,omitempty"`
	// List of groups applied to the first process run in the sandbox, in
	// addition to the pod's primary GID.
	SupplementalGroups []int64 `json:"supplementalGroups,omitempty"`
	// Set these sysctls in the pod.
	Sysctls []Sysctl `json:"sysctls,omitempty"`
}

// NamespaceOption provides options for Linux namespaces.
type NamespaceOption struct {
	// Network namespace for this container/sandbox.
	// Note: There is currently no way to set CONTAINER scoped network in the Kubernetes API.
	// Namespaces currently set by the kubelet: POD, NODE
	Network NamespaceMode `json:"network,omitempty"`
	// PID namespace for this container/sandbox.
	// Note: The CRI default is POD, but the v1.PodSpec default is CONTAINER.
	// The kubelet's runtime manager will set this to CONTAINER explicitly for v1 pods.
	// Namespaces currently set by the kubelet: POD, CONTAINER, NODE
	Pid NamespaceMode `json:"pid,omitempty"`
	// IPC namespace for this container/sandbox.
	// Note: There is currently no way to set CONTAINER scoped IPC in the Kubernetes API.
	// Namespaces currently set by the kubelet: POD, NODE
	Ipc NamespaceMode `json:"ipc,omitempty"`
}

type NamespaceMode int32

const (
	// A POD namespace is common to all containers in a pod.
	// For example, a container with a PID namespace of POD expects to view
	// all of the processes in all of the containers in the pod.
	NamespaceModePod NamespaceMode = 0
	// A CONTAINER namespace is restricted to a single container.
	// For example, a container with a PID namespace of CONTAINER expects to
	// view only the processes in that container.
	NamespaceModeContainer NamespaceMode = 1
	// A NODE namespace is the namespace of the Kubernetes node.
	// For example, a container with a PID namespace of NODE expects to view
	// all of the processes on the host running the kubelet.
	NamespaceModeNode NamespaceMode = 2
)

// Sysctl defines a kernel parameter to be set.
type Sysctl struct {
	// Name of a property to set.
	Name string `json:"name"`
	// Value of a property to set.
	Value string `json:"value"`
}

// Definition for Volumes.
type Volume struct {
	// Name of the Volume. This is used when referencing a Volume from a Unit
	// definition.
	Name         string `json:"name"`
	VolumeSource `json:",inline,omitempty,squash"`
}

type VolumeSource struct {
	// If specified, an emptyDir will be created to back this Volume.
	EmptyDir *EmptyDir `json:"emptyDir,omitempty"`
	// This is a file or directory inside a package that will be mapped into
	// the rootfs of a Unit.
	PackagePath *PackagePath `json:"packagePath,omitempty"`
	// ConfigMap represents a configMap that should populate this volume
	ConfigMap *ConfigMapVolumeSource `json:"configMap,omitempty"`
	// Secret represents a secret that should populate this volume.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#secret
	// +optional
	Secret *SecretVolumeSource `json:"secret,omitempty"`
	// HostPath represents a pre-existing file or directory on the host
	// machine that is directly exposed to the container. This is generally
	// used for system agents or other privileged things that are allowed
	// to see the host machine. Most containers will NOT need this.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath
	// +optional
	HostPath *HostPathVolumeSource `json:"hostPath,omitempty"`
	// Items for all in one resources secrets, configmaps, and downward API
	Projected *ProjectedVolumeSource `json:"projected,omitempty"`
}

// Represents a host path mapped into a pod.
// Host path volumes do not support ownership management or SELinux relabeling.
type HostPathVolumeSource struct {
	// Path of the directory on the host.
	// If the path is a symlink, it will follow the link to the real path.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath
	Path string `json:"path" protobuf:"bytes,1,opt,name=path"`
	// Type for HostPath Volume
	// Defaults to ""
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath
	// +optional
	Type *HostPathType `json:"type,omitempty" protobuf:"bytes,2,opt,name=type"`
}

type HostPathType string

const (
	// For backwards compatible, leave it empty if unset
	HostPathUnset HostPathType = ""
	// If nothing exists at the given path, an empty directory will be created there
	// as needed with file mode 0755, having the same group and ownership with Kubelet.
	HostPathDirectoryOrCreate HostPathType = "DirectoryOrCreate"
	// A directory must exist at the given path
	HostPathDirectory HostPathType = "Directory"
	// If nothing exists at the given path, an empty file will be created there
	// as needed with file mode 0644, having the same group and ownership with Kubelet.
	HostPathFileOrCreate HostPathType = "FileOrCreate"
	// A file must exist at the given path
	HostPathFile HostPathType = "File"
	// A UNIX socket must exist at the given path
	HostPathSocket HostPathType = "Socket"
	// A character device must exist at the given path
	HostPathCharDev HostPathType = "CharDevice"
	// A block device must exist at the given path
	HostPathBlockDev HostPathType = "BlockDevice"
)

// Adapts a Secret into a volume.
//
// The contents of the target Secret's Data field will be presented in a volume
// as files using the keys in the Data field as the file names.
type SecretVolumeSource struct {
	// Name of the secret in the pod's namespace to use.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#secret
	// +optional
	SecretName string `json:"secretName,omitempty" protobuf:"bytes,1,opt,name=secretName"`
	// If unspecified, each key-value pair in the Data field of the referenced
	// Secret will be projected into the volume as a file whose name is the
	// key and content is the value. If specified, the listed keys will be
	// projected into the specified paths, and unlisted keys will not be
	// present. If a key is specified which is not present in the Secret,
	// the volume setup will error unless it is marked optional. Paths must be
	// relative and may not contain the '..' path or start with '..'.
	Items []KeyToPath `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
	// Optional: mode bits to use on created files by default. Must be a
	// value between 0 and 0777. Defaults to 0644.
	// Directories within the path are not affected by this setting.
	// This might be in conflict with other options that affect the file
	// mode, like fsGroup, and the result can be other mode bits set.
	DefaultMode *int32 `json:"defaultMode,omitempty"`
	// Specify whether the Secret or its keys must be defined
	Optional *bool `json:"optional,omitempty"`
}

// Adapts a ConfigMap into a volume.
//
// The contents of the target ConfigMap's Data field will be presented in a
// volume as files using the keys in the Data field as the file names, unless
// the items element is populated with specific mappings of keys to paths.
// ConfigMap volumes support ownership management and SELinux relabeling.
type ConfigMapVolumeSource struct {
	LocalObjectReference `json:",inline"`
	// If unspecified, each key-value pair in the Data field of the referenced
	// ConfigMap will be projected into the volume as a file whose name is the
	// key and content is the value. If specified, the listed keys will be
	// projected into the specified paths, and unlisted keys will not be
	// present. If a key is specified which is not present in the ConfigMap,
	// the volume setup will error unless it is marked optional. Paths must be
	// relative and may not contain the '..' path or start with '..'.
	Items []KeyToPath `json:"items,omitempty"`
	// Optional: mode bits to use on created files by default. Must be a
	// value between 0 and 0777. Defaults to 0644.
	// Directories within the path are not affected by this setting.
	// This might be in conflict with other options that affect the file
	// mode, like fsGroup, and the result can be other mode bits set.
	DefaultMode *int32 `json:"defaultMode,omitempty"`
	// Specify whether the ConfigMap or its keys must be defined
	Optional *bool `json:"optional,omitempty"`
}

// Maps a string key to a path within a volume.
type KeyToPath struct {
	// The key to project.
	Key string `json:"key" protobuf:"bytes,1,opt,name=key"`

	// The relative path of the file to map the key to.
	// May not be an absolute path.
	// May not contain the path element '..'.
	// May not start with the string '..'.
	Path string `json:"path"`
	// Optional: mode bits to use on this file, must be a value between 0
	// and 0777. If not specified, the volume defaultMode will be used.
	// This might be in conflict with other options that affect the file
	// mode, like fsGroup, and the result can be other mode bits set.
	Mode *int32 `json:"mode,omitempty"`
}

// Backing storage for Volumes.
type StorageMedium string

const (
	StorageMediumDefault StorageMedium = ""       // Use default (disk).
	StorageMediumMemory  StorageMedium = "Memory" // Use tmpfs.
	// Supporting huge pages will require some extra steps.
	//StorageMediumHugePages StorageMedium = "HugePages" // use hugepages
)

// EmptyDir is is disk or memory-backed Volume. Units can use it as
// scratch space, or for inter-unit communication (e.g. one Unit
// fetching files into an emptyDir, another running a webserver,
// serving these static files from the emptyDir).
type EmptyDir struct {
	// Backing medium for the emptyDir. The default is "" (to use disk
	// space).  The other option is "Memory", for creating a tmpfs
	// volume.
	Medium StorageMedium `json:"medium,omitempty"`
	// SizeLimit is only meaningful for tmpfs. It is the size of the tmpfs
	// volume.
	SizeLimit int64 `json:"sizeLimit,omitempty"`
}

// Source for a file or directory from a package that will be mapped into the
// rootfs of a Unit.
type PackagePath struct {
	// Path of the directory or file on the host.
	Path string `json:"path"`
}

// Represents a projected volume source
type ProjectedVolumeSource struct {
	// list of volume projections
	Sources []VolumeProjection `json:"sources"`
	// Mode bits to use on created files by default. Must be a value between
	// 0 and 0777.
	// Directories within the path are not affected by this setting.
	// This might be in conflict with other options that affect the file
	// mode, like fsGroup, and the result can be other mode bits set.
	// +optional
	DefaultMode *int32 `json:"defaultMode,omitempty"`
}

// Projection that may be projected along with other supported volume types
type VolumeProjection struct {
	// all types below are the supported types for projection into the same volume

	// information about the secret data to project
	// +optional
	Secret *SecretProjection `json:"secret,omitempty"`
	// // information about the downwardAPI data to project
	// // +optional
	// DownwardAPI *DownwardAPIProjection `json:"downwardAPI,omitempty"`
	// information about the configMap data to project
	// +optional
	ConfigMap *ConfigMapProjection `json:"configMap,omitempty"`
	// information about the serviceAccountToken data to project
	// +optional
	//ServiceAccountToken *ServiceAccountTokenProjection `json:"serviceAccountToken,omitempty"`
}

const (
	ProjectedVolumeSourceDefaultMode int32 = 0644
	SecretVolumeSourceDefaultMode    int32 = 0644
	ConfigMapVolumeSourceDefaultMode int32 = 0644
)

// Adapts a secret into a projected volume.
//
// The contents of the target Secret's Data field will be presented in a
// projected volume as files using the keys in the Data field as the file names.
// Note that this is identical to a secret volume source without the default
// mode.
type SecretProjection struct {
	LocalObjectReference `json:",inline" protobuf:"bytes,1,opt,name=localObjectReference"`
	// If unspecified, each key-value pair in the Data field of the referenced
	// Secret will be projected into the volume as a file whose name is the
	// key and content is the value. If specified, the listed keys will be
	// projected into the specified paths, and unlisted keys will not be
	// present. If a key is specified which is not present in the Secret,
	// the volume setup will error unless it is marked optional. Paths must be
	// relative and may not contain the '..' path or start with '..'.
	// +optional
	Items []KeyToPath `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
	// Specify whether the Secret or its key must be defined
	// +optional
	Optional *bool `json:"optional,omitempty" protobuf:"varint,4,opt,name=optional"`
}

// Adapts a ConfigMap into a projected volume.
//
// The contents of the target ConfigMap's Data field will be presented in a
// projected volume as files using the keys in the Data field as the file names,
// unless the items element is populated with specific mappings of keys to paths.
// Note that this is identical to a configmap volume source without the default
// mode.
type ConfigMapProjection struct {
	LocalObjectReference `json:",inline" protobuf:"bytes,1,opt,name=localObjectReference"`
	// If unspecified, each key-value pair in the Data field of the referenced
	// ConfigMap will be projected into the volume as a file whose name is the
	// key and content is the value. If specified, the listed keys will be
	// projected into the specified paths, and unlisted keys will not be
	// present. If a key is specified which is not present in the ConfigMap,
	// the volume setup will error unless it is marked optional. Paths must be
	// relative and may not contain the '..' path or start with '..'.
	// +optional
	Items []KeyToPath `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
	// Specify whether the ConfigMap or its keys must be defined
	// +optional
	Optional *bool `json:"optional,omitempty" protobuf:"varint,4,opt,name=optional"`
}

// // Represents downward API info for projecting into a projected volume.
// // Note that this is identical to a downwardAPI volume source without the default
// // mode.
// type DownwardAPIProjection struct {
// 	// Items is a list of DownwardAPIVolume file
// 	// +optional
// 	Items []DownwardAPIVolumeFile `json:"items,omitempty" protobuf:"bytes,1,rep,name=items"`
// }

// // DownwardAPIVolumeFile represents information to create the file containing the pod field
// type DownwardAPIVolumeFile struct {
// 	// Required: Path is  the relative path name of the file to be created. Must not be absolute or contain the '..' path. Must be utf-8 encoded. The first item of the relative path must not start with '..'
// 	Path string `json:"path" protobuf:"bytes,1,opt,name=path"`
// 	// Required: Selects a field of the pod: only annotations, labels, name and namespace are supported.
// 	// +optional
// 	FieldRef *ObjectFieldSelector `json:"fieldRef,omitempty" protobuf:"bytes,2,opt,name=fieldRef"`
// 	// Selects a resource of the container: only resources limits and requests
// 	// (limits.cpu, limits.memory, requests.cpu and requests.memory) are currently supported.
// 	// +optional
// 	ResourceFieldRef *ResourceFieldSelector `json:"resourceFieldRef,omitempty" protobuf:"bytes,3,opt,name=resourceFieldRef"`
// 	// Optional: mode bits to use on this file, must be a value between 0
// 	// and 0777. If not specified, the volume defaultMode will be used.
// 	// This might be in conflict with other options that affect the file
// 	// mode, like fsGroup, and the result can be other mode bits set.
// 	// +optional
// 	Mode *int32 `json:"mode,omitempty" protobuf:"varint,4,opt,name=mode"`
// }

// // ObjectFieldSelector selects an APIVersioned field of an object.
// type ObjectFieldSelector struct {
// 	// Version of the schema the FieldPath is written in terms of, defaults to "v1".
// 	// +optional
// 	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,1,opt,name=apiVersion"`
// 	// Path of the field to select in the specified API version.
// 	FieldPath string `json:"fieldPath" protobuf:"bytes,2,opt,name=fieldPath"`
// }

// // ResourceFieldSelector represents container resources (cpu, memory) and their output format
// type ResourceFieldSelector struct {
// 	// Container name: required for volumes, optional for env vars
// 	// +optional
// 	ContainerName string `json:"containerName,omitempty" protobuf:"bytes,1,opt,name=containerName"`
// 	// Required: resource to select
// 	Resource string `json:"resource" protobuf:"bytes,2,opt,name=resource"`
// 	// Specifies the output format of the exposed resources, defaults to "1"
// 	// +optional
// 	Divisor resource.Quantity `json:"divisor,omitempty" protobuf:"bytes,3,opt,name=divisor"`
// }

const (
	ContainerInstanceType = "ContainerInstance"
)

// ResourceSpec is used to specify resource requirements for the Node
// that will run a Pod.
type ResourceSpec struct {
	// The number of cpus on the instance.  Must be a string but can
	// be a fractional amount to accomodate shared cpu instance types
	// (e.g. 0.5)
	CPU string `json:"cpu,omitempty"`
	// The quantity of memory on the instance. Since this is a quantity
	// gigabytes should be expressed as "Gi".  E.G. memory: "3Gi"
	Memory string `json:"memory,omitempty"`
	// Number of GPUs present on the instance.
	GPU string `json:"gpu,omitempty"`
	// Root volume size. Both AWS and GCE specify volumes in GiB.
	// However according to their docs, AWS will bill you in
	// GB.
	VolumeSize string `json:"volumeSize,omitempty"`
	// Request an instance with dedicated or non-shared CPU. For AWS
	// T2 instances have a shared CPU, all other instance families
	// have a dedicated CPU.  Set dedicatedCPU to true if you do
	// not want Milpa to consider using a T2 instance for your Pod.
	DedicatedCPU bool `json:"dedicatedCPU,omitempty"`
	// Request unlimited CPU for T2 shared instance in AWS Only.
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/t2-unlimited.html
	SustainedCPU *bool `json:"sustainedCPU,omitempty"`
	// If PrivateIPOnly is true, the Pod will be launched on a Node
	// without a public IP address.  By default the Pod will run on
	// a Node with a public IP address.
	PrivateIPOnly bool `json:"privateIPOnly,omitempty"`
	// If ContainerInstance is true, the pod will be run as a cloud
	// container, in AWS, the pod will be run on Fargate{
	ContainerInstance *bool `json:"containerInstance,omitempty"`
}

// Units run applications. A Pod consists of one or more Units.
type Unit struct {
	// Name of the Unit.
	Name string `json:"name"`
	// The Docker image that will be pulled for this Unit. Usual Docker
	// conventions are used to specify an image, see
	// **[https://docs.docker.com/engine/reference/commandline/tag/#extended-description](https://docs.docker.com/engine/reference/commandline/tag/#extended-description)**
	// for a detailed explanation on specifying an image.
	//
	// Examples:
	//
	// - `library/python:3.6-alpine`
	//
	// - `myregistry.local:5000/testing/test-image`
	//
	Image string `json:"image,omitempty"`
	// The command that will be run to start the Unit. If empty, the entrypoint
	// of the image will be used. See
	// https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	Command []string `json:"command,omitempty"`
	// Arguments to the command. If empty, the cmd from the image will be used.
	Args []string `json:"args,omitempty"`
	// List of environment variables that will be exported inside the Unit
	// before start the application.
	Env []EnvVar `json:"env,omitempty"`
	// A list of Volumes that will be attached to the Unit.
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty"`
	// A list of ports that will be opened up for this Unit.
	Ports []ContainerPort `json:"ports,omitempty"`
	// Working directory to change to before running the command for the Unit.
	WorkingDir string `json:"workingDir,omitempty"`
	// Unit security context.
	SecurityContext *SecurityContext `json:"securityContext,omitempty"`
	// Periodic probe of container liveness.  Container will be
	// restarted if the probe fails.  Cannot be updated.  More info:
	// https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	LivenessProbe *Probe `json:"livenessProbe,omitempty"`
	// Periodic probe of container service readiness.  Container will
	// be removed from service endpoints if the probe fails.  Cannot
	// be updated.  More info:
	// https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	ReadinessProbe *Probe `json:"readinessProbe,omitempty"`
	//StartupProbe indicates that the Pod has successfully
	//initialized. If specified, no other probes are executed until
	//this completes successfully.
	StartupProbe *Probe `json:"startupProbe,omitempty"`
}

// Optional security context that overrides whatever is set for the pod.
//
// Example yaml:
//
// securityContext:
//           capabilities:
//             add:
//             - NET_BIND_SERVICE
//             drop:
//             - ALL
//
type SecurityContext struct {
	// Capabilities to add or drop.
	Capabilities *Capabilities `json:"capabilities,omitempty"`
	// UID to run unit processes as.
	RunAsUser *int64 `json:"runAsUser,omitempty"`
	// Username to run unit processes as.
	RunAsGroup *int64 `json:"runAsGroup,omitempty"`
}

// Capability contains the capabilities to add or drop.
type Capabilities struct {
	// List of capabilities to add.
	Add []string `json:"add,omitempty"`
	// List of capabilities to drop.
	Drop []string `json:"drop,omitempty"`
}

// ExecAction describes a "run in container" action.
type ExecAction struct {
	// Command is the command line to execute inside the container,
	// the working directory for the command is root ('/') in the
	// container's filesystem. The command is simply exec'd, it is not
	// run inside a shell, so traditional shell instructions ('|',
	// etc) won't work. To use a shell, you need to explicitly call
	// out to that shell.  Exit status of 0 is treated as live/healthy
	// and non-zero is unhealthy.
	Command []string `json:"command,omitempty"`
}

// URIScheme identifies the scheme used for connection to a host for Get actions
type URIScheme string

const (
	// URISchemeHTTP means that the scheme used will be http://
	URISchemeHTTP URIScheme = "HTTP"
	// URISchemeHTTPS means that the scheme used will be https://
	URISchemeHTTPS URIScheme = "HTTPS"
)

// HTTPHeader describes a custom header to be used in HTTP probes
type HTTPHeader struct {
	// The header field name
	Name string `json:"name"`
	// The header field value
	Value string `json:"value"`
}

// HTTPGetAction describes an action based on HTTP Get requests.
type HTTPGetAction struct {
	// Path to access on the HTTP server.
	Path string `json:"path,omitempty"`
	// Name or number of the port to access on the container.
	// Number must be in the range 1 to 65535.
	// Name must be an IANA_SVC_NAME.
	Port intstr.IntOrString `json:"port"`
	// Host name to connect to, defaults to the pod IP. You probably want to set
	// "Host" in httpHeaders instead.
	Host string `json:"host,omitempty"`
	// Scheme to use for connecting to the host.
	// Defaults to HTTP.
	Scheme URIScheme `json:"scheme,omitempty"`
	// Custom headers to set in the request. HTTP allows repeated headers.
	// +optional
	HTTPHeaders []HTTPHeader `json:"httpHeaders,omitempty"`
}

// TCPSocketAction describes an action based on opening a socket
type TCPSocketAction struct {
	// Number or name of the port to access on the container.
	// Number must be in the range 1 to 65535.
	// Name must be an IANA_SVC_NAME.
	Port intstr.IntOrString `json:"port"`
	// Optional: Host name to connect to, defaults to the pod IP.
	// +optional
	Host string `json:"host,omitempty"`
}

// Handler defines a specific action that should be taken
type Handler struct {
	// One and only one of the following should be specified.
	// Exec specifies the action to take.
	Exec *ExecAction `json:"exec,omitempty"`
	// HTTPGet specifies the http request to perform.
	HTTPGet *HTTPGetAction `json:"httpGet,omitempty"`
	// TCPSocket specifies an action involving a TCP port.
	// TCP hooks not yet supported
	TCPSocket *TCPSocketAction `json:"tcpSocket,omitempty"`
}

// Probe describes a health check to be performed against a container
// to determine whether it is alive or ready to receive traffic.
type Probe struct {
	// The action taken to determine the health of a container
	Handler `json:",inline"`
	// Number of seconds after the container has started before
	// liveness probes are initiated.  More info:
	// https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty"`
	// Number of seconds after which the probe times out.  Defaults to
	// 1 second. Minimum value is 1.  More info:
	// https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty"`
	// How often (in seconds) to perform the probe.  Default to 10
	// seconds. Minimum value is 1.
	PeriodSeconds int32 `json:"periodSeconds,omitempty"`
	// Minimum consecutive successes for the probe to be considered
	// successful after having failed.  Defaults to 1. Must be 1 for
	// liveness. Minimum value is 1.
	SuccessThreshold int32 `json:"successThreshold,omitempty"`
	// Minimum consecutive failures for the probe to be considered
	// failed after having succeeded.  Defaults to 3. Minimum value is
	// 1.
	FailureThreshold int32 `json:"failureThreshold,omitempty"`
}

// VolumeMount specifies what Volumes to attach to the Unit and the path where
// they will be located inside the Unit.
type VolumeMount struct {
	// Name of the Volume to attach.
	Name string `json:"name"`
	// Path where this Volume will be attached inside the Unit.
	MountPath string `json:"mountPath"`
}

// Environment variables.
type EnvVar struct {
	// Name of the environment variable.
	Name string `json:"name"`
	// Value of the environment variable.
	Value string `json:"value,omitempty"`
}

// LocalObjectReference contains enough information to let you locate the referenced object inside the same namespace.
type LocalObjectReference struct {
	//TODO: Add other useful fields.  apiVersion, kind, uid?
	Name string `json:"name,omitempty"`
}

// Selects a key from a ConfigMap.
type ConfigMapKeySelector struct {
	// The ConfigMap to select from.
	LocalObjectReference `json:",inline"`
	// The key to select.
	Key string `json:"key"`
	// Specify whether the ConfigMap or its key must be defined
	Optional *bool `json:"optional,omitempty"`
}

// SecretKeySelector selects a key of a Secret.
type SecretKeySelector struct {
	// The Secret to select from.
	LocalObjectReference
	// The key of the Secret to select from.  Must be a valid secret key.
	Key string `json:"key"`
	// Kubernetes allows optional Secrets.  We can add that soon
	Optional *bool `json:"optional,omitempty"`
}

// Spot policy. Can be "always", "preferred" or "never", meaning to always use
// a spot instance, use one when available, or never use a spot instance for
// running a Pod.
type SpotPolicy string

const (
	SpotAlways SpotPolicy = "Always"
	SpotNever  SpotPolicy = "Never"
)

// PodSpot is the policy that determines if a spot instance may be used for a
// Pod.
type PodSpot struct {
	// Spot policy. Can be "always", "preferred" or "never", meaning to always
	// use a spot instance, use one when available, or never use a spot
	// instance for running a Pod.
	Policy SpotPolicy `json:"policy"`
	// Notify string     `json:"notify"`
}

type NetworkAddressType string

const (
	PublicIP   NetworkAddressType = "PublicIP"
	PrivateIP  NetworkAddressType = "PrivateIP"
	PodIP      NetworkAddressType = "PodIP"
	PublicDNS  NetworkAddressType = "PublicDNS"
	PrivateDNS NetworkAddressType = "PrivateDNS"
)

type NetworkAddress struct {
	Type    NetworkAddressType `json:"type"`
	Address string             `json:"address"`
}

// Last observed status of the Pod. This is maintained by the system.
type PodStatus struct {
	// Phase is the last observed phase of the Pod. Can be "creating",
	// "dispatching", "running", "succeeded", "failed" or "terminated".
	Phase PodPhase `json:"phase"`
	// Time of the last phase change
	LastPhaseChange Time `json:"lastPhaseChange"`
	// Name of the node running this Pod.
	BoundNodeName string `json:"boundNodeName"`
	// ID of the node running this Pod.
	BoundInstanceID string `json:"boundInstanceID"`
	// IP addresses and DNS names of the Node running this Pod.
	Addresses []NetworkAddress `json:"addresses"`
	// Number of failures encountered while Milpa tried to start a Pod.
	StartFailures int `json:"startFailures"`
	// Shows the status of the Units on the Pod with one entry for
	// each Unit in the Pod's Spec.
	UnitStatuses []UnitStatus `json:"unitStatuses"`
	// Shows the status of the init Units on the Pod with one entry for each
	// init Unit in the Pod's Spec.
	InitUnitStatuses []UnitStatus `json:"initUnitStatuses"`
}

// Phase is the last observed phase of the Pod. Can be "creating",
// "dispatching", "running", "succeeded", "failed" or "terminated".
type PodPhase string

const (
	// PodWaiting means that we're waiting for the Pod to begin running.
	PodWaiting PodPhase = "Waiting"
	// PodDispatching means that we have a Node to put this Pod on
	// and we're in the process of starting the app on the Node.
	PodDispatching PodPhase = "Dispatching"
	// PodRunning means that the Pod is up and running.
	PodRunning PodPhase = "Running"
	// Pod succeeded means all the Units in the Pod returned success. It is a
	// terminal phase, i.e. the final phase when a Pod finished. Once the Pod
	// finished, Spec.Phase and Status.Phase are the same.
	PodSucceeded PodPhase = "Succeeded"
	// Pod has failed, either a Unit failed, or some other problem occurred
	// (e.g. dispatch error). This is a terminal phase.
	PodFailed PodPhase = "Failed"
	// PodTerminated means that the Pod has stopped by request. It is a
	// terminal phase.
	PodTerminated PodPhase = "Terminated"
)

func IsTerminalPodPhase(phase PodPhase) bool {
	switch phase {
	case PodTerminated, PodSucceeded, PodFailed:
		return true
	default:
		return false
	}
}

// Restart policy for all Units in this Pod. It can be "always", "onFailure" or
// "never". Default is "always".
type RestartPolicy string

const (
	RestartPolicyAlways    RestartPolicy = "Always"
	RestartPolicyOnFailure RestartPolicy = "OnFailure"
	RestartPolicyNever     RestartPolicy = "Never"
)

type PodList struct {
	TypeMeta `json:",inline"`
	Items    []*Pod `json:"items"`
}

// Node is a cloud instance that can run a Pod.
type Node struct {
	TypeMeta `json:",inline,squash"`
	// Object metadata.
	ObjectMeta `json:"metadata"`
	// Spec is the desired behavior of the Node.
	Spec NodeSpec `json:"spec"`
	// Status is the observed status of the Node. It is kept up to date by
	// Milpa.
	Status NodeStatus `json:"status"`
}

// NodeSpec defines the desired behavior of the Node.
type NodeSpec struct {
	// Cloud instance type of this Node.
	InstanceType string `json:"instanceType"`
	// Cloud image that is used for this instance.
	BootImage string `json:"bootImage"`
	// Indicates that this Node has been requested to be terminated.
	Terminate bool `json:"terminate,omitempty"`
	// This is a spot cloud instance.
	Spot bool `json:"spot"`
	// Resource requirements necessary for booting this Node. If both
	// instanceType and memory and cpu resources are specified,
	// instanceType will take precedence.  If the cloud provider
	// allows a variable number of CPUs/memory for an instance type,
	// the combination of resources and instance type will be used.
	Resources ResourceSpec `json:"resources,omitempty"`
	// Placement of the Node in the infrastructure.
	Placement PlacementSpec `json:"placement,omitempty"`
}

type PlacementSpec struct {
	AvailabilityZone string `json:"availabilityZone,omitempty"`

	// Future additions: In addition to explicitly specifying a subnet
	// we could make it so that users can use a selector to match
	// cloud tags on a subnet.
}

// NodeStatus is the last observed status of a Node.
type NodeStatus struct {
	// Phase is the last observed phase of the Node.
	Phase NodePhase `json:"phase"`
	// Cloud instance ID of this Node.
	InstanceID string `json:"instanceID"`
	// IP addresses and DNS names of this Node.
	Addresses []NetworkAddress `json:"addresses"`
	// If a Pod is bound to this Node, this is the name of that Pod.
	BoundPodName string `json:"boundPodName"`
}

// NodePhase is the last observed phase of the Node. Can be "creating",
// "created", "available", "claimed", "cleaning", "terminating" or
// "terminated".
type NodePhase string

const (
	NodeCreating    NodePhase = "Creating"
	NodeCreated     NodePhase = "Created"
	NodeAvailable   NodePhase = "Available"
	NodeClaimed     NodePhase = "Claimed"
	NodeCleaning    NodePhase = "Cleaning"
	NodeTerminating NodePhase = "Terminating"
	NodeTerminated  NodePhase = "Terminated"
)

type NodeList struct {
	TypeMeta `json:",inline"`
	Items    []*Node `json:"items"`
}

// ContainerPort represents a network port in a single container.
type ContainerPort struct {
	// If specified, this must be an IANA_SVC_NAME and unique within the pod. Each
	// named port in a pod must have a unique name. Name for the port that can be
	// referred to by services.
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// Number of port to expose on the host.
	// If specified, this must be a valid port number, 0 < x < 65536.
	// If HostNetwork is specified, this must match ContainerPort.
	// Most containers do not need this.
	// +optional
	HostPort int32 `json:"hostPort,omitempty" protobuf:"varint,2,opt,name=hostPort"`
	// Number of port to expose on the pod's IP address.
	// This must be a valid port number, 0 < x < 65536.
	ContainerPort int32 `json:"containerPort" protobuf:"varint,3,opt,name=containerPort"`
	// Protocol for port. Must be UDP, TCP, or SCTP.
	// Defaults to "TCP".
	// +optional
	Protocol Protocol `json:"protocol,omitempty" protobuf:"bytes,4,opt,name=protocol,casttype=Protocol"`
	// What host IP to bind the external port to.
	// +optional
	HostIP string `json:"hostIP,omitempty" protobuf:"bytes,5,opt,name=hostIP"`
}

// Protocol defines network protocols supported for things like ports.
type Protocol string

func MakeProtocol(p string) Protocol {
	return Protocol(strings.ToUpper(p))
}

const (
	ProtocolTCP  Protocol = "TCP"
	ProtocolUDP  Protocol = "UDP"
	ProtocolSCTP Protocol = "SCTP"
	ProtocolICMP Protocol = "ICMP"
)

// ServiceStatus represents the current status of a Service.
type ServiceStatus struct {
	// LoadBalancer contains the current status of the load-balancer,
	// if one is present.
	LoadBalancer LoadBalancerStatus `json:"loadBalancer,omitempty"`
}

// LoadBalancerStatus represents the status of a load-balancer.
type LoadBalancerStatus struct {
	// Ingress is a list containing ingress points for the load-balancer;
	// traffic intended for the Service should be sent to these ingress points.
	Ingress []LoadBalancerIngress `json:"ingress,omitempty"`
}

// LoadBalancerIngress represents the status of a load-balancer ingress point
// traffic intended for the Service should be sent to an ingress point.
type LoadBalancerIngress struct {
	// IP is set for load-balancer ingress points that are IP based
	// (typically GCE or OpenStack load-balancers)
	// +optional
	IP string `json:"ip,omitempty"`
	// Hostname is set for load-balancer ingress points that are DNS
	// based such as AWS load-balancers.
	Hostname string `json:"hostname,omitempty"`
}

// PodTemplateSpec is the object that describes the Pod that will be created if
// insufficient replicas are detected.
type PodTemplateSpec struct {
	// Object metadata.
	ObjectMeta `json:"metadata"`
	// Spec defines the behavior of a Pod.
	Spec PodSpec `json:"spec,omitempty"`
}

type StorageType string

const (
	StorageGP2             StorageType = "gp2"
	StorageStandardSSD     StorageType = "StandardSSD"
	StandardPersistentDisk StorageType = "StandardPersistentDisk"
)

// There are two different styles of label selectors used in versioned types:
// an older style which is represented as just a string in versioned types, and
// a newer style that is structured. LabelSelector is an internal
// representation for the latter style. A label selector is a label query over
// a set of resources. The result of matchLabels and matchExpressions are
// ANDed. An empty label selector matches all objects. A null label selector
// matches no objects.
type LabelSelector struct {
	// matchLabels is a map of {key,value} pairs. A single {key,value} in the
	// matchLabels map is equivalent to an element of matchExpressions, whose
	// key field is "key", the operator is "In", and the values array contains
	// only "value". The requirements are ANDed.
	MatchLabels map[string]string `json:"matchLabels,omitempty" protobuf:"bytes,1,rep,name=matchLabels"`
	// matchExpressions is a list of label selector requirements. The
	// requirements are ANDed.
	MatchExpressions []LabelSelectorRequirement `json:"matchExpressions,omitempty" protobuf:"bytes,2,rep,name=matchExpressions"`
}

// A label selector requirement is a selector that contains values, a key, and
// an operator that relates the key and values.
type LabelSelectorRequirement struct {
	// key is the label key that the selector applies to.
	Key string `json:"key" patchStrategy:"merge" patchMergeKey:"key" protobuf:"bytes,1,opt,name=key"`
	// operator represents a key's relationship to a set of values.  Valid
	// operators ard In, NotIn, Exists and DoesNotExist.
	Operator LabelSelectorOperator `json:"operator" protobuf:"bytes,2,opt,name=operator,casttype=LabelSelectorOperator"`
	// values is an array of string values. If the operator is In or NotIn, the
	// values array must be non-empty. If the operator is Exists or
	// DoesNotExist, the values array must be empty. This array is replaced
	// during a strategic merge patch.
	Values []string `json:"values,omitempty" protobuf:"bytes,3,rep,name=values"`
}

// A label selector operator is the set of operators that can be used in a
// selector requirement. Can be "in", "notIn", "exists" and "doesNotExist".
type LabelSelectorOperator string

const (
	LabelSelectorOpIn           LabelSelectorOperator = "In"
	LabelSelectorOpNotIn        LabelSelectorOperator = "NotIn"
	LabelSelectorOpExists       LabelSelectorOperator = "Exists"
	LabelSelectorOpDoesNotExist LabelSelectorOperator = "DoesNotExist"
)

func (p Pod) IsMilpaObject()         {}
func (p PodList) IsMilpaObject()     {}
func (p Node) IsMilpaObject()        {}
func (p NodeList) IsMilpaObject()    {}
func (p Event) IsMilpaObject()       {}
func (p EventList) IsMilpaObject()   {}
func (p LogFile) IsMilpaObject()     {}
func (p LogFileList) IsMilpaObject() {}
func (p Metrics) IsMilpaObject()     {}
func (p MetricsList) IsMilpaObject() {}

// ObjectReference contains enough information to be able to retrieve the
// object from the registry.
type ObjectReference struct {
	Kind string `json:"kind,omitempty"`
	Name string `json:"name,omitempty"`
	UID  string `json:"uid,omitempty"`
}

// Event is a report of an event that happened in Milpa. They are stored
// separately from the objects they apply to.
type Event struct {
	TypeMeta `json:",inline,squash"`

	ObjectMeta `json:"metadata"`

	// The object that this event is about.
	InvolvedObject ObjectReference `json:"involvedObject"`

	// Should be a short, machine understandable string that describes the
	// current status of the referred object. This should not give the reason
	// for being in this state.  Examples: "running", "cantStart",
	// "cantSchedule", "deleted".  It's OK for components to make up statuses
	// to report here, but the same string should always be used for the same
	// status.
	Status string `json:"status,omitempty"`

	// The component reporting this Event. Should be a short machine
	// understandable string.
	Source string `json:"source,omitempty"`

	// Human readable message about what happened.
	Message string `json:"message,omitempty"`
}

// A list of Events.
type EventList struct {
	TypeMeta `json:",inline"`
	Items    []*Event `json:"items"`
}

// LogFile holds the log data created by a Pod Unit or a Node.
type LogFile struct {
	TypeMeta `json:",inline,squash"`

	ObjectMeta `json:"metadata"`

	// The object that created this log.
	ParentObject ObjectReference `json:"parentObject,omitempty"`

	// The content of the logfile. If the logfile is long, this will
	// likely be the tail of the file.
	Content string `json:"Content,omitempty"`
}

// A list of logfiles.
type LogFileList struct {
	TypeMeta `json:",inline"`
	Items    []*LogFile `json:"items"`
}

type UnitStateWaiting struct {
	Reason       string `json:"reason,omitempty"`
	StartFailure bool   `json:"startFailure,omitempty"`
}

type UnitStateRunning struct {
	StartedAt Time `json:"startedAt,omitempty"`
}

type UnitStateTerminated struct {
	ExitCode   int32  `json:"exitCode"`
	FinishedAt Time   `json:"finishedAt,omitempty"`
	Reason     string `json:"reason,omitempty"`
	Message    string `json:"message,omitempty"`
	StartedAt  Time   `json:"startedAt,omitempty"`
}

// UnitState holds a possible state of a Pod Unit.  Only one of its
// members may be specified.  If none of them is specified, the
// default one is UnitStateRunning.
type UnitState struct {
	Waiting    *UnitStateWaiting    `json:"waiting,omitempty"`
	Running    *UnitStateRunning    `json:"running,omitempty"`
	Terminated *UnitStateTerminated `json:"terminated,omitempty"`
}

type UnitStatus struct {
	Name                 string    `json:"name"`
	State                UnitState `json:"state,omitempty"`
	LastTerminationState UnitState `json:"lastState,omitempty"`
	RestartCount         int32     `json:"restartCount"`
	Image                string    `json:"image"`
	Ready                bool      `json:"ready"`
	Started              *bool     `json:"started"`
}

type Metrics struct {
	TypeMeta   `json:",inline,squash"`
	ObjectMeta `json:"metadata"`

	// The time at the end of the metrics collection window.
	Timestamp Time `json:"timestamp,omitempty"`

	// The interval of time over which the metrics were collected:
	// [Timestamp-Window, Timestamp]
	Window Duration `json:"window,omitempty"`

	// A map of lower case metric names to metric values
	ResourceUsage ResourceMetrics `json:"resourceUsage,omitempty"`
}

type ResourceMetrics map[string]float64

type MetricsList struct {
	TypeMeta `json:",inline"`
	Items    []*Metrics
}
