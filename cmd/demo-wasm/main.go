//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/Pipapoen/pipapoen-node-descriptor/internal/democlient"
)

func main() {
	client, err := democlient.New()
	if err != nil {
		panic(err)
	}
	request := js.FuncOf(func(_ js.Value, args []js.Value) any {
		result := client.Request(args[0].String(), args[1].String())
		b, err := json.Marshal(result)
		if err != nil {
			return `{"ok":false,"result":{"error":"error"}}`
		}
		return string(b)
	})
	js.Global().Set("pndDemoRequest", request)
	select {}
}
