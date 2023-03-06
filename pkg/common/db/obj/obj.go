package obj

import (
	"context"
	"net/http"
	"time"
)

type BucketObject struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

type ApplyPutArgs struct {
	Bucket        string
	Name          string
	Effective     time.Duration // 申请有效时间
	Header        http.Header   // header
	MaxObjectSize int64
}

type ObjectInfo struct {
	URL        string
	Size       int64
	Hash       string
	Expiration time.Time
}

type Interface interface {
	// Name 存储名字
	Name() string
	// MinFragmentSize 最小允许的分片大小
	MinFragmentSize() int64
	// MaxFragmentNum 最大允许的分片数量
	MaxFragmentNum() int
	// MinExpirationTime 最小过期时间
	MinExpirationTime() time.Duration
	// TempBucket 临时桶名，用于上传
	TempBucket() string
	// DataBucket 永久存储的桶名
	DataBucket() string
	// GetURL 通过桶名和对象名返回URL
	GetURL(bucket string, name string) string
	// PresignedPutURL 申请上传,返回PUT的上传地址
	PresignedPutURL(ctx context.Context, args *ApplyPutArgs) (string, error)
	// GetObjectInfo 获取对象信息
	GetObjectInfo(ctx context.Context, args *BucketObject) (*ObjectInfo, error)
	// CopyObjet 复制对象
	CopyObjet(ctx context.Context, src *BucketObject, dst *BucketObject) error
	// DeleteObjet 删除对象
	DeleteObjet(ctx context.Context, info *BucketObject) error
	// ComposeObject 合并对象
	ComposeObject(ctx context.Context, src []BucketObject, dst *BucketObject) error
	// IsNotFound 判断是不是不存在导致的错误
	IsNotFound(err error) bool
	// CheckName 检查名字是否可用
	CheckName(name string) error
}