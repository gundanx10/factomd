// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

type IDatabase interface {
	Close() error
	Put(bucket, key []byte, data BinaryMarshallable) error
	Get(bucket, key []byte, destination BinaryMarshallable) (BinaryMarshallable, error)
	Delete(bucket, key []byte) error
	ListAllKeys(bucket []byte) ([][]byte, error)
	GetAll(bucket []byte, sample BinaryMarshallableAndCopyable) ([]BinaryMarshallableAndCopyable, error)
	Clear(bucket []byte) error
	PutInBatch(records []Record) error
}

type Record struct {
	Bucket []byte
	Key    []byte
	Data   BinaryMarshallable
}

type DatabaseBatchable interface {
	BinaryMarshallable
	GetDatabaseHeight() uint32

	DatabasePrimaryIndex() IHash   //block.KeyMR()
	DatabaseSecondaryIndex() IHash //block.GetHash()

	GetChainID() []byte
}
