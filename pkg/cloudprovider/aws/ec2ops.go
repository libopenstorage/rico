/*
Package aws implements the cloud interface for AWS
Copyright 2018 Portworx

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
package aws

// NOTE:
// THIS FILE IS STILL WORK-IN-PROGRESS
// It will be fixed for rico, then combined with storageops
// TBD where it will live once combined

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/libopenstorage/openstorage/pkg/storageops"
	"github.com/portworx/sched-ops/task"
	"go.pedge.io/dlog"
)

type Ec2Ops struct {
	ec2   *ec2.EC2
	mutex sync.Mutex
}

// ErrAWSEnvNotAvailable is the error type when aws credentails are not set
var ErrAWSEnvNotAvailable = fmt.Errorf("AWS credentials are not set in environment")

// NewEc2Ops creates a new aws storage ops instance
func NewEc2Ops(ec2 *ec2.EC2) *Ec2Ops {
	return &Ec2Ops{
		ec2: ec2,
	}
}

func (s *Ec2Ops) filters(
	labels map[string]string,
	keys []string,
) []*ec2.Filter {
	if len(labels) == 0 {
		return nil
	}
	f := make([]*ec2.Filter, len(labels)+len(keys))
	i := 0
	for k, v := range labels {
		s := string("tag:") + k
		value := v
		f[i] = &ec2.Filter{Name: &s, Values: []*string{&value}}
		i++
	}
	for _, k := range keys {
		s := string("tag-key:") + k
		f[i] = &ec2.Filter{Name: &s}
		i++
	}
	return f
}

func (s *Ec2Ops) tags(labels map[string]string) []*ec2.Tag {
	if len(labels) == 0 {
		return nil
	}
	t := make([]*ec2.Tag, len(labels))
	i := 0
	for k, v := range labels {
		key := k
		value := v
		t[i] = &ec2.Tag{Key: &key, Value: &value}
		i++
	}
	return t
}

func (s *Ec2Ops) waitStatus(id string, desired string) error {
	request := &ec2.DescribeVolumesInput{VolumeIds: []*string{&id}}
	actual := ""

	_, err := task.DoRetryWithTimeout(
		func() (interface{}, bool, error) {
			awsVols, err := s.ec2.DescribeVolumes(request)
			if err != nil {
				return nil, true, err
			}

			if len(awsVols.Volumes) != 1 {
				return nil, true, fmt.Errorf("expected one volume %v got %v",
					id, len(awsVols.Volumes))
			}

			if awsVols.Volumes[0].State == nil {
				return nil, true, fmt.Errorf("Nil volume state for %v", id)
			}

			actual = *awsVols.Volumes[0].State
			if actual == desired {
				return nil, false, nil
			}

			return nil, true, fmt.Errorf(
				"Volume %v did not transition to %v current state %v",
				id, desired, actual)

		},
		storageops.ProviderOpsTimeout,
		storageops.ProviderOpsRetryInterval)

	return err

}

func (s *Ec2Ops) waitAttachmentStatus(
	volumeID string,
	desired string,
	timeout time.Duration,
) (*ec2.Volume, error) {
	id := volumeID
	request := &ec2.DescribeVolumesInput{VolumeIds: []*string{&id}}
	actual := ""
	interval := 2 * time.Second

	var outVol *ec2.Volume
	for elapsed := 0 * time.Second; actual != desired && elapsed < timeout; elapsed += interval {
		awsVols, err := s.ec2.DescribeVolumes(request)
		if err != nil {
			return nil, err
		}
		if len(awsVols.Volumes) != 1 {
			return nil, fmt.Errorf("expected one volume %v got %v",
				volumeID, len(awsVols.Volumes))
		}
		outVol = awsVols.Volumes[0]
		awsAttachment := awsVols.Volumes[0].Attachments
		if awsAttachment == nil || len(awsAttachment) == 0 {
			// We have encountered scenarios where AWS returns a nil attachment state
			// for a volume transitioning from detaching -> attaching.
			actual = ec2.VolumeAttachmentStateDetached
		} else {
			actual = *awsAttachment[0].State
		}
		if actual == desired {
			break
		}
		time.Sleep(interval)
	}
	if actual != desired {
		return nil, fmt.Errorf("Volume %v failed to transition to  %v current state %v",
			volumeID, desired, actual)
	}
	return outVol, nil
}

func (s *Ec2Ops) Name() string { return "aws" }

func (s *Ec2Ops) ApplyTags(volumeID string, labels map[string]string) error {
	req := &ec2.CreateTagsInput{
		Resources: []*string{&volumeID},
		Tags:      s.tags(labels),
	}
	_, err := s.ec2.CreateTags(req)
	return err
}

func (s *Ec2Ops) RemoveTags(volumeID string, labels map[string]string) error {
	req := &ec2.DeleteTagsInput{
		Resources: []*string{&volumeID},
		Tags:      s.tags(labels),
	}
	_, err := s.ec2.DeleteTags(req)
	return err
}

func (s *Ec2Ops) matchTag(tag *ec2.Tag, match string) bool {
	return tag.Key != nil &&
		tag.Value != nil &&
		len(*tag.Key) != 0 &&
		len(*tag.Value) != 0 &&
		*tag.Key == match
}

func (s *Ec2Ops) DeviceMappings(instanceID string) (map[string]string, error) {
	instance, err := s.DescribeInstance(instanceID)
	if err != nil {
		return nil, err
	}
	devPrefix := "/dev/sd"
	m := make(map[string]string)
	for _, d := range instance.BlockDeviceMappings {
		if d.DeviceName != nil && d.Ebs != nil && d.Ebs.VolumeId != nil {
			devName := *d.DeviceName
			// Skip the root device
			if devName == *instance.RootDeviceName {
				continue
			}
			// AWS EBS volumes get mapped from /dev/sdN -->/dev/xvdN
			if strings.HasPrefix(devName, devPrefix) {
				devName = "/dev/xvd" + devName[len(devPrefix):]
			}
			m[devName] = *d.Ebs.VolumeId
		}
	}
	return m, nil
}

// DescribeInstance describes a single specific instance.
func (s *Ec2Ops) DescribeInstance(instanceID string) (*ec2.Instance, error) {
	request := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{&instanceID},
	}
	out, err := s.ec2.DescribeInstances(request)
	if err != nil {
		return nil, err
	}
	if len(out.Reservations) != 1 {
		return nil, fmt.Errorf("DescribeInstances(%v) returned %v reservations, expect 1",
			instanceID, len(out.Reservations))
	}
	if len(out.Reservations[0].Instances) != 1 {
		return nil, fmt.Errorf("DescribeInstances(%v) returned %v Reservations, expect 1",
			instanceID, len(out.Reservations[0].Instances))
	}
	return out.Reservations[0].Instances[0], nil
}

func (s *Ec2Ops) FreeDevices(
	blockDeviceMappings []interface{},
	rootDeviceName string,
) ([]string, error) {
	initial := []byte("fghijklmnop")
	devPrefix := "/dev/sd"
	for _, d := range blockDeviceMappings {
		dev := d.(*ec2.InstanceBlockDeviceMapping)

		if dev.DeviceName == nil {
			return nil, fmt.Errorf("Nil device name")
		}
		devName := *dev.DeviceName

		if devName == rootDeviceName {
			continue
		}
		if !strings.HasPrefix(devName, devPrefix) {
			devPrefix = "/dev/xvd"
			if !strings.HasPrefix(devName, devPrefix) {
				return nil, fmt.Errorf("bad device name %q", devName)
			}
		}
		letter := devName[len(devPrefix):]
		// Reset devPrefix for next devices
		devPrefix = "/dev/sd"

		// AWS instances can have the following device names
		// /dev/xvd[b-c][a-z]
		if len(letter) == 1 {
			index := letter[0] - 'f'
			if index > ('p' - 'f') {
				continue
			}
			initial[index] = '0'
		} else if len(letter) == 2 {
			// We do not attach EBS volumes with "/dev/xvdc[a-z]" formats
			continue
		} else {
			return nil, fmt.Errorf("cannot parse device name %q", devName)
		}
	}
	free := make([]string, len(initial))
	count := 0
	for _, b := range initial {
		if b != '0' {
			free[count] = devPrefix + string(b)
			count++
		}
	}
	if count == 0 {
		return nil, fmt.Errorf("No more free devices")
	}
	return free[:count], nil
}

func (s *Ec2Ops) rollbackCreate(id string, createErr error) error {
	err := s.Delete(id)
	if err != nil {
		dlog.Warnf("Rollback failed volume %v, Error %v", id, err)
	}
	return createErr
}

func (s *Ec2Ops) refreshVol(id *string) (*ec2.Volume, error) {
	vols, err := s.Inspect([]*string{id})
	if err != nil {
		return nil, err
	}

	if len(vols) != 1 {
		return nil, fmt.Errorf("failed to get vol: %s."+
			"Found: %d volumes on inspecting", *id, len(vols))
	}

	resp, ok := vols[0].(*ec2.Volume)
	if !ok {
		return nil, storageops.NewStorageError(storageops.ErrVolInval,
			fmt.Sprintf("Invalid volume returned by inspect API for vol: %s", *id),
			"")
	}

	return resp, nil
}

func (s *Ec2Ops) deleted(v *ec2.Volume) bool {
	return *v.State == ec2.VolumeStateDeleting ||
		*v.State == ec2.VolumeStateDeleted
}

func (s *Ec2Ops) available(v *ec2.Volume) bool {
	return *v.State == ec2.VolumeStateAvailable
}

func (s *Ec2Ops) GetDeviceID(vol interface{}) (string, error) {
	if d, ok := vol.(*ec2.Volume); ok {
		return *d.VolumeId, nil
	} else if d, ok := vol.(*ec2.Snapshot); ok {
		return *d.SnapshotId, nil
	} else {
		return "", fmt.Errorf("invalid type: %v given to GetDeviceID", vol)
	}
}

func (s *Ec2Ops) Inspect(volumeIds []*string) ([]interface{}, error) {
	req := &ec2.DescribeVolumesInput{VolumeIds: volumeIds}
	resp, err := s.ec2.DescribeVolumes(req)
	if err != nil {
		return nil, err
	}

	var awsVols = make([]interface{}, len(resp.Volumes))
	for i, v := range resp.Volumes {
		awsVols[i] = v
	}

	return awsVols, nil
}

func (s *Ec2Ops) Tags(volumeID string) (map[string]string, error) {
	vol, err := s.refreshVol(&volumeID)
	if err != nil {
		return nil, err
	}

	labels := make(map[string]string)
	for _, tag := range vol.Tags {
		labels[*tag.Key] = *tag.Value
	}
	return labels, nil
}

func (s *Ec2Ops) Enumerate(
	volumeIds []*string,
	labels map[string]string,
	setIdentifier string,
) (map[string][]interface{}, error) {
	sets := make(map[string][]interface{})

	// Enumerate all volumes that have same labels.
	f := s.filters(labels, nil)
	req := &ec2.DescribeVolumesInput{Filters: f, VolumeIds: volumeIds}
	awsVols, err := s.ec2.DescribeVolumes(req)
	if err != nil {
		return nil, err
	}

	// Volume sets are identified by volumes with the same setIdentifer.
	found := false
	for _, vol := range awsVols.Volumes {
		if s.deleted(vol) {
			continue
		}
		if len(setIdentifier) == 0 {
			storageops.AddElementToMap(sets, vol, storageops.SetIdentifierNone)
		} else {
			found = false
			for _, tag := range vol.Tags {
				if s.matchTag(tag, setIdentifier) {
					storageops.AddElementToMap(sets, vol, *tag.Value)
					found = true
					break
				}
			}
			if !found {
				storageops.AddElementToMap(sets, vol, storageops.SetIdentifierNone)
			}
		}
	}

	return sets, nil
}

func (s *Ec2Ops) Create(
	v interface{},
	labels map[string]string,
) (interface{}, error) {
	vol, ok := v.(*ec2.Volume)
	if !ok {
		return nil, storageops.NewStorageError(storageops.ErrVolInval,
			"Invalid volume template given", "")
	}

	req := &ec2.CreateVolumeInput{
		AvailabilityZone: vol.AvailabilityZone,
		Encrypted:        vol.Encrypted,
		KmsKeyId:         vol.KmsKeyId,
		Size:             vol.Size,
		VolumeType:       vol.VolumeType,
		SnapshotId:       vol.SnapshotId,
	}
	if *vol.VolumeType == opsworks.VolumeTypeIo1 {
		req.Iops = vol.Iops
	}

	resp, err := s.ec2.CreateVolume(req)
	if err != nil {
		return nil, err
	}
	if err = s.waitStatus(
		*resp.VolumeId,
		ec2.VolumeStateAvailable,
	); err != nil {
		return nil, s.rollbackCreate(*resp.VolumeId, err)
	}
	if len(labels) > 0 {
		if err = s.ApplyTags(*resp.VolumeId, labels); err != nil {
			return nil, s.rollbackCreate(*resp.VolumeId, err)
		}
	}

	return s.refreshVol(resp.VolumeId)
}

func (s *Ec2Ops) Delete(id string) error {
	req := &ec2.DeleteVolumeInput{VolumeId: &id}
	_, err := s.ec2.DeleteVolume(req)
	return err
}

func (s *Ec2Ops) Attach(instanceID, volumeID string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	self, err := s.DescribeInstance(instanceID)
	if err != nil {
		return "", err
	}

	var blockDeviceMappings = make([]interface{}, len(self.BlockDeviceMappings))
	for i, b := range self.BlockDeviceMappings {
		blockDeviceMappings[i] = b
	}

	devices, err := s.FreeDevices(blockDeviceMappings, *self.RootDeviceName)
	if err != nil {
		return "", err
	}
	req := &ec2.AttachVolumeInput{
		Device:     &devices[0],
		InstanceId: &instanceID,
		VolumeId:   &volumeID,
	}
	if _, err = s.ec2.AttachVolume(req); err != nil {
		return "", err
	}
	vol, err := s.waitAttachmentStatus(
		volumeID,
		ec2.VolumeAttachmentStateAttached,
		time.Minute,
	)
	if err != nil {
		return "", err
	}
	return s.DevicePath(instanceID, *vol.VolumeId)
}

func (s *Ec2Ops) Detach(instanceID, volumeID string) error {
	force := false
	req := &ec2.DetachVolumeInput{
		InstanceId: &instanceID,
		VolumeId:   &volumeID,
		Force:      &force,
	}
	if _, err := s.ec2.DetachVolume(req); err != nil {
		return err
	}
	_, err := s.waitAttachmentStatus(volumeID,
		ec2.VolumeAttachmentStateDetached,
		time.Minute,
	)
	return err
}

func (s *Ec2Ops) Snapshot(
	volumeID string,
	readonly bool,
) (interface{}, error) {
	request := &ec2.CreateSnapshotInput{
		VolumeId: &volumeID,
	}
	return s.ec2.CreateSnapshot(request)
}

func (s *Ec2Ops) SnapshotDelete(snapID string) error {
	request := &ec2.DeleteSnapshotInput{
		SnapshotId: &snapID,
	}

	_, err := s.ec2.DeleteSnapshot(request)
	return err
}

func (s *Ec2Ops) DevicePath(instanceID, volumeID string) (string, error) {
	vol, err := s.refreshVol(&volumeID)
	if err != nil {
		return "", err
	}

	if vol.Attachments == nil || len(vol.Attachments) == 0 {
		return "", storageops.NewStorageError(storageops.ErrVolDetached,
			"Volume is detached", *vol.VolumeId)
	}
	if vol.Attachments[0].InstanceId == nil {
		return "", storageops.NewStorageError(storageops.ErrVolInval,
			"Unable to determine volume instance attachment", "")
	}
	if instanceID != *vol.Attachments[0].InstanceId {
		return "", storageops.NewStorageError(storageops.ErrVolAttachedOnRemoteNode,
			fmt.Sprintf("Volume attached on %q current instance %q",
				*vol.Attachments[0].InstanceId, instanceID),
			*vol.Attachments[0].InstanceId)

	}
	if vol.Attachments[0].State == nil {
		return "", storageops.NewStorageError(storageops.ErrVolInval,
			"Unable to determine volume attachment state", "")
	}
	if *vol.Attachments[0].State != ec2.VolumeAttachmentStateAttached {
		return "", storageops.NewStorageError(storageops.ErrVolInval,
			fmt.Sprintf("Invalid state %q, volume is not attached",
				*vol.Attachments[0].State), "")
	}
	if vol.Attachments[0].Device == nil {
		return "", storageops.NewStorageError(storageops.ErrVolInval,
			"Unable to determine volume attachment path", "")
	}
	dev := strings.TrimPrefix(*vol.Attachments[0].Device, "/dev/sd")
	if dev != *vol.Attachments[0].Device {
		dev = "/dev/xvd" + dev
	}
	return dev, nil
}
