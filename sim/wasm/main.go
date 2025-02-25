// +build wasm

package main

import (
	"log"
	"syscall/js"

	"github.com/wowsims/tbc/sim"
	"github.com/wowsims/tbc/sim/core"
	proto "github.com/wowsims/tbc/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

func init() {
	sim.RegisterAll()
}

func main() {
	c := make(chan struct{}, 0)

	js.Global().Set("computeStats", js.FuncOf(computeStats))
	js.Global().Set("gearList", js.FuncOf(gearList))
	js.Global().Set("raidSim", js.FuncOf(raidSim))
	js.Global().Set("raidSimAsync", js.FuncOf(raidSimAsync))
	js.Global().Set("statWeights", js.FuncOf(statWeights))
	js.Global().Set("statWeightsAsync", js.FuncOf(statWeightsAsync))
	js.Global().Call("wasmready")
	<-c
}

func computeStats(this js.Value, args []js.Value) interface{} {
	csr := &proto.ComputeStatsRequest{}
	if err := googleProto.Unmarshal(getArgsBinary(args[0]), csr); err != nil {
		log.Printf("Failed to parse request: %s", err)
		return nil
	}
	result := core.ComputeStats(csr)

	outbytes, err := googleProto.Marshal(result)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal result: %s", err.Error())
		return nil
	}

	outArray := js.Global().Get("Uint8Array").New(len(outbytes))
	js.CopyBytesToJS(outArray, outbytes)

	return outArray
}

func gearList(this js.Value, args []js.Value) interface{} {
	glr := &proto.GearListRequest{}
	if err := googleProto.Unmarshal(getArgsBinary(args[0]), glr); err != nil {
		log.Printf("Failed to parse request: %s", err)
		return nil
	}
	result := core.GetGearList(glr)

	outbytes, err := googleProto.Marshal(result)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal result: %s", err.Error())
		return nil
	}

	outArray := js.Global().Get("Uint8Array").New(len(outbytes))
	js.CopyBytesToJS(outArray, outbytes)

	return outArray
}

func raidSim(this js.Value, args []js.Value) interface{} {
	rsr := &proto.RaidSimRequest{}
	if err := googleProto.Unmarshal(getArgsBinary(args[0]), rsr); err != nil {
		log.Printf("Failed to parse request: %s", err)
		return nil
	}
	result := core.RunRaidSim(rsr)

	outbytes, err := googleProto.Marshal(result)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal result: %s", err.Error())
		return nil
	}

	outArray := js.Global().Get("Uint8Array").New(len(outbytes))
	js.CopyBytesToJS(outArray, outbytes)

	return outArray
}

func raidSimAsync(this js.Value, args []js.Value) interface{} {
	rsr := &proto.RaidSimRequest{}
	if err := googleProto.Unmarshal(getArgsBinary(args[0]), rsr); err != nil {
		log.Printf("Failed to parse request: %s", err)
		return nil
	}
	reporter := make(chan *proto.ProgressMetrics, 100)
	core.RunRaidSimAsync(rsr, reporter)

	result := processAsyncProgress(args[1], reporter)
	close(reporter)
	return result
}

func statWeights(this js.Value, args []js.Value) interface{} {
	swr := &proto.StatWeightsRequest{}
	if err := googleProto.Unmarshal(getArgsBinary(args[0]), swr); err != nil {
		log.Printf("Failed to parse request: %s", err)
		return nil
	}
	result := core.StatWeights(swr)

	outbytes, err := googleProto.Marshal(result)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal result: %s", err.Error())
		return nil
	}

	outArray := js.Global().Get("Uint8Array").New(len(outbytes))
	js.CopyBytesToJS(outArray, outbytes)

	return outArray
}

func statWeightsAsync(this js.Value, args []js.Value) interface{} {
	rsr := &proto.StatWeightsRequest{}
	if err := googleProto.Unmarshal(getArgsBinary(args[0]), rsr); err != nil {
		log.Printf("Failed to parse request: %s", err)
		return nil
	}
	reporter := make(chan *proto.ProgressMetrics, 100)
	core.StatWeightsAsync(rsr, reporter)

	result := processAsyncProgress(args[1], reporter)
	close(reporter)
	return result
}

// Assumes args[0] is a Uint8Array
func getArgsBinary(value js.Value) []byte {
	data := make([]byte, value.Get("length").Int())
	js.CopyBytesToGo(data, value)
	return data
}

func processAsyncProgress(progFunc js.Value, reporter chan *proto.ProgressMetrics) js.Value {
reader:
	for {
		// TODO: cleanup so we dont collect these
		select {
		case progMetric, ok := <-reporter:
			if !ok {
				break reader
			}
			outbytes, err := googleProto.Marshal(progMetric)
			if err != nil {
				log.Printf("[ERROR] Failed to marshal result: %s", err.Error())
				return js.Undefined()
			}
			outArray := js.Global().Get("Uint8Array").New(len(outbytes))
			js.CopyBytesToJS(outArray, outbytes)
			progFunc.Invoke(outArray)

			if progMetric.FinalWeightResult != nil || progMetric.FinalRaidResult != nil {
				return outArray
			}
		}
	}

	return js.Undefined()
}
