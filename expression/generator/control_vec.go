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

// +build ignore

package main

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"log"
	"path/filepath"
	"text/template"

	. "github.com/pingcap/tidb/expression/generator/helper"
)

const header = `// Copyright 2019 PingCAP, Inc.
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

// Code generated by go generate in expression/generator; DO NOT EDIT.

package expression

import (
	"github.com/pingcap/tidb/types"
	"github.com/pingcap/tidb/util/chunk"
)
`

var builtinIfNullVec = template.Must(template.New("builtinIfNullVec").Parse(`
{{ range .Sigs }}{{ with .Arg0 }}
func (b *builtinIfNull{{ .TypeName }}Sig) vecEval{{ .TypeName }}(input *chunk.Chunk, result *chunk.Column) error {
	n := input.NumRows()

	{{- if .Fixed }}
	if err := b.args[0].VecEval{{ .TypeName }}(b.ctx, input, result); err != nil {
		return err
	}
	buf1, err := b.bufAllocator.get(types.ET{{ .ETName }}, n)
	if err != nil {
		return err
	}
	defer b.bufAllocator.put(buf1)
	if err := b.args[1].VecEval{{ .TypeName }}(b.ctx, input, buf1); err != nil {
		return err
	}

	arg0 := result.{{ .TypeNameInColumn }}s()
	arg1 := buf1.{{ .TypeNameInColumn }}s()
	for i := 0; i < n; i++ {
		if result.IsNull(i) && !buf1.IsNull(i) {
			result.SetNull(i, false)
			arg0[i] = arg1[i]
		}
	}
	{{ else }}
	buf0, err := b.bufAllocator.get(types.ET{{ .ETName }}, n)
	if err != nil {
		return err
	}
	defer b.bufAllocator.put(buf0)
	if err := b.args[0].VecEval{{ .TypeName }}(b.ctx, input, buf0); err != nil {
		return err
	}
	buf1, err := b.bufAllocator.get(types.ET{{ .ETName }}, n)
	if err != nil {
		return err
	}
	defer b.bufAllocator.put(buf1)
	if err := b.args[1].VecEval{{ .TypeName }}(b.ctx, input, buf1); err != nil {
		return err
	}

	result.Reserve{{ .TypeNameInColumn }}(n)
	for i := 0; i < n; i++ {
		if !buf0.IsNull(i) {
			result.Append{{ .TypeNameInColumn }}(buf0.Get{{ .TypeNameInColumn }}(i))
		} else if !buf1.IsNull(i) {
			result.Append{{ .TypeNameInColumn }}(buf1.Get{{ .TypeNameInColumn }}(i))
		} else {
			result.AppendNull()
		}
	}
	{{ end -}}
	return nil
}

func (b *builtinIfNull{{ .TypeName }}Sig) vectorized() bool {
	return true
}
{{ end }}{{/* with */}}
{{ end }}{{/* range .Sigs */}}
`))

var builtinIfVec = template.Must(template.New("builtinIfVec").Parse(`
{{ range .Sigs }}{{ with .Arg0 }}
func (b *builtinIf{{ .TypeName }}Sig) vecEval{{ .TypeName }}(input *chunk.Chunk, result *chunk.Column) error {
	n := input.NumRows()
	buf0, err := b.bufAllocator.get(types.ETInt, n)
	if err != nil {
		return err
	}
	defer b.bufAllocator.put(buf0)
	if err := b.args[0].VecEvalInt(b.ctx, input, buf0); err != nil {
		return err
	}

{{- if .Fixed }}
	if err := b.args[1].VecEval{{ .TypeName }}(b.ctx, input, result); err != nil {
		return err
	}
{{- else }}
	buf1, err := b.bufAllocator.get(types.ET{{ .ETName }}, n)
	if err != nil {
		return err
	}
	defer b.bufAllocator.put(buf1)
	if err := b.args[1].VecEval{{ .TypeName }}(b.ctx, input, buf1); err != nil {
		return err
	}
{{- end }}
	buf2, err := b.bufAllocator.get(types.ET{{ .ETName }}, n)
	if err != nil {
		return err
	}
	defer b.bufAllocator.put(buf2)
	if err := b.args[2].VecEval{{ .TypeName }}(b.ctx, input, buf2); err != nil {
		return err
	}

{{ if not .Fixed }}
	result.Reserve{{ .TypeNameInColumn }}(n)
{{- end }}
	arg0 := buf0.Int64s()
{{- if .Fixed }}
	arg2 := buf2.{{ .TypeNameInColumn }}s()
	rs := result.{{ .TypeNameInColumn }}s()
{{- end }}
	for i := 0; i < n; i++ {
		arg := arg0[i]
		isNull0 := buf0.IsNull(i)
		switch {
		case isNull0 || arg == 0:
{{- if .Fixed }}
			if buf2.IsNull(i) {
				result.SetNull(i, true)
			} else {
				result.SetNull(i, false)
				rs[i] = arg2[i]
			}
{{- else }}
			if buf2.IsNull(i) {
				result.AppendNull()
			} else {
				result.Append{{ .TypeNameInColumn }}(buf2.Get{{ .TypeNameInColumn }}(i))
			}
{{- end }}
		case arg != 0:
{{- if .Fixed }}
{{- else }}
			if buf1.IsNull(i) {
				result.AppendNull()
			} else {
				result.Append{{ .TypeNameInColumn }}(buf1.Get{{ .TypeNameInColumn }}(i))
			}
{{- end }}
		}
	}
	return nil
}

func (b *builtinIf{{ .TypeName }}Sig) vectorized() bool {
	return true
}
{{ end }}{{/* with */}}
{{ end }}{{/* range .Sigs */}}
`))

