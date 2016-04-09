package context

import (
	"github.com/boltdb/bolt"
)

type Context struct {
	BucketName []byte
	DB         *bolt.DB
}

func New(bucketName []byte, db *bolt.DB) *Context {
	return &Context{BucketName: bucketName, DB: db}
}
