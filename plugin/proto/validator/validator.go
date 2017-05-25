package validator

import (
	"io/ioutil"
	"os"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/plugin"
	"log"
	"reflect"
	"text/template"
	"turbo/plugin/proto/validator"
	"fmt"
)

func main() {
	request := new(plugin_go.CodeGeneratorRequest)
	response := new(plugin_go.CodeGeneratorResponse)
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Println("reading input error:", err)
	}
	if err = proto.Unmarshal(data, request); err != nil {
		log.Println("parsing input proto:", err)
	}
	generateValidator(request, response)
	reflect.String.String()
}

func generateValidator(req *plugin_go.CodeGeneratorRequest, resp *plugin_go.CodeGeneratorResponse) {
	// 得到文件列表
	files := req.ProtoFile
	// 对每个文件循环
	for _, f := range files {
		// 得到message列表
		messages := f.MessageType
		// 对每个message循环
		for _, m := range messages {
			// 得到field列表
			fields := m.Field
			// 对每个field循环
			for _, field := range fields {
				// 得到declaredOption列表
				declaredOptions, err := proto.ExtensionDescs(field.Options)
				if err != nil {

				}
				// 得到field的kind，得到kind对应的合法option列表
				kind, ok := FieldDescriptorProto_Type_Kind[int32(*field.Type)]
				if !ok {
					panic("unknown kind:" + kind.String())
				}
				checkedOptions := findOptions(kind)
				// 对declaredOption中的每个option循环
				if declaredOptions != nil {
					for _, option := range declaredOptions {
						// 如果不在合法option列表中，panic
						checkOption(option, checkedOptions)
						// 否则执行模版，生成代码

						ext, err := proto.GetExtension(declaredOptions, checkedOptions)
						if err != nil {
							fmt.Println(err.Error())
							continue
						}

					}
				}
			}
		}

	}
	// todo 重复定义多个会panic

	validations := ""

	tmpl, err := template.New("stringSetter").Parse(stringSetter)
	if err != nil {
		panic(err)
	}
	f := os.Stdout
	err = tmpl.Execute(f, setterValues{
		ValueKind:   reflect.String.String(),
		Validations: validations})
	if err != nil {
		panic(err)
	}
}

func checkOption(option *proto.ExtensionDesc, validOptions []*proto.ExtensionDesc) {
	for _, op := range validOptions {
		if option.Field == op.Field {
			return
		}
	}
	panic("wrong option")
}

func findOptions(kind reflect.Kind) []*proto.ExtensionDesc {
	switch kind {
	case reflect.String:
		return []*proto.ExtensionDesc{
			validator.E_MaxLength,
			validator.E_NotBlank}
	case reflect.Int:
	case reflect.Int32:

	case reflect.Int64:
	default:
		return []*proto.ExtensionDesc{}
	}
}

var FieldDescriptorProto_Type_Kind = map[int32]reflect.Kind{
	1: reflect.Float64,
	2: reflect.Float32,
	3: reflect.Int64,
	4: reflect.Uint64,
	5: reflect.Int32,
	6: reflect.Uint64,
	7: reflect.Uint32,
	8: reflect.Bool,
	9: reflect.String,
	//10: "TYPE_GROUP",
	//11: "TYPE_MESSAGE",
	//12: "TYPE_BYTES",
	13: reflect.Uint32,
	//14: "TYPE_ENUM",
	15: reflect.Uint32,
	16: reflect.Uint64,
	17: reflect.Int32,
	18: reflect.Int64,
}

type setterValues struct {
	ValueKind       string
	RequestTypeName string
	FieldName       string
	Validations     string
}

var stringSetter = `func (r *{{.RequestTypeName}}) Set{{.FieldName}}(v {{.ValueKind}}) error {
	{{.Validations}}
	r.YourName = v
	return nil
}
`

type stringNotBlankValues struct {
	FieldName string
}

var stringNotBlank = `
	// NotBlank
	if len(v) == 0 {
		return errors.New("validate fail! NotBlank, [{{.FieldName}}] is blank")
	}`

type stringMaxLengthValues struct {
	MaxLength int
}

var stringMaxLength = `
	// MaxLength
	if len(v) > {{.MaxLength}} {
		return errors.New("validate fail! MaxLength, expected:<={{.MaxLength}}, actual:" + strconv.Itoa(len(v)))
	}`
