package gentool

import (
	"fmt"
	"reflect"
	"strings"
)

// GenerateObjectProxy !
type GenerateObjectProxy struct {
	rpcservicePkname string
	packageName      string
	objectName       string
	filePath         string
}

// NewGenerateObjectProxy !!
func NewGenerateObjectProxy(rpcservicePkname, packageName, objectName, filePath string) *GenerateObjectProxy {
	return &GenerateObjectProxy{
		rpcservicePkname: rpcservicePkname,
		packageName:      packageName,
		objectName:       objectName,
		filePath:         filePath,
	}
}

func (me *GenerateObjectProxy) parseParam(t reflect.Type) string {
	var strSubt string

	switch t.Kind() {
	case reflect.Array:
	case reflect.Slice:
		strSubt = me.parseArrayParam(t)
	case reflect.Map:
		strSubt = me.parseMapParam(t)
	case reflect.Chan:
		strSubt = me.parseChanParam(t)
	case reflect.Ptr:
		strSubt = me.parsePtrParam(t)
	case reflect.Func:
		strSubt, _, _ = me.parseFuncParam(false, "", t)
	case reflect.Interface:
		strSubt = t.Name()
		if strSubt == "" {
			strSubt = "interface{}"
		}
	default:
		strSubt = t.Name()
	}

	return strSubt
}

func (me *GenerateObjectProxy) parseFuncParam(isInstanceFun bool, funcName string, t reflect.Type) (string, []string, []string) {
	if t.Kind() != reflect.Func {
		return "", nil, nil
	}

	isVariadic := t.IsVariadic()

	paramCount := t.NumIn()
	paramSpan := 0
	if isInstanceFun {
		paramCount--
		paramSpan = 1
	}

	strIns := make([]string, paramCount)
	arguNames := make([]string, paramCount)
	strOuts := make([]string, t.NumOut())

	strInParams := ""
	for i := 0; i < paramCount; i++ {
		inType := t.In(i + paramSpan)

		if i == paramCount-1 && isVariadic {
			typeTmp := me.parseParam(inType)
			if strings.HasPrefix(typeTmp, "[]") {
				strIns[i] = "..." + typeTmp[2:]
			} else {
				strIns[i] = "..." + typeTmp
			}
		} else {
			strIns[i] = me.parseParam(inType)
		}

		arguNames[i] = fmt.Sprintf("v%v", i)
		strInParams = strInParams + arguNames[i] + " " + strIns[i]
		if i < paramCount-1 {
			strInParams = strInParams + ","
		}
	}

	strOutParams := ""
	for i := 0; i < t.NumOut(); i++ {
		outType := t.Out(i)
		strOuts[i] = me.parseParam(outType)
		strOutParams = strOutParams + fmt.Sprintf(" r%v ", i) + strOuts[i]
		if i < t.NumOut()-1 {
			strOutParams = strOutParams + ","
		}
	}

	return fmt.Sprintf("func %v(%v) (%v)", funcName, strInParams, strOutParams), arguNames, strOuts
}

func (me *GenerateObjectProxy) parsePtrParam(t reflect.Type) string {
	if t.Kind() != reflect.Ptr {
		return ""
	}

	valueTypeName := me.parseParam(t.Elem())
	if valueTypeName == "" {
		return ""
	}

	return "*" + valueTypeName
}

func (me *GenerateObjectProxy) parseChanParam(t reflect.Type) string {
	if t.Kind() != reflect.Chan {
		return ""
	}

	chanDir := t.ChanDir()
	strChanDir := "chan"
	switch chanDir {
	case reflect.RecvDir:
		strChanDir = "<-chan"
	case reflect.SendDir:
		strChanDir = "chan<-"
	}

	valueTypeName := me.parseParam(t.Elem())
	if valueTypeName == "" {
		return ""
	}

	return strChanDir + " " + valueTypeName
}

func (me *GenerateObjectProxy) parseMapParam(t reflect.Type) string {
	if t.Kind() != reflect.Map {
		return ""
	}

	keyType := t.Key()
	keyTypeName := me.parseParam(keyType)
	if keyTypeName == "" {
		return ""
	}

	valueTypeName := me.parseParam(t.Elem())
	if valueTypeName == "" {
		return ""
	}

	return "map[" + keyTypeName + "]" + valueTypeName
}

