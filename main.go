//go:build wasm

// JS Bindings for D2
// Based on https://github.com/alixander/d2/tree/wasm-build/d2js

package main

import (
	"context"
	"encoding/json"
	"syscall/js"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2target"
	"oss.terrastruct.com/d2/d2themes"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"oss.terrastruct.com/util-go/go2"
)

func main() {
	js.Global().Set("d2RenderSVG", js.FuncOf(jsRenderSVG))

	initCallback := js.Global().Get("onWasmInitialized")
	if !initCallback.IsUndefined() {
		initCallback.Invoke()
	}
	select {}
}

type jsObjRenderSVG struct {
	SVG       string `json:"svg"`
	Error     string `json:"error"`
	UserError string `json:"userError"`
	D2Error   string `json:"d2Error"`
}

func jsRenderSVG(this js.Value, args []js.Value) interface{} {
	obj := args[0]
	if obj.Type() != js.TypeObject {
		ret := jsObjRenderSVG{UserError: "first argument must be an object"}
		str, _ := json.Marshal(ret)
		return string(str)
	}

	dsl := obj.Get("dsl").String()
	dark := obj.Get("dark").Bool()
	sketch := obj.Get("sketch").Bool()
	var center *bool = nil
	if !obj.Get("center").IsUndefined() {
		center = go2.Pointer(obj.Get("center").Bool())
	} else {
		center = go2.Pointer(true)
	}
	var padding *int64 = nil
	if !obj.Get("padding").IsUndefined() {
		padding = go2.Pointer(int64(obj.Get("padding").Int()))
	} else {
		padding = go2.Pointer(int64(5))
	}
	var scale *float64 = nil
	if !obj.Get("scale").IsUndefined() {
		scale = go2.Pointer(obj.Get("scale").Float())
	} else {
		scale = go2.Pointer(1.0)
	}

	ruler, _ := textmeasure.NewRuler()
	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return d2dagrelayout.DefaultLayout, nil
	}

	var theme *d2themes.Theme
	var overrides *d2target.ThemeOverrides
	if dark {
		theme = &d2themescatalog.DarkFlagshipTerrastruct
		overrides = &d2target.ThemeOverrides{
			B1:  go2.Pointer("#F0F0F0"),
			B2:  go2.Pointer("#989898"),
			B3:  go2.Pointer("#707070"),
			B4:  go2.Pointer("#303030"),
			B5:  go2.Pointer("#212121"),
			B6:  go2.Pointer("#111111"),
			AA2: go2.Pointer("#989898"),
			AA4: go2.Pointer("#303030"),
			AA5: go2.Pointer("#212121"),
			AB4: go2.Pointer("#303030"),
			AB5: go2.Pointer("#212121"),
		}
	} else {
		theme = &d2themescatalog.NeutralGrey
	}

	renderOpts := &d2svg.RenderOpts{
		Pad:            padding,
		ThemeID:        &theme.ID,
		ThemeOverrides: overrides,
		Sketch:         &sketch,
		Center:         center,
		Scale:          scale,
	}

	compileOpts := &d2lib.CompileOptions{
		LayoutResolver: layoutResolver,
		Ruler:          ruler,
	}

	diagram, _, err := d2lib.Compile(context.Background(), dsl, compileOpts, renderOpts)

	if err != nil {
		ret := jsObjRenderSVG{D2Error: err.Error()}
		str, _ := json.Marshal(ret)
		return string(str)
	}

	out, err := d2svg.Render(diagram, renderOpts)

	if err != nil {
		ret := jsObjRenderSVG{D2Error: err.Error()}
		str, _ := json.Marshal(ret)
		return string(str)
	}

	ret := jsObjRenderSVG{
		SVG: string(out),
	}
	str, _ := json.Marshal(ret)
	return string(str)
}