var testFile = template.Must(template.New("testFile").Parse(`// Copyright 2019 PingCAP, Inc.
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

// Code generated by go generate in expression/generator; DO NOT EDIT.

package expression

import (
	"math/rand"
	"testing"

	. "github.com/pingcap/check"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/tidb/types"
)

var defaultControlIntGener = &controlIntGener{zeroRation: 0.3, defaultGener: defaultGener{0.3, types.ETInt}}

type controlIntGener struct {
	zeroRation float64
	defaultGener
}

func (g *controlIntGener) gen() interface{} {
	if rand.Float64() < g.zeroRation {
		return int64(0)
	}
	return g.defaultGener.gen()
}

{{/* Add more test cases here if we have more functions in this file */}}
var vecBuiltin{{.Category}}Cases = map[string][]vecExprBenchCase{
{{ with index .Functions 0 }}
	ast.Ifnull: {
	{{ range .Sigs }}
		{retEvalType: types.ET{{ .Arg0.ETName }}, childrenTypes: []types.EvalType{types.ET{{ .Arg0.ETName }}, types.ET{{ .Arg0.ETName }}}},
	{{ end }}
	},
{{ end }}

{{ with index .Functions 1 }}
	ast.If: {
	{{ range .Sigs }}
		{retEvalType: types.ET{{ .Arg0.ETName }}, childrenTypes: []types.EvalType{types.ETInt, types.ET{{ .Arg0.ETName }}, types.ET{{ .Arg0.ETName }}}, geners: []dataGenerator{defaultControlIntGener}},
	{{ end }}
	},
{{ end }}
}

func (s *testEvaluatorSuite) TestVectorizedBuiltin{{.Category}}EvalOneVecGenerated(c *C) {
	testVectorizedEvalOneVec(c, vecBuiltinControlCases)
}

func (s *testEvaluatorSuite) TestVectorizedBuiltin{{.Category}}FuncGenerated(c *C) {
	testVectorizedBuiltinFunc(c, vecBuiltinControlCases)
}

func BenchmarkVectorizedBuiltin{{.Category}}EvalOneVecGenerated(b *testing.B) {
	benchmarkVectorizedEvalOneVec(b, vecBuiltinControlCases)
}

func BenchmarkVectorizedBuiltin{{.Category}}FuncGenerated(b *testing.B) {
	benchmarkVectorizedBuiltinFunc(b, vecBuiltinControlCases)
}
`))

type typeContext struct {
	// Describe the name of "github.com/pingcap/tidb/types".ET{{ .ETName }}
	ETName string
	// Describe the name of "github.com/pingcap/tidb/expression".VecExpr.VecEval{{ .TypeName }}
	// If undefined, it's same as ETName.
	TypeName string
	// Describe the name of "github.com/pingcap/tidb/util/chunk".*Column.Append{{ .TypeNameInColumn }},
	// Resize{{ .TypeNameInColumn }}, Reserve{{ .TypeNameInColumn }}, Get{{ .TypeNameInColumn }} and
	// {{ .TypeNameInColumn }}s.
	// If undefined, it's same as TypeName.
	TypeNameInColumn string
	// Same as "github.com/pingcap/tidb/util/chunk".getFixedLen()
	Fixed bool
}

var ifNullSigs = []sig{
	{Arg0: TypeInt},
	{Arg0: TypeReal},
	{Arg0: TypeDecimal},
	{Arg0: TypeString},
	{Arg0: TypeDatetime},
	{Arg0: TypeDuration},
}

var ifSigs = []sig{
	{Arg0: TypeInt},
	{Arg0: TypeReal},
	{Arg0: TypeDecimal},
	{Arg0: TypeString},
	{Arg0: TypeDatetime},
	{Arg0: TypeDuration},
}

type sig struct {
	Arg0 TypeContext
}

type function struct {
	FuncName string
	Sigs     []sig
	Tmpl     *template.Template
}

var tmplVal = struct {
	Category  string
	Functions []function
}{
	Category: "Control",
	Functions: []function{
		{FuncName: "Ifnull", Sigs: ifNullSigs, Tmpl: builtinIfNullVec},
		{FuncName: "If", Sigs: ifSigs, Tmpl: builtinIfVec},
	},
}

func generateDotGo(fileName string) error {
	w := new(bytes.Buffer)
	w.WriteString(header)
	for _, function := range tmplVal.Functions {
		err := function.Tmpl.Execute(w, function)
		if err != nil {
			return err
		}
	}
	data, err := format.Source(w.Bytes())
	if err != nil {
		log.Println("[Warn]", fileName+": gofmt failed", err)
		data = w.Bytes() // write original data for debugging
	}
	return ioutil.WriteFile(fileName, data, 0644)
}

func generateTestDotGo(fileName string) error {
	w := new(bytes.Buffer)
	err := testFile.Execute(w, tmplVal)
	if err != nil {
		return err
	}
	data, err := format.Source(w.Bytes())
	if err != nil {
		log.Println("[Warn]", fileName+": gofmt failed", err)
		data = w.Bytes() // write original data for debugging
	}
	return ioutil.WriteFile(fileName, data, 0644)
}

// generateOneFile generate one xxx.go file and the associated xxx_test.go file.
func generateOneFile(fileNamePrefix string) (err error) {

	err = generateDotGo(fileNamePrefix + ".go")
	if err != nil {
		return
	}
	err = generateTestDotGo(fileNamePrefix + "_test.go")
	return
}

func main() {
	var err error
	outputDir := "."
	err = generateOneFile(filepath.Join(outputDir, "builtin_control_vec_generated"))
	if err != nil {
		log.Fatalln("generateOneFile", err)
	}
}
