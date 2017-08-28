// Autogenerated by Thrift Compiler (0.10.0)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

package main

import (
        "flag"
        "fmt"
        "math"
        "net"
        "net/url"
        "os"
        "strconv"
        "strings"
        "git.apache.org/thrift.git/lib/go/thrift"
        "github.com/vaporz/turbo/test/testservice/gen/thrift/gen-go/gen"
)


func Usage() {
  fmt.Fprintln(os.Stderr, "Usage of ", os.Args[0], " [-h host:port] [-u url] [-f[ramed]] function [arg1 [arg2...]]:")
  flag.PrintDefaults()
  fmt.Fprintln(os.Stderr, "\nFunctions:")
  fmt.Fprintln(os.Stderr, "  SayHelloResponse sayHello(CommonValues values, string yourName, i64 int64Value, bool boolValue, double float64Value, i64 uint64Value, i32 int32Value, i16 int16Value,  stringList,  i32List,  boolList)")
  fmt.Fprintln(os.Stderr, "  TestJsonResponse testJson(TestJsonRequest request)")
  fmt.Fprintln(os.Stderr)
  os.Exit(0)
}

func main() {
  flag.Usage = Usage
  var host string
  var port int
  var protocol string
  var urlString string
  var framed bool
  var useHttp bool
  var parsedUrl url.URL
  var trans thrift.TTransport
  _ = strconv.Atoi
  _ = math.Abs
  flag.Usage = Usage
  flag.StringVar(&host, "h", "localhost", "Specify host and port")
  flag.IntVar(&port, "p", 9090, "Specify port")
  flag.StringVar(&protocol, "P", "binary", "Specify the protocol (binary, compact, simplejson, json)")
  flag.StringVar(&urlString, "u", "", "Specify the url")
  flag.BoolVar(&framed, "framed", false, "Use framed transport")
  flag.BoolVar(&useHttp, "http", false, "Use http")
  flag.Parse()
  
  if len(urlString) > 0 {
    parsedUrl, err := url.Parse(urlString)
    if err != nil {
      fmt.Fprintln(os.Stderr, "Error parsing URL: ", err)
      flag.Usage()
    }
    host = parsedUrl.Host
    useHttp = len(parsedUrl.Scheme) <= 0 || parsedUrl.Scheme == "http"
  } else if useHttp {
    _, err := url.Parse(fmt.Sprint("http://", host, ":", port))
    if err != nil {
      fmt.Fprintln(os.Stderr, "Error parsing URL: ", err)
      flag.Usage()
    }
  }
  
  cmd := flag.Arg(0)
  var err error
  if useHttp {
    trans, err = thrift.NewTHttpClient(parsedUrl.String())
  } else {
    portStr := fmt.Sprint(port)
    if strings.Contains(host, ":") {
           host, portStr, err = net.SplitHostPort(host)
           if err != nil {
                   fmt.Fprintln(os.Stderr, "error with host:", err)
                   os.Exit(1)
           }
    }
    trans, err = thrift.NewTSocket(net.JoinHostPort(host, portStr))
    if err != nil {
      fmt.Fprintln(os.Stderr, "error resolving address:", err)
      os.Exit(1)
    }
    if framed {
      trans = thrift.NewTFramedTransport(trans)
    }
  }
  if err != nil {
    fmt.Fprintln(os.Stderr, "Error creating transport", err)
    os.Exit(1)
  }
  defer trans.Close()
  var protocolFactory thrift.TProtocolFactory
  switch protocol {
  case "compact":
    protocolFactory = thrift.NewTCompactProtocolFactory()
    break
  case "simplejson":
    protocolFactory = thrift.NewTSimpleJSONProtocolFactory()
    break
  case "json":
    protocolFactory = thrift.NewTJSONProtocolFactory()
    break
  case "binary", "":
    protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
    break
  default:
    fmt.Fprintln(os.Stderr, "Invalid protocol specified: ", protocol)
    Usage()
    os.Exit(1)
  }
  client := gen.NewTestServiceClientFactory(trans, protocolFactory)
  if err := trans.Open(); err != nil {
    fmt.Fprintln(os.Stderr, "Error opening socket to ", host, ":", port, " ", err)
    os.Exit(1)
  }
  
  switch cmd {
  case "sayHello":
    if flag.NArg() - 1 != 11 {
      fmt.Fprintln(os.Stderr, "SayHello requires 11 args")
      flag.Usage()
    }
    arg9 := flag.Arg(1)
    mbTrans10 := thrift.NewTMemoryBufferLen(len(arg9))
    defer mbTrans10.Close()
    _, err11 := mbTrans10.WriteString(arg9)
    if err11 != nil {
      Usage()
      return
    }
    factory12 := thrift.NewTSimpleJSONProtocolFactory()
    jsProt13 := factory12.GetProtocol(mbTrans10)
    argvalue0 := gen.NewCommonValues()
    err14 := argvalue0.Read(jsProt13)
    if err14 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    argvalue2, err16 := (strconv.ParseInt(flag.Arg(3), 10, 64))
    if err16 != nil {
      Usage()
      return
    }
    value2 := argvalue2
    argvalue3 := flag.Arg(4) == "true"
    value3 := argvalue3
    argvalue4, err18 := (strconv.ParseFloat(flag.Arg(5), 64))
    if err18 != nil {
      Usage()
      return
    }
    value4 := argvalue4
    argvalue5, err19 := (strconv.ParseInt(flag.Arg(6), 10, 64))
    if err19 != nil {
      Usage()
      return
    }
    value5 := argvalue5
    tmp6, err20 := (strconv.Atoi(flag.Arg(7)))
    if err20 != nil {
      Usage()
      return
    }
    argvalue6 := int32(tmp6)
    value6 := argvalue6
    tmp7, err21 := (strconv.Atoi(flag.Arg(8)))
    if err21 != nil {
      Usage()
      return
    }
    argvalue7 := int16(tmp7)
    value7 := argvalue7
    arg22 := flag.Arg(9)
    mbTrans23 := thrift.NewTMemoryBufferLen(len(arg22))
    defer mbTrans23.Close()
    _, err24 := mbTrans23.WriteString(arg22)
    if err24 != nil { 
      Usage()
      return
    }
    factory25 := thrift.NewTSimpleJSONProtocolFactory()
    jsProt26 := factory25.GetProtocol(mbTrans23)
    containerStruct8 := gen.NewTestServiceSayHelloArgs()
    err27 := containerStruct8.ReadField9(jsProt26)
    if err27 != nil {
      Usage()
      return
    }
    argvalue8 := containerStruct8.StringList
    value8 := argvalue8
    arg28 := flag.Arg(10)
    mbTrans29 := thrift.NewTMemoryBufferLen(len(arg28))
    defer mbTrans29.Close()
    _, err30 := mbTrans29.WriteString(arg28)
    if err30 != nil { 
      Usage()
      return
    }
    factory31 := thrift.NewTSimpleJSONProtocolFactory()
    jsProt32 := factory31.GetProtocol(mbTrans29)
    containerStruct9 := gen.NewTestServiceSayHelloArgs()
    err33 := containerStruct9.ReadField10(jsProt32)
    if err33 != nil {
      Usage()
      return
    }
    argvalue9 := containerStruct9.I32List
    value9 := argvalue9
    arg34 := flag.Arg(11)
    mbTrans35 := thrift.NewTMemoryBufferLen(len(arg34))
    defer mbTrans35.Close()
    _, err36 := mbTrans35.WriteString(arg34)
    if err36 != nil { 
      Usage()
      return
    }
    factory37 := thrift.NewTSimpleJSONProtocolFactory()
    jsProt38 := factory37.GetProtocol(mbTrans35)
    containerStruct10 := gen.NewTestServiceSayHelloArgs()
    err39 := containerStruct10.ReadField11(jsProt38)
    if err39 != nil {
      Usage()
      return
    }
    argvalue10 := containerStruct10.BoolList
    value10 := argvalue10
    fmt.Print(client.SayHello(value0, value1, value2, value3, value4, value5, value6, value7, value8, value9, value10))
    fmt.Print("\n")
    break
  case "testJson":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "TestJson requires 1 args")
      flag.Usage()
    }
    arg40 := flag.Arg(1)
    mbTrans41 := thrift.NewTMemoryBufferLen(len(arg40))
    defer mbTrans41.Close()
    _, err42 := mbTrans41.WriteString(arg40)
    if err42 != nil {
      Usage()
      return
    }
    factory43 := thrift.NewTSimpleJSONProtocolFactory()
    jsProt44 := factory43.GetProtocol(mbTrans41)
    argvalue0 := gen.NewTestJsonRequest()
    err45 := argvalue0.Read(jsProt44)
    if err45 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.TestJson(value0))
    fmt.Print("\n")
    break
  case "":
    Usage()
    break
  default:
    fmt.Fprintln(os.Stderr, "Invalid function ", cmd)
  }
}
