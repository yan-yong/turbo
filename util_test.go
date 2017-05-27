package turbo

import (
	sjson "github.com/bitly/go-simplejson"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type args struct {
}

type testStruct struct {
	TestId   int64
	PtrValue *args
}

func TestFilterFieldInt64Str(t *testing.T) {
	s := &testStruct{TestId: 123}
	tp := reflect.TypeOf(s).Elem()
	v := reflect.ValueOf(s).Elem()
	json, _ := sjson.NewJson([]byte("{\"test_id\": \"123\"}"))
	structFieldFilter(tp.Field(0))(json, tp.Field(0), v.Field(0))
	jsonBytes, _ := json.MarshalJSON()

	assert.Equal(t, "{\"test_id\":123}", string(jsonBytes))
}

func TestFilterFieldInt64Number(t *testing.T) {
	s := &testStruct{TestId: 123}
	json, _ := sjson.NewJson([]byte("{\"test_id\": 123}"))
	filterStruct(json, reflect.TypeOf(s).Elem(), reflect.ValueOf(s).Elem())
	jsonBytes, _ := json.MarshalJSON()

	assert.Equal(t, "{\"ptr_value\":null,\"test_id\":123}", string(jsonBytes))
}

func TestFilterFieldNullPointer(t *testing.T) {
	s := &testStruct{TestId: 123}
	tp := reflect.TypeOf(s).Elem()
	v := reflect.ValueOf(s).Elem()
	json, _ := sjson.NewJson([]byte("{\"test_id\": \"123\"}"))
	structFieldFilter(tp.Field(1))(json, tp.Field(1), v.Field(1))
	jsonBytes, _ := json.MarshalJSON()

	assert.Equal(t, "{\"ptr_value\":null,\"test_id\":\"123\"}", string(jsonBytes))
}

func TestFilterStruct(t *testing.T) {
	s := &testStruct{TestId: 123}
	json, _ := sjson.NewJson([]byte("{\"test_id\": \"123\"}"))
	filterStruct(json, reflect.TypeOf(s).Elem(), reflect.ValueOf(s).Elem())
	jsonBytes, _ := json.MarshalJSON()

	assert.Equal(t, "{\"ptr_value\":null,\"test_id\":123}", string(jsonBytes))
}

func TestFilterStruct_Missing_Key(t *testing.T) {
	s := &testStruct{TestId: 123}
	json, _ := sjson.NewJson([]byte("{}"))
	filterStruct(json, reflect.TypeOf(s).Elem(), reflect.ValueOf(s).Elem())
	jsonBytes, _ := json.MarshalJSON()

	assert.Equal(t, "{\"ptr_value\":null,\"test_id\":123}", string(jsonBytes))
}

type nestedValue struct {
	PtrValue *args
}

type nestedStruct struct {
	TestId      int64
	NestedValue *nestedValue
}

func TestFilterNestedStruct_Nil_field(t *testing.T) {
	s := &nestedStruct{TestId: 123}
	json, _ := sjson.NewJson([]byte("{\"test_id\": \"123\"}"))
	filterStruct(json, reflect.TypeOf(s).Elem(), reflect.ValueOf(s).Elem())
	jsonBytes, _ := json.MarshalJSON()
	assert.Equal(t, "{\"nested_value\":null,\"test_id\":123}", string(jsonBytes))
}

func TestFilterNestedStructField_Empty_Field(t *testing.T) {
	s := &nestedStruct{TestId: 123, NestedValue: &nestedValue{}}
	json, _ := sjson.NewJson([]byte("{\"test_id\": \"123\", \"nested_value\":{}}"))
	structField := reflect.TypeOf(s).Elem().Field(1)
	structFieldFilter(structField)(json, structField, reflect.ValueOf(s).Elem().Field(1))
	jsonBytes, _ := json.MarshalJSON()
	assert.Equal(t, "{\"nested_value\":{\"ptr_value\":null},\"test_id\":\"123\"}", string(jsonBytes))
}

func TestFilterNestedStruct_Empty_Field(t *testing.T) {
	s := &nestedStruct{TestId: 123, NestedValue: &nestedValue{}}
	json, _ := sjson.NewJson([]byte("{\"test_id\": \"123\", \"nested_value\":{}}"))
	filterStruct(json, reflect.TypeOf(s).Elem(), reflect.ValueOf(s).Elem())
	jsonBytes, _ := json.MarshalJSON()
	assert.Equal(t, "{\"nested_value\":{\"ptr_value\":null},\"test_id\":123}", string(jsonBytes))
}

type testTag struct {
	value   int `protobuf:"varint,1,opt,name=test_name_proto" json:"id,omitempty"`
	value_1 int `protobuf:"varint,1,opt" json:"test_name_json,omitempty"`
}

func TestLookupNameInProtoTag(t *testing.T) {
	var v testTag
	sf := reflect.TypeOf(v).Field(0)
	name, _ := lookupNameInProtoTag(sf)
	assert.Equal(t, "test_name_proto", name)
}

func TestLookupNameInJsonTag(t *testing.T) {
	var v testTag
	sf := reflect.TypeOf(v).Field(1)
	name, _ := lookupNameInJsonTag(sf)
	assert.Equal(t, "test_name_json", name)
}

func TestLookupNameInTag(t *testing.T) {
	var v testTag
	sf := reflect.TypeOf(v).Field(0)
	name, _ := lookupNameInTag(sf)
	assert.Equal(t, "test_name_proto", name)

	sf = reflect.TypeOf(v).Field(1)
	name, _ = lookupNameInTag(sf)
	assert.Equal(t, "test_name_json", name)
}

type someArgs struct {
}

type childValue struct {
	TestId      int64
	StringValue string
	IntArray    []int64
	Args        *someArgs
}

type complexNestedValue struct {
	TestId        int64
	StringValue   string
	IntArray      []int64
	ChildValueArr []*childValue
	ChildValue1   *childValue
}

type complexNestedStruct struct {
	TestId              int64
	StringValue         string  `protobuf:"varint,1,opt,name=s_value" json:"id,omitempty"`
	IntArray            []int64 `protobuf:"varint,1,opt,name=new_name" json:"id,omitempty"`
	ComplexNestedValue  *complexNestedValue
	ComplexNestedValue1 *complexNestedValue `protobuf:"varint,1,opt,name=c_n_v1" json:"c_n_v111,omitempty"`
	ComplexNestedValue2 *complexNestedValue `protobuf:"varint,1,opt" json:"c_n_v2,omitempty"`
}

func TestFilterComplexNestedStructWithTags(t *testing.T) {
	cv := &childValue{TestId: 123, StringValue: "a string"}
	cv1 := &childValue{TestId: 456, Args: &someArgs{}}
	cv2 := &childValue{TestId: 789, IntArray: []int64{44, 55, 66}}
	cnv := &complexNestedValue{TestId: 456, IntArray: []int64{11, 22, 33}, ChildValueArr: []*childValue{cv1, cv2}, ChildValue1: cv}
	s := &complexNestedStruct{StringValue: "struct string", ComplexNestedValue: cnv}

	bytes := []byte("{\"s_value\":\"struct string\", \"complex_nested_value\":{\"test_id\":\"456\"" +
		", \"int_array\":[\"11\",\"22\",\"33\"], \"child_value_arr\":[{\"test_id\":\"456\",\"args\":{}}," +
		"{\"test_id\":\"789\",\"int_array\":[\"44\",\"55\",\"66\"]}]" +
		", \"child_value1\":{\"test_id\":\"123\",\"string_value\":\"a string\"}}}")
	bytes, _ = FilterJsonWithStruct(bytes, s)
	assert.Equal(t, "{\"c_n_v1\":null,\"c_n_v2\":null,\"complex_nested_value\":{\"child_value1\":{\"args\":null,\"int_array\":[],\"string_value\":"+
		"\"a string\",\"test_id\":123},\"child_value_arr\":[{\"args\":{},\"int_array\":[],\"string_value\":\"\",\"test_id\":456},"+
		"{\"args\":null,\"int_array\":[44,55,66],\"string_value\":\"\",\"test_id\":789}],\"int_array\":[11,22,33],\"string_value\":\"\""+
		",\"test_id\":456},\"new_name\":[],\"s_value\":\"struct string\","+
		"\"test_id\":0}", string(bytes))
	/*
		Before filter:
		{
		    "string_value": "struct string",
		    "complex_nested_value": {
		        "test_id": "456",
		        "int_array": [
		            "11",
		            "22",
		            "33"
		        ],
		        "child_value_arr": [
		            {
		                "test_id": "456",
		                "Args": {}
		            },
		            {
		                "test_id": "789",
		                "int_array": [
		                    "44",
		                    "55",
		                    "66"
		                ]
		            }
		        ],
		        "child_value1": {
		            "test_id": "123",
		            "string_value": "a string"
		        }
		    }
		}

		After filter:
		{
		    "complex_nested_value": {
		        "child_value1": {
		            "Args": null,
		            "int_array": [],
		            "string_value": "a string",
		            "test_id": 123
		        },
		        "child_value_arr": [
		            {
		                "Args": {},
		                "int_array": [],
		                "string_value": "",
		                "test_id": 456
		            },
		            {
		                "Args": null,
		                "int_array": [
		                    44,
		                    55,
		                    66
		                ],
		                "string_value": "",
		                "test_id": 789
		            }
		        ],
		        "int_array": [
		            11,
		            22,
		            33
		        ],
		        "string_value": "",
		        "test_id": 456
		    },
		    "complex_nested_value1": null,
		    "complex_nested_value2": null,
		    "int_array": [],
		    "string_value": "struct string",
		    "test_id": 0
		}
	*/
}