package core

import (
	"net/http"
)

func GenerateTronAddress() (r interface{}, err error) { return IncSigner.GenerateKey() }

func GenerateTronAddressHandler() http.HandlerFunc { return JSONHandlerWrap(GenerateTronAddress) }
