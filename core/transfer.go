package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const gridURL = "https://api.trongrid.io"

var rateLimitForGrid = rate.NewLimiter(rate.Every(time.Millisecond*300), 2)

type HTTPResponse struct {
	http.Response
	BodyBytes []byte
}

type TransferParams struct {
	To         string  `json:"to"`
	Amount     float64 `json:"amount"`
	PrivateKey string  `json:"private_key"`
}

func TransferTRXHandler() http.HandlerFunc {
	return JSONHandlerWrap(func(p *TransferParams) (interface{}, error) {
		return transferTRX(p.To, p.Amount, p.PrivateKey)
	})
}

func TransferUSDTHandler() http.HandlerFunc {
	return JSONHandlerWrap(func(p *TransferParams) (interface{}, error) {
		return transferUSDT(p.To, p.Amount, p.PrivateKey)
	})
}

func request(url string, method string, body interface{}) (rp *HTTPResponse, err error) {
	_ = rateLimitForGrid.Wait(context.Background())

	var reader io.Reader
	if body != nil {
		var b []byte
		switch _body := body.(type) {
		case string:
			b = []byte(_body)
		case []byte:
			b = _body
		default:
			b, err = json.Marshal(body)
			if err != nil {
				return
			}
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return
	}
	if reader != nil {
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
	}
	rsp, err := http.DefaultClient.Do(req)
	defer func() {
		if rsp != nil && rsp.Body != nil {
			_ = rsp.Body.Close()
		}
	}()
	if err != nil {
		return
	}
	rp = &HTTPResponse{Response: *rsp}
	rp.BodyBytes, err = io.ReadAll(rsp.Body)
	return
}

func signAndBroadcast(transactionStr, privateKey string) (ref Map, err error) {
	if err = SignTransactionTron(&transactionStr, privateKey); err != nil {
		err = fmt.Errorf("签名失败:%v", err)
		return
	}
	url := gridURL + "/wallet/broadcasttransaction"
	rp, err := request(url, http.MethodPost, transactionStr)
	if err != nil {
		return
	}
	if err = json.Unmarshal(rp.BodyBytes, &ref); err != nil {
		return
	}
	if _rs := ref["result"]; _rs != true {
		msg := AnyToString(ref["message"])
		if strings.Contains(msg, "Account resource insufficient error") {
			err = ErrorParams("广播交易失败:账户资源不足,请先转入TRX")
			return
		}
		err = fmt.Errorf("广播交易失败: %s", msg)
		return
	}
	return
}

func transferTRX(to string, amount float64, privateKey string) (v interface{}, err error) {
	sk, err := NewSignerKeyFromPrivateKey(privateKey)
	if err != nil {
		return
	}
	from := sk.Base58

	if from == "" || to == "" || amount <= 0 || privateKey == "" {
		err = ErrorParams("参数错误")
		return
	}
	amountInt := int64(math.Floor(amount * math.Pow10(6)))
	if amountInt <= 0 {
		err = ErrorParams("金额错误")
		return
	}
	url := gridURL + "/wallet/createtransaction"
	body := Map{
		"owner_address": from,
		"to_address":    to,
		"amount":        amountInt,
		"visible":       true,
	}
	rp, err := request(url, http.MethodPost, body)
	if err != nil {
		return
	}
	ref := Map{}
	if err = json.Unmarshal(rp.BodyBytes, &ref); err != nil {
		return
	}
	if e, has := ref["Error"]; has {
		errorMsg := AnyToString(e)
		if strings.Contains(errorMsg, "no OwnerAccount") {
			err = ErrorParams("owner账户不存在")
			return
		}
		if strings.Contains(errorMsg, "balance is not sufficient") {
			err = ErrorParams("余额不足")
			return
		}
		err = errors.New(AnyToString(e))
		return
	}
	_, has := ref["raw_data_hex"]
	if !has {
		err = ErrorParams("创建交易失败,raw_data_hex不存在")
		return
	}
	return signAndBroadcast(string(rp.BodyBytes), privateKey)
}

func transferUSDT(to string, amount float64, privateKey string) (v interface{}, err error) {
	sk, err := NewSignerKeyFromPrivateKey(privateKey)
	if err != nil {
		return
	}
	from := sk.Base58

	if from == "" || to == "" || amount <= 0 || privateKey == "" {
		err = ErrorParams("参数错误")
		return
	}
	amountInt := int64(math.Floor(amount * math.Pow10(6)))
	if amountInt <= 0 {
		err = ErrorParams("金额错误")
		return
	}
	url := gridURL + "/wallet/triggersmartcontract"
	body := Map{
		"owner_address":     from,
		"contract_address":  "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		"function_selector": "transfer(address,uint256)",
		"fee_limit":         100000000,
		"visible":           true,
	}
	parameter, err := Trc20ParameterEncodeTransfer(TransferParameter{
		ToAddress: to,
		Amount:    amountInt,
	})
	if err != nil {
		return
	}
	body["parameter"] = parameter
	rp, err := request(url, http.MethodPost, body)
	if err != nil {
		return
	}
	ref := Map{}
	if err = json.Unmarshal(rp.BodyBytes, &ref); err != nil {
		return
	}
	transaction, has := ref["transaction"]
	if !has {
		err = ErrorParams("创建交易失败")
		return
	}
	return signAndBroadcast(AnyToString(transaction), privateKey)
}
