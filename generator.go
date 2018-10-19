/*
 * Copyright © 2017 Xiao Zhang <zzxx513@gmail.com>.
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file.
 */
package turbo

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

// Generator generates proto/thrift code
type Generator struct {
	RpcType        string
	PkgPath        string
	ConfigFileName string
	Options        string
	c              *Config
}

// Generate proto/thrift code
func (g *Generator) Generate() {
	if g.RpcType != "grpc" && g.RpcType != "thrift" {
		panic("Invalid server type, should be (grpc|thrift)")
	}
	g.c = NewConfig(g.RpcType, GOPATH()+"/src/"+g.PkgPath+"/"+g.ConfigFileName+".yaml")
	if g.RpcType == "grpc" {
		g.GenerateProtobufStub()
		g.c.loadFieldMapping()
		g.GenerateGrpcSwitcher()
	} else if g.RpcType == "thrift" {
		g.GenerateThriftStub()
		g.GenerateBuildThriftParameters()
		g.c.loadFieldMapping()
		g.GenerateThriftSwitcher()
	}
}

func writeFileWithTemplate(filePath string, data interface{}, text string) {
	f, err := os.Create(filePath)
	panicIf(err)

	tmpl, err := template.New("").Parse(text)
	panicIf(err)

	err = tmpl.Execute(f, data)
	panicIf(err)
}

// GenerateGrpcSwitcher generates "grpcswither.go"
func (g *Generator) GenerateGrpcSwitcher() {
	if _, err := os.Stat(g.c.ServiceRootPathAbsolute() + "/gen"); os.IsNotExist(err) {
		os.Mkdir(g.c.ServiceRootPathAbsolute()+"/gen", 0755)
	}
	serviceMethodMap := methodNames(g.c.mappings[urlServiceMaps])
	structFields := make(map[string][]string, len(serviceMethodMap))
	for s, methods := range serviceMethodMap {
		fields := make([]string, len(methods))
		for i, m := range methods {
			fields[i] = g.structFields(m + "Request")
		}
		structFields[s] = fields
	}
	writeFileWithTemplate(
		g.c.ServiceRootPathAbsolute()+"/gen/grpcswitcher.go",
		struct {
			ServiceMethodMap map[string][]string
			PkgPath          string
			ServiceName      []string
			StructFields     map[string][]string
		}{
			serviceMethodMap,
			g.PkgPath,
			g.c.GrpcServiceNames(),
			structFields,
		},
		`// Code generated by turbo. DO NOT EDIT.
package gen

import (
	g "{{.PkgPath}}/gen/proto"
	"github.com/vaporz/turbo"
	"net/http"
	"errors"
)

// GrpcSwitcher is a runtime func with which a server starts.
var GrpcSwitcher = func(s turbo.Servable, serviceName, methodName string, resp http.ResponseWriter, req *http.Request) (rpcResponse interface{}, err error) {
	callOptions, header, trailer, peer := turbo.CallOptions(serviceName, methodName, req){{range $Service, $Methods := .ServiceMethodMap}}
	if serviceName == "{{$Service}}" {
		switch methodName { {{range $i, $MethodName := $Methods}}
		case "{{$MethodName}}":
			request := &g.{{$MethodName}}Request{ {{index $.StructFields $Service $i}} }
			err = turbo.BuildRequest(s, request, req)
			if err != nil {
				return nil, err
			}
			rpcResponse, err = s.Service("{{$Service}}").(g.{{$Service}}Client).{{$MethodName}}(req.Context(), request, callOptions...){{end}}
		default:
			return nil, errors.New("No such method[" + methodName + "]")
		}
	}{{end}}
	if rpcResponse==nil && err==nil {
		return nil, errors.New("No such service[" + serviceName + "]")
	}
	turbo.WithCallOptions(req, header, trailer, peer)
	return
}
`)
}

