// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package expression

import (
	"github.com/pingcap/errors"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/tidb/types"
	"github.com/pingcap/tidb/types/json"
	"github.com/pingcap/tidb/util/chunk"
)

func (b *builtinJSONDepthSig) vectorized() bool {
	return true
}

func (b *builtinJSONDepthSig) vecEvalInt(input *chunk.Chunk, result *chunk.Column) error {
	n := input.NumRows()
	buf, err := b.bufAllocator.get(types.ETJson, n)
	if err != nil {
		return err
	}
	defer b.bufAllocator.put(buf)
	if err := b.args[0].VecEvalJSON(b.ctx, input, buf); err != nil {
		return err
	}
	result.ResizeInt64(n, false)
	result.MergeNulls(buf)
	int64s := result.Int64s()
	for i := 0; i < n; i++ {
		if result.IsNull(i) {
			continue
		}
		j := buf.GetJSON(i)
		int64s[i] = int64(j.GetElemDepth())
	}
	return nil
}

func (b *builtinJSONKeysSig) vectorized() bool {
	return true
}

func (b *builtinJSONKeysSig) vecEvalJSON(input *chunk.Chunk, result *chunk.Column) error {
	n := input.NumRows()
	buf, err := b.bufAllocator.get(types.ETJson, n)
	if err != nil {
		return err
	}
	defer b.bufAllocator.put(buf)
	if err := b.args[0].VecEvalJSON(b.ctx, input, buf); err != nil {
		return err
	}

	result.ReserveJSON(n)
	var j json.BinaryJSON
	for i := 0; i < n; i++ {
		if buf.IsNull(i) {
			result.AppendNull()
			continue
		}

		j = buf.GetJSON(i)
		if j.TypeCode != json.TypeCodeObject {
			return json.ErrInvalidJSONData
		}
		result.AppendJSON(j.GetKeys())
	}
	return nil
}

func (b *builtinJSONInsertSig) vectorized() bool {
	return false
}

func (b *builtinJSONInsertSig) vecEvalJSON(input *chunk.Chunk, result *chunk.Column) error {
	return errors.Errorf("not implemented")
}

func (b *builtinJSONReplaceSig) vectorized() bool {
	return false
}

func (b *builtinJSONReplaceSig) vecEvalJSON(input *chunk.Chunk, result *chunk.Column) error {
	return errors.Errorf("not implemented")
}

func (b *builtinJSONArraySig) vectorized() bool {
	return false
}

func (b *builtinJSONArraySig) vecEvalJSON(input *chunk.Chunk, result *chunk.Column) error {
	return errors.Errorf("not implemented")
}

func (b *builtinJSONContainsSig) vectorized() bool {
	return true
}

func (b *builtinJSONContainsSig) vecEvalInt(input *chunk.Chunk, result *chunk.Column) error {
	nr := input.NumRows()

	objCol, err := b.bufAllocator.get(types.ETJson, nr)
	if err != nil {
		return err
	}
	defer b.bufAllocator.put(objCol)

	if err := b.args[0].VecEvalJSON(b.ctx, input, objCol); err != nil {
		return err
	}

	targetCol, err := b.bufAllocator.get(types.ETJson, nr)
	if err != nil {
		return err
	}
	defer b.bufAllocator.put(targetCol)

	if err := b.args[1].VecEvalJSON(b.ctx, input, targetCol); err != nil {
		return err
	}

	result.ResizeInt64(nr, false)
	resI64s := result.Int64s()

	if len(b.args) == 3 {
		pathCol, err := b.bufAllocator.get(types.ETString, nr)
		if err != nil {
			return err
		}
		defer b.bufAllocator.put(pathCol)

		if err := b.args[2].VecEvalString(b.ctx, input, pathCol); err != nil {
			return err
		}

		result.MergeNulls(objCol, targetCol, pathCol)

		var pathExpr json.PathExpression
		for i := 0; i < nr; i++ {
			if result.IsNull(i) {
				continue
			}
			pathExpr, err = json.ParseJSONPathExpr(pathCol.GetString(i))
			if err != nil {
				return err
			}
			if pathExpr.ContainsAnyAsterisk() {
				return json.ErrInvalidJSONPathWildcard
			}

			obj, exists := objCol.GetJSON(i).Extract([]json.PathExpression{pathExpr})
			if !exists {
				result.SetNull(i, true)
				continue
			}

			if json.ContainsBinary(obj, targetCol.GetJSON(i)) {
				resI64s[i] = 1
			} else {
				resI64s[i] = 0
			}
		}
	} else {
		result.MergeNulls(objCol, targetCol)
		for i := 0; i < nr; i++ {
			if result.IsNull(i) {
				continue
			}
			if json.ContainsBinary(objCol.GetJSON(i), targetCol.GetJSON(i)) {
				resI64s[i] = 1
			} else {
				resI64s[i] = 0
			}
		}
	}

	return nil
}

