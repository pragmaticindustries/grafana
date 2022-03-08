package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/infra/filestorage"
	"github.com/grafana/grafana/pkg/models"
)

type SaveDashboardRequest struct {
	Path    string
	User    *models.SignedInUser
	Body    json.RawMessage // []byte
	Message string
}

type storageTree interface {
	// Called from the UI when a dashboard is saved
	GetFile(ctx context.Context, path string) (*filestorage.File, error)

	// Get a single dashboard
	ListFolder(ctx context.Context, path string) (*data.Frame, error)
}

//-------------------------------------------
// INTERNAL
//-------------------------------------------

type rootStorageState struct {
	Meta RootStorageMeta

	Store filestorage.FileStorage
}

// TEMPORARY! internally, used for listing and building an index
type DashboardQueryResultForSearchIndex struct {
	Id       int64
	IsFolder bool   `xorm:"is_folder"`
	FolderID int64  `xorm:"folder_id"`
	Slug     string `xorm:"slug"` // path when GIT/ETC
	Data     []byte
	Created  time.Time
	Updated  time.Time
}

type DashboardBodyIterator func() *DashboardQueryResultForSearchIndex

type RootStorageMeta struct {
	ReadOnly bool          `json:"editable,omitempty"`
	Builtin  bool          `json:"builtin,omitempty"`
	Ready    bool          `json:"ready"` // can connect
	Notice   []data.Notice `json:"notice,omitempty"`

	Config RootStorageConfig `json:"config"`
}