package main

import (
	"log"
	"net/http"

	"localtools/core"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/encrypt", core.EncryptHandler())
	mux.HandleFunc("/decrypt", core.DecryptHandler())
	mux.HandleFunc("/generateTronAddress", core.GenerateTronAddressHandler())
	mux.HandleFunc("/transferTRX", core.TransferTRXHandler())
	mux.HandleFunc("/transferUSDT", core.TransferUSDTHandler())
	log.Fatalln(http.ListenAndServe(":8080", mux))
}
