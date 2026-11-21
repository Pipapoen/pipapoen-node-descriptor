// Generate public examples using a synthetic, public all-zero demonstration key.
package main

import (
	"encoding/json"
	"os"
	"github.com/Pipapoen/pipapoen-node-descriptor/internal/catalog"
	"github.com/Pipapoen/pipapoen-node-descriptor/internal/codec"
)

func main() {
	engine, err := codec.New(make([]byte, 32), catalog.Dictionary())
	if err != nil { panic(err) }
	inputs := []string{
		`{"nodeId":"0123456789abcdef0123"}`,
		`{"nodeId":"0123456789abcdef0123","geo":5128581,"provider":1,"cpu":1,"ram":4,"storage":2}`,
		`{"nodeId":"0123456789abcdef0123","geo":5128581,"provider":0,"cpu":3,"ram":8,"storage":5,"ipv4":"192.0.2.10"}`,
	}
	rows := []map[string]any{}
	for _, input := range inputs {
		var d codec.Descriptor
		if err := json.Unmarshal([]byte(input), &d); err != nil { panic(err) }
		compact, err := engine.Encode(d,"x","example.org",false)
		if err != nil { panic(err) }
		padded, err := engine.Encode(d,"x","example.org",true)
		if err != nil { panic(err) }
		rows=append(rows,map[string]any{"descriptor":d,"compact":compact,"padded":padded})
	}
	b, err := json.MarshalIndent(rows,"","  ")
	if err != nil { panic(err) }
	if err := os.WriteFile("demo/src/examples.json",append(b,'\n'),0644); err != nil { panic(err) }
}
