package scaleway

import (
	"errors"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	// MinimumVolumeSizeInBytes represents the size of the smallest block volume on Scaleway
	MinimumVolumeSizeInBytes int64 = 1 * 1000 * 1000 * 1000
	// MaximumVolumeSizeInBytes represents the size of the biggest block volume on Scaleway
	MaximumVolumeSizeInBytes int64 = 1 * 1000 * 1000 * 1000 * 1000
	// MaxVolumesPerNode represents the number max of volumes attached to one node
	MaxVolumesPerNode = 16

	// DefaultVolumeType is the default type for Scaleway Block volumes
	DefaultVolumeType = instance.VolumeTypeBSSD
)

var (
	// ErrMultipleVolumes is the error returned when multiples volumes exists with the same name
	ErrMultipleVolumes = errors.New("multiple volumes exists with the same name")
	// ErrDifferentSize is the error returned when a volume match the given name, but the size doesn't match
	ErrDifferentSize = errors.New("volume exists with a different size")
	// ErrVolumeNotFound is the error returned when the volume was not found
	ErrVolumeNotFound = errors.New("volume not found")

	// ErrSnapshotNotFound is the error returned when the snapshot was not found
	ErrSnapshotNotFound = errors.New("snapshot not found")
	// ErrSnapshotSameName is the error returned when a snapshot with the same name already exists
	ErrSnapshotSameName = errors.New("a snapshot with the same name exists")
	// ErrSnapshotStillSnapshotting is the error returned when a snapshot is still snapshotting
	ErrSnapshotStillSnapshotting = errors.New("snapshot is still snapshotting")
)

// Scaleway is the struct used to communicate withe the Scaleway provider
type Scaleway struct {
	InstanceAPI
}

// NewScaleway returns a new Scaleway object which will use the given user agent
func NewScaleway(userAgent string) *Scaleway {
	client, err := scw.NewClient(
		scw.WithEnv(),
		scw.WithUserAgent(userAgent),
	)
	if err != nil {
		panic(err)
	}
	api := instance.NewAPI(client)
	return &Scaleway{api}
}

// Metadata is an interface for the instance metadata
type Metadata interface {
	GetMetadata() (m *instance.Metadata, err error)
}

// NewMetadata returns a new Metadata object to be used from a Scaleway instance
func NewMetadata() Metadata {
	return instance.NewMetadataAPI()
}

// InstanceAPI is an interface for the Scaleway Go SDK for instance
type InstanceAPI interface {
	// ListVolumes is an interface for the SDK ListVolumes method
	ListVolumes(req *instance.ListVolumesRequest, opts ...scw.RequestOption) (*instance.ListVolumesResponse, error)

	// CreateVolume is an interface for the SDK CreateVolume method
	CreateVolume(req *instance.CreateVolumeRequest, opts ...scw.RequestOption) (*instance.CreateVolumeResponse, error)

	// GetVolume is an interface for the SDK GetVolume method
	GetVolume(req *instance.GetVolumeRequest, opts ...scw.RequestOption) (*instance.GetVolumeResponse, error)

	// DeleteVolume is an interface for the SDK DeleteVolume method
	DeleteVolume(req *instance.DeleteVolumeRequest, opts ...scw.RequestOption) error

	// GetServer is an interface for the SDK GetServer method
	GetServer(req *instance.GetServerRequest, opts ...scw.RequestOption) (*instance.GetServerResponse, error)

	// AttachVolume is an interface for the SDK AttachVolume method
	AttachVolume(req *instance.AttachVolumeRequest, opts ...scw.RequestOption) (*instance.AttachVolumeResponse, error)

	// DetachVolume is an interface for the SDK DetachVolume method
	DetachVolume(req *instance.DetachVolumeRequest, opts ...scw.RequestOption) (*instance.DetachVolumeResponse, error)

	// GetSnapshot is an interface for the SDK  GetSnapshot method
	GetSnapshot(req *instance.GetSnapshotRequest, opts ...scw.RequestOption) (*instance.GetSnapshotResponse, error)

	// ListSnapshots is an interface for the SDK ListSnapshots method
	ListSnapshots(req *instance.ListSnapshotsRequest, opts ...scw.RequestOption) (*instance.ListSnapshotsResponse, error)

	// CreateSnapshot is an interface for the SDK CreateSnapshot method
	CreateSnapshot(req *instance.CreateSnapshotRequest, opts ...scw.RequestOption) (*instance.CreateSnapshotResponse, error)

	// DeleteSnapshot is an interface for the SDK CreateSnapshot method
	DeleteSnapshot(req *instance.DeleteSnapshotRequest, opts ...scw.RequestOption) error
}

// GetVolumeByName is a helper to find a volume by it's name, type and given size
func (s *Scaleway) GetVolumeByName(name string, size int64, volumeType instance.VolumeType) (*instance.Volume, error) {
	volumesResp, err := s.ListVolumes(&instance.ListVolumesRequest{
		Name:       &name,
		VolumeType: volumeType,
	}, scw.WithAllPages())
	if err != nil {
		return nil, err
	}
	volumes := volumesResp.Volumes
	if len(volumes) != 0 {
		if len(volumes) > 1 {
			return nil, ErrMultipleVolumes
		}
		volume := volumes[0]
		if uint64(volume.Size) != uint64(size) {
			return nil, ErrDifferentSize
		}
		return volume, nil
	}
	return nil, ErrVolumeNotFound
}

// GetSnapshotByName is a helper to find a snapshot by it's name and it's source volume ID and zone
func (s *Scaleway) GetSnapshotByName(name string, sourceVolumeID string, sourceVolumeZone scw.Zone) (*instance.Snapshot, error) {
	snapshots, err := s.ListSnapshots(&instance.ListSnapshotsRequest{
		Name: &name,
		Zone: sourceVolumeZone,
	}, scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	for _, snapshot := range snapshots.Snapshots {
		if snapshot.Name == name { // fuzzy search on the API
			if snapshot.BaseVolume == nil || snapshot.BaseVolume.ID == sourceVolumeID {
				if snapshot.State == instance.SnapshotStateSnapshotting {
					return nil, ErrSnapshotStillSnapshotting
				}
				return snapshot, nil
			}
			return nil, ErrSnapshotSameName
		}
	}
	return nil, ErrSnapshotNotFound
}