func (g *Generator) structFields(structName string) string {
	fields, ok := g.c.fieldMappings[structName]
	if !ok {
		return ""
	}
	var fieldStr string
	for _, field := range fields {
		if len(strings.TrimSpace(field)) == 0 {
			continue
		}
		pair := strings.Split(field, " ")
		nameSlice := []rune(pair[1])
		name := strings.ToUpper(string(nameSlice[0])) + string(nameSlice[1:])
		typeName := pair[0]
		fieldStr = fieldStr + name + ": &g." + typeName + "{" + g.structFields(typeName) + "},"
	}
	return fieldStr
}

// GenerateProtobufStub generates protobuf stub codes
func (g *Generator) GenerateProtobufStub() {
	if _, err := os.Stat(g.c.ServiceRootPathAbsolute() + "/gen/proto"); os.IsNotExist(err) {
		os.MkdirAll(g.c.ServiceRootPathAbsolute()+"/gen/proto", 0755)
	}
	cmd := "protoc " + g.Options + " --go_out=plugins=grpc:" + g.c.ServiceRootPathAbsolute() + "/gen/proto" +
		" --buildfields_out=service_root_path=" + g.c.ServiceRootPathAbsolute() + ":" + g.c.ServiceRootPathAbsolute() + "/gen/proto"

	executeCmd("bash", "-c", cmd)
}

// GenerateBuildThriftParameters generates "build.go"
func (g *Generator) GenerateBuildThriftParameters() {
	writeFileWithTemplate(
		g.c.ServiceRootPathAbsolute()+"/gen/thrift/build.go",
		struct {
			PkgPath         string
			ServiceNames     []string
			ServiceRootPath string
			// todo
			ServiceMethodMap map[string][]string
		}{
			g.PkgPath,
			g.c.GrpcServiceNames(),
			g.c.ServiceRootPathAbsolute(),
			methodNames(g.c.mappings[urlServiceMaps])},
		buildThriftParameters,
	)
	g.runBuildThriftFields()
}

func (g *Generator) runBuildThriftFields() {
	cmd := "go run " + g.c.ServiceRootPathAbsolute() + "/gen/thrift/build.go"
	c := exec.Command("bash", "-c", cmd)
	c.Stdin = os.Stdin
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	panicIf(c.Run())
}

