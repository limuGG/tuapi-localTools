package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
)

var IncSigner = new(Signer)

type Map map[string]interface{}

type Signer struct{}

func (*Signer) GenerateKey() (val *SignerKey, err error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return
	}
	val = &SignerKey{
		PrivateKey: hex.EncodeToString(crypto.FromECDSA(key)),
		PublicKey:  hex.EncodeToString(crypto.FromECDSAPub(&key.PublicKey)),
	}
	tronAddress := address.PubkeyToAddress(key.PublicKey)
	val.Hex = strings.TrimPrefix(tronAddress.Hex(), "0x")
	val.Base58 = tronAddress.String()
	val.ToUpper()
	return
}

func (*Signer) Sign(prv, input string) (signature string, err error) {
	key, err := crypto.HexToECDSA(prv)
	if err != nil {
		return
	}
	data, err := hex.DecodeString(input)
	if err != nil {
		return
	}
	hash := sha256.Sum256(data)
	signatureBytes, err := crypto.Sign(hash[:], key)
	if err != nil {
		return
	}
	signature = hex.EncodeToString(signatureBytes)
	return
}

type SignerKey struct {
	PrivateKey string `json:"private_key,omitempty"`
	PublicKey  string `json:"public_key,omitempty"`
	Hex        string `json:"hex,omitempty"`
	Base58     string `json:"base58,omitempty"`
}

func (s *SignerKey) ToUpper() {
	s.PrivateKey = strings.ToUpper(s.PrivateKey)
	s.PublicKey = strings.ToUpper(s.PublicKey)
	s.Hex = strings.ToUpper(s.Hex)
}

func NewSignerKeyFromPrivateKey(privateKey string) (rst *SignerKey, err error) {
	rst = new(SignerKey)
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return
	}
	rst.PrivateKey = hex.EncodeToString(crypto.FromECDSA(pk))
	rst.PublicKey = hex.EncodeToString(crypto.FromECDSAPub(&pk.PublicKey))
	tronAddress := address.PubkeyToAddress(pk.PublicKey)
	rst.Hex = strings.TrimPrefix(tronAddress.Hex(), "0x")
	rst.Base58 = tronAddress.String()
	rst.ToUpper()
	return
}

func SignTransactionTron(txPtr *string, privateKey string) (err error) {
	tx := *txPtr
	var m Map
	if err = json.Unmarshal([]byte(tx), &m); err != nil {
		return
	}
	rawHex, has := m["raw_data_hex"]
	if !has {
		return errors.New("数据错误")
	}
	signature, err := IncSigner.Sign(privateKey, AnyToString(rawHex))
	if err != nil {
		return
	}
	m["signature"] = []string{signature}
	v1, err := json.Marshal(m)
	if err != nil {
		return
	}
	*txPtr = string(v1)
	return
}
