package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
)

var RawJSON = `
[
	{
		"type": "function",
		"name": "transfer",
		"inputs": [{"name": "to", "type": "address"}, {"name": "value", "type": "uint256"}],
		"stateMutability": "payable",
		"payable": true
	},
	{
		"type": "function",
		"name": "balanceOf",
		"inputs": [{"name": "who", "type": "address"}],
		"outputs": [{"type": "uint256"}],
		"stateMutability": "view"
	}
]
`
var innerAbiTool = new(abiTool).init()

type TransferParameter struct {
	ToAddress string
	Amount    int64
}

type abiTool struct {
	abi abi.ABI
}

func (a *abiTool) encodeInputs(methodName string, inputs ...interface{}) (parameterStr string, err error) {
	packed, err := a.abi.Methods[methodName].Inputs.Pack(inputs...)
	if err != nil {
		return
	}
	parameterStr = hex.EncodeToString(packed)
	return
}

// init ...
func (a *abiTool) init() *abiTool {
	var err error
	a.abi, err = abi.JSON(strings.NewReader(RawJSON))
	if err != nil {
		log.Fatalln(err)
	}
	return a
}

func AnyToString(i interface{}) string {
	i = indirectToStringerOrError(i)
	switch s := i.(type) {
	case string:
		return s
	case bool:
		return strconv.FormatBool(s)
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(s), 'f', -1, 32)
	case int:
		return strconv.Itoa(s)
	case int64:
		return strconv.FormatInt(s, 10)
	case int32:
		return strconv.Itoa(int(s))
	case int16:
		return strconv.FormatInt(int64(s), 10)
	case int8:
		return strconv.FormatInt(int64(s), 10)
	case uint:
		return strconv.FormatUint(uint64(s), 10)
	case uint64:
		return strconv.FormatUint(uint64(s), 10)
	case uint32:
		return strconv.FormatUint(uint64(s), 10)
	case uint16:
		return strconv.FormatUint(uint64(s), 10)
	case uint8:
		return strconv.FormatUint(uint64(s), 10)
	case json.Number:
		return s.String()
	case []byte:
		return string(s)
	case template.HTML:
		return string(s)
	case template.URL:
		return string(s)
	case template.JS:
		return string(s)
	case template.CSS:
		return string(s)
	case template.HTMLAttr:
		return string(s)
	case nil:
		return ""
	case fmt.Stringer:
		return s.String()
	case error:
		return s.Error()
	default:
		jsonBytes, _ := json.Marshal(i)
		return string(jsonBytes)
	}
}

func Trc20ParameterEncodeTransfer(in TransferParameter) (parameterStr string, err error) {
	addr, err := address.Base58ToAddress(in.ToAddress)
	if err != nil {
		return
	}
	hexAddr := strings.Replace(addr.Hex(), "0x", "", 1)
	return innerAbiTool.encodeInputs("transfer", common.HexToAddress(hexAddr), big.NewInt(in.Amount))
}

func indirectToStringerOrError(a interface{}) interface{} {
	if a == nil {
		return nil
	}

	var errorType = reflect.TypeOf((*error)(nil)).Elem()
	var fmtStringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

	v := reflect.ValueOf(a)
	for !v.Type().Implements(fmtStringerType) && !v.Type().Implements(errorType) && v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}
