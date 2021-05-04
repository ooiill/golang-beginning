package acme

import (
    "bytes"
    "fmt"
    "strconv"
)

func Serialize(value interface{}) (result string, err error) {
    buf := new(bytes.Buffer)
    err = encodeValue(buf, value)
    if err == nil {
        result = buf.String()
    }
    return
}

const TypeValueSeparator = ':'
const ValuesSeparator = ';'

type PhpObject struct {
    members   map[interface{}]interface{}
    className string
}

func NewPhpObject() *PhpObject {
    d := &PhpObject{
        members: make(map[interface{}]interface{}),
    }
    return d
}

func (obj *PhpObject) GetClassName() string {
    return obj.className
}

func (obj *PhpObject) SetClassName(cName string) {
    obj.className = cName
}

func (obj *PhpObject) GetMembers() map[interface{}]interface{} {
    return obj.members
}

func (obj *PhpObject) GetPrivateMemberValue(memberName string) (interface{}, bool) {
    key := fmt.Sprintf("\x00%s\x00%s", obj.className, memberName)
    v, ok := obj.members[key]
    return v, ok
}

func (obj *PhpObject) SetPrivateMemberValue(memberName string, value interface{}) {
    key := fmt.Sprintf("\x00%s\x00%s", obj.className, memberName)
    obj.members[key] = value
}

func (obj *PhpObject) GetProtectedMemberValue(memberName string) (interface{}, bool) {
    key := "\x00*\x00" + memberName
    v, ok := obj.members[key]
    return v, ok
}

func (obj *PhpObject) SetProtectedMemberValue(memberName string, value interface{}) {
    key := "\x00*\x00" + memberName
    obj.members[key] = value
}

func (obj *PhpObject) GetPublicMemberValue(memberName string) (interface{}, bool) {
    v, ok := obj.members[memberName]
    return v, ok
}

func (obj *PhpObject) SetPublicMemberValue(memberName string, value interface{}) {
    obj.members[memberName] = value
}

func encodeValue(buf *bytes.Buffer, value interface{}) (err error) {
    switch t := value.(type) {
    default:
        err = fmt.Errorf("Unexpected type %T", t)
    case bool:
        buf.WriteString("b")
        buf.WriteRune(TypeValueSeparator)
        if t {
            buf.WriteString("1")
        } else {
            buf.WriteString("0")
        }
        buf.WriteRune(ValuesSeparator)
    case nil:
        buf.WriteString("N")
        buf.WriteRune(ValuesSeparator)
    case int, int64, int32, int16, int8:
        buf.WriteString("i")
        buf.WriteRune(TypeValueSeparator)
        strValue := fmt.Sprintf("%v", t)
        buf.WriteString(strValue)
        buf.WriteRune(ValuesSeparator)
    case float32:
        buf.WriteString("d")
        buf.WriteRune(TypeValueSeparator)
        strValue := strconv.FormatFloat(float64(t), 'f', -1, 64)
        buf.WriteString(strValue)
        buf.WriteRune(ValuesSeparator)
    case float64:
        buf.WriteString("d")
        buf.WriteRune(TypeValueSeparator)
        strValue := strconv.FormatFloat(float64(t), 'f', -1, 64)
        buf.WriteString(strValue)
        buf.WriteRune(ValuesSeparator)
    case string:
        buf.WriteString("s")
        buf.WriteRune(TypeValueSeparator)
        encodeString(buf, t)
        buf.WriteRune(ValuesSeparator)
    case map[interface{}]interface{}:
        buf.WriteString("a")
        buf.WriteRune(TypeValueSeparator)
        err = encodeArrayCore(buf, t)
    case *PhpObject:
        buf.WriteString("O")
        buf.WriteRune(TypeValueSeparator)
        encodeString(buf, t.GetClassName())
        buf.WriteRune(TypeValueSeparator)
        err = encodeArrayCore(buf, t.GetMembers())
    }
    return
}

func encodeString(buf *bytes.Buffer, strValue string) {
    valLen := strconv.Itoa(len(strValue))
    buf.WriteString(valLen)
    buf.WriteRune(TypeValueSeparator)
    buf.WriteRune('"')
    buf.WriteString(strValue)
    buf.WriteRune('"')
}

func encodeArrayCore(buf *bytes.Buffer, arrValue map[interface{}]interface{}) (err error) {
    valLen := strconv.Itoa(len(arrValue))
    buf.WriteString(valLen)
    buf.WriteRune(TypeValueSeparator)

    buf.WriteRune('{')
    for k, v := range arrValue {
        if intKey, _err := strconv.Atoi(fmt.Sprintf("%v", k)); _err == nil {
            if err = encodeValue(buf, intKey); err != nil {
                break
            }
        } else {
            if err = encodeValue(buf, k); err != nil {
                break
            }
        }
        if err = encodeValue(buf, v); err != nil {
            break
        }
    }
    buf.WriteRune('}')
    return err
}