func (b *builtinJSONQuoteSig) vectorized() bool {
	return true
}

func (b *builtinJSONQuoteSig) vecEvalString(input *chunk.Chunk, result *chunk.Column) error {
	n := input.NumRows()
	buf, err := b.bufAllocator.get(types.ETJson, n)
	if err != nil {
		return err
	}
	defer b.bufAllocator.put(buf)
	if err := b.args[0].VecEvalJSON(b.ctx, input, buf); err != nil {
		return err
	}

	result.ReserveString(n)
	for i := 0; i < n; i++ {
		if buf.IsNull(i) {
			result.AppendNull()
			continue
		}
		result.AppendString(buf.GetJSON(i).Quote())
	}
	return nil
}

func (b *builtinJSONSearchSig) vectorized() bool {
	return false
}

func (b *builtinJSONSearchSig) vecEvalJSON(input *chunk.Chunk, result *chunk.Column) error {
	return errors.Errorf("not implemented")
}

func (b *builtinJSONSetSig) vectorized() bool {
	return false
}

func (b *builtinJSONSetSig) vecEvalJSON(input *chunk.Chunk, result *chunk.Column) error {
	return errors.Errorf("not implemented")
}

func (b *builtinJSONObjectSig) vectorized() bool {
	return true
}

func (b *builtinJSONObjectSig) vecEvalJSON(input *chunk.Chunk, result *chunk.Column) error {
	nr := input.NumRows()
	if len(b.args)&1 == 1 {
		err := ErrIncorrectParameterCount.GenWithStackByArgs(ast.JSONObject)
		return err
	}

	jsons := make([]map[string]interface{}, nr)
	for i := 0; i < nr; i++ {
		jsons[i] = make(map[string]interface{}, len(b.args)>>1)
	}

	argBuffers := make([]*chunk.Column, len(b.args))
	var err error
	for i := 0; i < len(b.args); i++ {
		if i&1 == 0 {
			if argBuffers[i], err = b.bufAllocator.get(types.ETString, nr); err != nil {
				return err
			}
			defer func(buf *chunk.Column) {
				b.bufAllocator.put(buf)
			}(argBuffers[i])

			if err = b.args[i].VecEvalString(b.ctx, input, argBuffers[i]); err != nil {
				return err
			}
		} else {
			if argBuffers[i], err = b.bufAllocator.get(types.ETJson, nr); err != nil {
				return err
			}
			defer func(buf *chunk.Column) {
				b.bufAllocator.put(buf)
			}(argBuffers[i])

			if err = b.args[i].VecEvalJSON(b.ctx, input, argBuffers[i]); err != nil {
				return err
			}
		}
	}

	result.ReserveJSON(nr)
	for i := 0; i < len(b.args); i++ {
		if i&1 == 1 {
			keyCol := argBuffers[i-1]
			valueCol := argBuffers[i]

			var key string
			var value json.BinaryJSON
			for j := 0; j < nr; j++ {
				if keyCol.IsNull(j) {
					err := errors.New("JSON documents may not contain NULL member names")
					return err
				}
				key = keyCol.GetString(j)
				if valueCol.IsNull(j) {
					value = json.CreateBinary(nil)
				} else {
					value = valueCol.GetJSON(j)
				}
				jsons[j][key] = value
			}
		}
	}

	for i := 0; i < nr; i++ {
		result.AppendJSON(json.CreateBinary(jsons[i]))
	}
	return nil
}