var buildThriftParameters = `package main

import (
	"flag"
	"fmt"
	g "{{.PkgPath}}/gen/thrift/gen-go/gen"
	"io"
	"os"
	"reflect"
	"strings"
	"text/template"
)

var serviceMethodName = flag.String("n", "", "")

func main() {
	flag.Parse()
	if len(strings.TrimSpace(*serviceMethodName)) > 0 {
		names := strings.Split(*serviceMethodName, ",")
		str := buildParameterStr(names[0], names[1])
		fmt.Print(str)
	} else {
		buildFields()
	}
}

func buildFields() {
	services := []interface{}{ {{range $i, $ServiceName := .ServiceNames}}
		new(g.{{$ServiceName}}),{{end}}
	}
	var list string
	for _, i := range services {
		t := reflect.TypeOf(i).Elem()
		numMethod := t.NumMethod()
		items := make([]string, 0)
		for i := 0; i < numMethod; i++ {
			method := t.Method(i)
			numIn := method.Type.NumIn()
			for j := 0; j < numIn; j++ {
				argType := method.Type.In(j)
				argStr := argType.String()
				if argType.Kind() == reflect.Ptr && argType.Elem().Kind() == reflect.Struct {
					arr := strings.Split(argStr, ".")
					name := arr[len(arr)-1:][0]
					items = findItem(items, name, argType)
				}
			}
		}
		for _, s := range items {
			list += s + "\n"
		}
	}
	writeFileWithTemplate(
		"{{.ServiceRootPath}}/gen/thriftfields.yaml",
		fieldsYaml,
		fieldsYamlValues{List: list},
	)
}

func findItem(items []string, name string, structType reflect.Type) []string {
	numField := structType.Elem().NumField()
	item := "  - " + name + "["
	for i := 0; i < numField; i++ {
		fieldType := structType.Elem().Field(i)
		if fieldType.Type.Kind() == reflect.Ptr && fieldType.Type.Elem().Kind() == reflect.Struct {
			arr := strings.Split(fieldType.Type.String(), ".")
			typeName := arr[len(arr)-1:][0]
			argName := fieldType.Name
			item += fmt.Sprintf("%s %s,", typeName, argName)
			items = findItem(items, typeName, fieldType.Type)
		}
	}
	item += "]"
	return append(items, item)
}

func writeWithTemplate(wr io.Writer, text string, data interface{}) {
	tmpl, err := template.New("").Parse(text)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(wr, data)
	if err != nil {
		panic(err)
	}
}

func writeFileWithTemplate(filePath, text string, data interface{}) {
	f, err := os.Create(filePath)
	if err != nil {
		panic("fail to create file:" + filePath)
	}
	writeWithTemplate(f, text, data)
}

type fieldsYamlValues struct {
	List string
}

var fieldsYaml string = ` + "`" + `thrift-fieldmapping:
{{printf "%s" "{{.List}}"}}
` + "`" + `

func buildParameterStr(serviceName, methodName string) string { {{range $ServiceName, $Methods := .ServiceMethodMap}}
	if serviceName == "{{- $ServiceName -}}" {
		switch methodName { {{range $i, $MethodName := $Methods}}
		case "{{$MethodName}}":
			var result string
			args := g.{{- $ServiceName -}}{{$MethodName}}Args{}
			at := reflect.TypeOf(args)
			num := at.NumField()
			for i := 0; i < num; i++ {
				result += fmt.Sprintf(
					"\n\t\t\t\tparams[%d].Interface().(%s),",
					i, at.Field(i).Type.String())
			}
			return result{{end}}
		default:
			return "error"
		}
	}{{end}}
	return "error"
}
`

// GenerateThriftSwitcher generates "thriftswitcher.go"
func (g *Generator) GenerateThriftSwitcher() {
	if _, err := os.Stat(g.c.ServiceRootPathAbsolute() + "/gen"); os.IsNotExist(err) {
		os.Mkdir(g.c.ServiceRootPathAbsolute()+"/gen", 0755)
	}
	serviceMethodMap := methodNames(g.c.mappings[urlServiceMaps])
	parameters := make(map[string][]string, len(serviceMethodMap))
	notEmptyParameters := make(map[string][]bool, len(serviceMethodMap))
	for s, methods := range serviceMethodMap {
		for _, v := range methods {
			p := g.thriftParameters(s, v)
			parameters[s] = append(parameters[s], p)
			notEmptyParameters[s] = append(notEmptyParameters[s], len(strings.TrimSpace(p)) > 0)
		}
	}
	var argCasesStr string
	fields := make([]string, 0, len(g.c.fieldMappings))
	structNames := make([]string, 0, len(g.c.fieldMappings))
	for k := range g.c.fieldMappings {
		structNames = append(structNames, k)
		fields = append(fields, g.structFields(k))
	}
	writeFileWithTemplate(
		g.c.ServiceRootPathAbsolute()+"/gen/thriftswitcher.go",
		struct {
			PkgPath            string
			BuildArgsCases     string
			ServiceNames       []string
			ServiceMethodMap   map[string][]string
			Parameters         map[string][]string
			NotEmptyParameters map[string][]bool
			StructNames        []string
			StructFields       []string
		}{
			g.PkgPath,
			argCasesStr,
			g.c.ThriftServiceNames(),
			serviceMethodMap,
			parameters,
			notEmptyParameters,
			structNames,
			fields},
		thriftSwitcherFunc,
	)
}

