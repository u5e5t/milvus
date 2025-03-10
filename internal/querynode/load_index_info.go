// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package querynode

/*
#cgo pkg-config: milvus_common milvus_segcore

#include "segcore/load_index_c.h"
#include "common/vector_index_c.h"
*/
import "C"

import (
	"path/filepath"
	"unsafe"

	"github.com/milvus-io/milvus/api/schemapb"
	"github.com/milvus-io/milvus/internal/proto/querypb"
	"github.com/milvus-io/milvus/internal/util/funcutil"
)

// LoadIndexInfo is a wrapper of the underlying C-structure C.CLoadIndexInfo
type LoadIndexInfo struct {
	cLoadIndexInfo C.CLoadIndexInfo
}

// newLoadIndexInfo returns a new LoadIndexInfo and error
func newLoadIndexInfo() (*LoadIndexInfo, error) {
	var cLoadIndexInfo C.CLoadIndexInfo
	status := C.NewLoadIndexInfo(&cLoadIndexInfo)
	if err := HandleCStatus(&status, "NewLoadIndexInfo failed"); err != nil {
		return nil, err
	}
	return &LoadIndexInfo{cLoadIndexInfo: cLoadIndexInfo}, nil
}

// deleteLoadIndexInfo would delete C.CLoadIndexInfo
func deleteLoadIndexInfo(info *LoadIndexInfo) {
	C.DeleteLoadIndexInfo(info.cLoadIndexInfo)
}

func (li *LoadIndexInfo) appendIndexInfo(bytesIndex [][]byte, indexInfo *querypb.FieldIndexInfo, fieldType schemapb.DataType) error {
	fieldID := indexInfo.FieldID
	indexParams := funcutil.KeyValuePair2Map(indexInfo.IndexParams)
	indexPaths := indexInfo.IndexFilePaths

	err := li.appendFieldInfo(fieldID, fieldType)
	if err != nil {
		return err
	}
	for key, value := range indexParams {
		err = li.appendIndexParam(key, value)
		if err != nil {
			return err
		}
	}
	err = li.appendIndexData(bytesIndex, indexPaths)
	return err
}

// appendIndexParam append indexParam to index
func (li *LoadIndexInfo) appendIndexParam(indexKey string, indexValue string) error {
	cIndexKey := C.CString(indexKey)
	defer C.free(unsafe.Pointer(cIndexKey))
	cIndexValue := C.CString(indexValue)
	defer C.free(unsafe.Pointer(cIndexValue))
	status := C.AppendIndexParam(li.cLoadIndexInfo, cIndexKey, cIndexValue)
	return HandleCStatus(&status, "AppendIndexParam failed")
}

// appendFieldInfo appends fieldID & fieldType to index
func (li *LoadIndexInfo) appendFieldInfo(fieldID FieldID, fieldType schemapb.DataType) error {
	cFieldID := C.int64_t(fieldID)
	cintDType := uint32(fieldType)
	status := C.AppendFieldInfo(li.cLoadIndexInfo, cFieldID, cintDType)
	return HandleCStatus(&status, "AppendFieldInfo failed")
}

// appendIndexData appends binarySet index to cLoadIndexInfo
func (li *LoadIndexInfo) appendIndexData(bytesIndex [][]byte, indexKeys []string) error {
	var cBinarySet C.CBinarySet
	status := C.NewBinarySet(&cBinarySet)
	defer C.DeleteBinarySet(cBinarySet)

	if err := HandleCStatus(&status, "NewBinarySet failed"); err != nil {
		return err
	}

	for i, byteIndex := range bytesIndex {
		indexPtr := unsafe.Pointer(&byteIndex[0])
		indexLen := C.int64_t(len(byteIndex))
		binarySetKey := filepath.Base(indexKeys[i])
		indexKey := C.CString(binarySetKey)
		status = C.AppendIndexBinary(cBinarySet, indexPtr, indexLen, indexKey)
		C.free(unsafe.Pointer(indexKey))
		if err := HandleCStatus(&status, "LoadIndexInfo AppendIndexBinary failed"); err != nil {
			return err
		}
	}

	status = C.AppendIndex(li.cLoadIndexInfo, cBinarySet)
	return HandleCStatus(&status, "AppendIndex failed")
}