func (b *builtinJSONArrayInsertSig) vectorized() bool {
	return false
}

func (b *builtinJSONArrayInsertSig) vecEvalJSON(input *chunk.Chunk, result *chunk.Column) error {
	return errors.Errorf("not implemented")
}

func (b *builtinJSONKeys2ArgsSig) vectorized() bool {
	return false
}

func (b *builtinJSONKeys2ArgsSig) vecEvalJSON(input *chunk.Chunk, result *chunk.Column) error {
	return errors.Errorf("not implemented")
}

func (b *builtinJSONLengthSig) vectorized() bool {
	return false
}

func (b *builtinJSONLengthSig) vecEvalInt(input *chunk.Chunk, result *chunk.Column) error {
	return errors.Errorf("not implemented")
}

func (b *builtinJSONTypeSig) vectorized() bool {
	return true
}

func (b *builtinJSONTypeSig) vecEvalString(input *chunk.Chunk, result *chunk.Column) error {
	n := input.NumRows()
	buf, err := b.bufAllocator.get(types.ETJson, n)
	if err != nil {
		return err
	}
	defer b.bufAllocator.put(buf)
	if err := b.args[0].VecEvalJSON(b.ctx, input, buf); err != nil {
		return err
	}

	result.ReserveString(n)
	for i := 0; i < n; i++ {
		if buf.IsNull(i) {
			result.AppendNull()
			continue
		}
		result.AppendString(buf.GetJSON(i).Type())
	}
	return nil
}

func (b *builtinJSONExtractSig) vectorized() bool {
	return false
}

func (b *builtinJSONExtractSig) vecEvalJSON(input *chunk.Chunk, result *chunk.Column) error {
	return errors.Errorf("not implemented")
}

func (b *builtinJSONRemoveSig) vectorized() bool {
	return false
}

func (b *builtinJSONRemoveSig) vecEvalJSON(input *chunk.Chunk, result *chunk.Column) error {
	return errors.Errorf("not implemented")
}

func (b *builtinJSONMergeSig) vectorized() bool {
	return false
}

func (b *builtinJSONMergeSig) vecEvalJSON(input *chunk.Chunk, result *chunk.Column) error {
	return errors.Errorf("not implemented")
}

func (b *builtinJSONContainsPathSig) vectorized() bool {
	return false
}

func (b *builtinJSONContainsPathSig) vecEvalInt(input *chunk.Chunk, result *chunk.Column) error {
	return errors.Errorf("not implemented")
}

func (b *builtinJSONArrayAppendSig) vectorized() bool {
	return false
}

func (b *builtinJSONArrayAppendSig) vecEvalJSON(input *chunk.Chunk, result *chunk.Column) error {
	return errors.Errorf("not implemented")
}

func (b *builtinJSONUnquoteSig) vectorized() bool {
	return true
}

func (b *builtinJSONUnquoteSig) vecEvalString(input *chunk.Chunk, result *chunk.Column) error {
	n := input.NumRows()
	buf, err := b.bufAllocator.get(types.ETJson, n)
	if err != nil {
		return err
	}
	defer b.bufAllocator.put(buf)
	if err := b.args[0].VecEvalJSON(b.ctx, input, buf); err != nil {
		return err
	}

	result.ReserveString(n)
	for i := 0; i < n; i++ {
		if buf.IsNull(i) {
			result.AppendNull()
			continue
		}
		res, err := buf.GetJSON(i).Unquote()
		if err != nil {
			return err
		}
		result.AppendString(res)
	}
	return nil
}