func (me *GenerateObjectProxy) parseArrayParam(t reflect.Type) string {
	if t.Kind() != reflect.Array && t.Kind() != reflect.Slice {
		return ""
	}

	typeName := me.parseParam(t.Elem())
	if typeName == "" {
		return ""
	}

	return "[]" + typeName
}

func (me *GenerateObjectProxy) parseExportFuncObjectHelper(f reflect.Method, funcName string, needRet bool) string {
	funcNamePostfix := ""
	if !needRet {
		funcNamePostfix = "WithoutReturn"
	}

	fullFuncName, InParamNames, OutParamTypes := me.parseFuncParam(true, funcName+funcNamePostfix, f.Type)

	callArgus := ""
	for i, arg := range InParamNames {
		callArgus = callArgus + arg
		if i < len(InParamNames)-1 {
			callArgus = callArgus + ","
		}
	}

	retStms := ""
	for i, retType := range OutParamTypes {
		retStms = retStms + fmt.Sprintf("if req.Ret[%v] != nil { r%v = req.Ret[%v].(%v)	}\r\n", i, i, i, retType)
	}

	strfmt := `
// %v Desc:
%v{
	needRet := %v
	req:= %v.NewRPCAction("%v", "%v", needRet, %v)
	if me.objIndex != -1 {
		me.rs.PushRPCAction(req.SetObjectIndex(me.objIndex))
		me.objIndex = -1		
	}else{
		me.rs.PushRPCAction(req)
	}
	me.rs.ConnectPoint <- %v.Action_RemoteCall
	if !needRet {
		return 
	}

	<-req.RetChannel
	close(req.RetChannel)
	if req.RetError != nil {
		oblog.Errorf("RPC call %v failed! err:%%v", req.RetError)
		return
	}
	if len(req.Ret) != %v{
		oblog.Errorf("returned values from RPC call %v service is invalid!")
		return
	}

	%v
	return
}

	`

	return fmt.Sprintf(strfmt, f.Name+funcNamePostfix, fullFuncName, needRet, me.rpcservicePkname, me.objectName, f.Name, callArgus, me.rpcservicePkname, funcName+funcNamePostfix, len(OutParamTypes), funcName+funcNamePostfix, retStms)
}

func (me *GenerateObjectProxy) parseExportFuncObject(f reflect.Method) string {
	funcName := fmt.Sprintf("(me *%vProxy) %v", me.objectName, f.Name)

	strret := me.parseExportFuncObjectHelper(f, funcName, true)
	strret = strret + me.parseExportFuncObjectHelper(f, funcName, false)

	return strret
}

func (me *GenerateObjectProxy) generateHeaders() string {
	strfmt := `
import (
	"github.com/luolingo/object-service-bridge/oblog"
	"github.com/luolingo/object-service-bridge/%v"
)

// %vProxy !
type %vProxy struct {
	rs *%v.RPCServiceExt
	objIndex int
}
	
// New%vProxy !
func New%vProxy(rs *%v.RPCServiceExt) *%vProxy {
	return &%vProxy{
		rs: rs,
		objIndex: -1,
	}
}

func (me *%vProxy) Object(index int) *%vProxy {
	if index >= 0 {
		me.objIndex = index
	}
	return me
}

`

	return fmt.Sprintf(strfmt, me.rpcservicePkname, me.objectName, me.objectName, me.rpcservicePkname, me.objectName,
		me.objectName, me.rpcservicePkname, me.objectName, me.objectName,
		me.objectName, me.objectName)
}

// GenerateProxyObject !!
func (me *GenerateObjectProxy) GenerateProxyObject(obj interface{}) {
	proxyObj := reflect.TypeOf(obj)

	proxyHeaders := me.generateHeaders()

	proxyStms := ""
	for i := 0; i < proxyObj.NumMethod(); i++ {
		exportMethod := proxyObj.Method(i)
		proxyStms = proxyStms + me.parseExportFuncObject(exportMethod)
	}

	fmt.Printf("%v\r\n%v\r\n", proxyHeaders, proxyStms)
}
