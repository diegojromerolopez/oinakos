package game

import (
	"fmt"
	"math"
	"math/rand"
	"gopkg.in/yaml.v3"
)

type IntInterval struct { Min, Max int; Mean, SD float64 `yaml:"mean,omitempty"`; Mode string `yaml:"mode,omitempty"` }
func (i *IntInterval) UnmarshalYAML(v *yaml.Node) error {
	var val int; if err := v.Decode(&val); err == nil { i.Min, i.Max = val, val; return nil }
	var list []int; if err := v.Decode(&list); err == nil && len(list) == 2 { i.Min, i.Max = list[0], list[1]; return nil }
	type mStruct struct {
		Min  int     `yaml:"min"`
		Max  int     `yaml:"max"`
		Mean float64 `yaml:"mean"`
		SD   float64 `yaml:"sd"`
		Mode string  `yaml:"mode"`
	}
	var m mStruct
	if err := v.Decode(&m); err == nil {
		i.Min, i.Max, i.Mode, i.Mean, i.SD = m.Min, m.Max, m.Mode, m.Mean, m.SD
		return nil
	}
	return fmt.Errorf("invalid int interval")
}
func (i IntInterval) Roll() int {
	if i.Min >= i.Max && i.Mode != "normal" { return i.Min }
	if i.Mode == "normal" {
		mean, sd := i.Mean, i.SD; if mean == 0 && sd == 0 && i.Min != i.Max { mean, sd = float64(i.Min+i.Max)/2.0, float64(i.Max-i.Min)/6.0 }
		res := rand.NormFloat64()*sd + mean
		if i.Max > i.Min { if res < float64(i.Min) { res = float64(i.Min) }; if res > float64(i.Max) { res = float64(i.Max) } }
		return int(math.Round(res))
	}
	return i.Min + rand.Intn(i.Max-i.Min+1)
}
func (i IntInterval) String() string { if i.Min == i.Max { return fmt.Sprintf("%d", i.Min) }; return fmt.Sprintf("%d-%d", i.Min, i.Max) }
func (i IntInterval) IsZero() bool { return i.Min == 0 && i.Max == 0 && i.Mode == "" }

type FloatInterval struct { Min, Max, Mean, SD float64; Mode string `yaml:"mode,omitempty"` }
func (i *FloatInterval) UnmarshalYAML(v *yaml.Node) error {
	var val float64; if err := v.Decode(&val); err == nil { i.Min, i.Max = val, val; return nil }
	var list []float64; if err := v.Decode(&list); err == nil && len(list) == 2 { i.Min, i.Max = list[0], list[1]; return nil }
	type mStruct struct {
		Min  float64 `yaml:"min"`
		Max  float64 `yaml:"max"`
		Mean float64 `yaml:"mean"`
		SD   float64 `yaml:"sd"`
		Mode string  `yaml:"mode"`
	}
	var m mStruct
	if err := v.Decode(&m); err == nil {
		i.Min, i.Max, i.Mode, i.Mean, i.SD = m.Min, m.Max, m.Mode, m.Mean, m.SD
		return nil
	}
	return fmt.Errorf("invalid float interval")
}
func (i FloatInterval) String() string { if i.Min == i.Max { if i.Mode == "normal" && i.Mean != 0 { return fmt.Sprintf("~%.2f", i.Mean) }; return fmt.Sprintf("%.2f", i.Min) }; return fmt.Sprintf("%.2f-%.2f", i.Min, i.Max) }
func (i FloatInterval) Roll() float64 {
	if i.Min >= i.Max && i.Mode != "normal" { return i.Min }
	if i.Mode == "normal" {
		mean, sd := i.Mean, i.SD; if mean == 0 && sd == 0 { mean, sd = (i.Min+i.Max)/2.0, (i.Max-i.Min)/6.0 }
		res := rand.NormFloat64()*sd + mean
		if i.Max > i.Min { if res < i.Min { res = i.Min }; if res > i.Max { res = i.Max } }
		return res
	}
	return i.Min + rand.Float64()*(i.Max-i.Min)
}
func (i FloatInterval) IsZero() bool { return i.Min == 0 && i.Max == 0 }
