package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
)

var errorJSONUnmarshal = ErrorParams("JSON解析失败,请检查参数")

type ErrorParams string

func (e ErrorParams) Error() string { return string(e) }

type Response struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data,omitempty"`
	Msg  string          `json:"msg,omitempty"`
}

// JSONHandlerWrap wraps a function to a http.HandlerFunc.
// The function must be a function with one parameter and one return value.
// The parameter must be a struct pointer.
// The return value must be one value and one error or two values.
// The first return value will be encoded to JSON and write to response body.
// e.g.: core.JSONHandlerWrap(func() interface{} { return 1 })
// e.g.: core.JSONHandlerWrap(func() (interface{}, error) { return 1, nil })
func JSONHandlerWrap(call interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reflected := reflect.ValueOf(call)
		if reflected.Kind() != reflect.Func {
			writeErrorResponse(w, errors.New("call must be a function"))
			return
		}
		callArgs := make([]reflect.Value, reflected.Type().NumIn())
		for i := 0; i < reflected.Type().NumIn(); i++ {
			argType := reflected.Type().In(i)
			if argType.Kind() == reflect.Ptr {
				argType = argType.Elem()
			}
			callArgs[i] = reflect.New(argType)
		}
		body, err := io.ReadAll(r.Body)
		defer func() { _ = r.Body.Close() }()
		if err != nil {
			writeErrorResponse(w, err)
			return
		}
		if len(body) > 0 {
			if len(callArgs) != 1 {
				writeErrorResponse(w, errors.New("call must be a function with one parameter"))
				return
			}
			callParam := callArgs[0].Interface()
			if err = json.Unmarshal(body, callParam); err != nil {
				writeErrorResponse(w, errorJSONUnmarshal)
				return
			}
		}
		results := reflected.Call(callArgs)
		var returnValue interface{}
		var returnError interface{}
		for i, result := range results {
			if result.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				returnError = result.Interface()
				results = append(results[:i], results[i+1:]...)
				break
			}
		}
		if returnError != nil {
			writeErrorResponse(w, returnError.(error))
			return
		}
		if len(results) != 1 {
			writeErrorResponse(w, errors.New("call must be a function with one return value"))
			return
		}
		if !results[0].IsZero() {
			returnValue = results[0].Interface()
		}
		if returnValue == nil {
			returnValue = Response{
				Code: 200,
				Data: []byte("success"),
			}
		}
		returnData, _ := jsonEncodeWithoutEscapeHTML(returnValue)
		responseData := Response{
			Code: 200,
			Data: returnData,
		}
		w.Header().Set("Content-Type", "application/json")
		writeData, _ := jsonEncodeWithoutEscapeHTML(responseData)
		_, _ = w.Write(writeData)
	}
}

func jsonEncodeWithoutEscapeHTML(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeErrorResponse(w http.ResponseWriter, err error) {
	var code int
	switch err.(type) {
	case ErrorParams:
		code = http.StatusBadRequest
	default:
		code = http.StatusInternalServerError
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_b, _ := jsonEncodeWithoutEscapeHTML(Response{
		Code: code,
		Msg:  err.Error(),
	})
	_, _ = w.Write(_b)
}