func (g *Generator) thriftParameters(serviceName, methodName string) string {
	cmd := "go run " + g.c.ServiceRootPathAbsolute() + "/gen/thrift/build.go -n " + serviceName + "," + methodName
	buf := &bytes.Buffer{}
	c := exec.Command("bash", "-c", cmd)
	c.Stdin = os.Stdin
	c.Stderr = os.Stderr
	c.Stdout = buf
	panicIf(c.Run())
	return buf.String() + " "
}

func methodNames(urlServiceMaps [][4]string) map[string][]string {
	methodNamesMap := make(map[string]map[string]int)
	for _, v := range urlServiceMaps {
		if methodNamesMap[v[2]] == nil {
			methodNamesMap[v[2]] = make(map[string]int)
		}
		methodNamesMap[v[2]][v[3]] = 0
	}
	methodNames := make(map[string][]string)
	for k, v := range methodNamesMap {
		methods := make([]string, 0)
		for m := range v {
			methods = append(methods, m)
		}
		methodNames[k] = methods
	}
	return methodNames
}

var thriftSwitcherFunc = `// Code generated by turbo. DO NOT EDIT.
package gen

import (
	"{{.PkgPath}}/gen/thrift/gen-go/gen"
	"github.com/vaporz/turbo"
	"reflect"
	"net/http"
	"errors"
)

// ThriftSwitcher is a runtime func with which a server starts.
var ThriftSwitcher = func(s turbo.Servable, serviceName, methodName string, resp http.ResponseWriter, req *http.Request) (serviceResponse interface{}, err error) { {{range $Service, $Methods := .ServiceMethodMap}}
	if serviceName == "{{$Service}}" {
		switch methodName { {{range $i, $MethodName := $Methods}}
		case "{{$MethodName}}":{{if index $.NotEmptyParameters $Service $i }}
			params, err := turbo.BuildThriftRequest(s, gen.{{$Service}}{{$MethodName}}Args{}, req, buildStructArg)
			if err != nil {
				return nil, err
			}{{end}}
			return s.Service("{{$Service}}").(*gen.{{$Service}}Client).{{$MethodName}}({{index $.Parameters $Service $i}}){{end}}
		default:
			return nil, errors.New("No such method[" + methodName + "]")
		}
	}
	{{end}}
	if serviceResponse == nil && err == nil {
		return nil, errors.New("No such service[" + serviceName + "]")
	}
	return
}

func buildStructArg(s turbo.Servable, typeName string, req *http.Request) (v reflect.Value, err error) {
	switch typeName {
{{range $i, $StructName := .StructNames}}
	case "{{$StructName}}":
		request := &gen.{{$StructName}}{ {{index $.StructFields $i}} }
		turbo.BuildStruct(s, reflect.TypeOf(request).Elem(), reflect.ValueOf(request).Elem(), req)
		return reflect.ValueOf(request), nil
{{end}}
	default:
		return v, errors.New("unknown typeName[" + typeName + "]")
	}
}
`

// GenerateThriftStub generates Thrift stub codes
func (g *Generator) GenerateThriftStub() {
	if _, err := os.Stat(g.c.ServiceRootPathAbsolute() + "/gen/thrift"); os.IsNotExist(err) {
		os.MkdirAll(g.c.ServiceRootPathAbsolute()+"/gen/thrift", 0755)
	}
	nameLower := strings.ToLower(g.c.ThriftServiceNames()[0]) // todo change a thrift file name
	cmd := "thrift " + g.Options + " -r --gen go:package_prefix=" + g.PkgPath + "/gen/thrift/gen-go/ -o" +
		" " + g.c.ServiceRootPathAbsolute() + "/" + "gen/thrift " + g.c.ServiceRootPathAbsolute() + "/" + nameLower + ".thrift"
	executeCmd("bash", "-c", cmd)
}

func executeCmd(cmd string, args ...string) {
	c := exec.Command(cmd, args...)
	c.Stdin = os.Stdin
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	panicIf(c.Run())
}
