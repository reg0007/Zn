package debug

import (
	"reflect"
	"time"
)

// Probe - debug probe
// log variable into when （__probe：xx，xx） is called
type Probe struct {
	info map[string][]ProbeLog
}

// ProbeLog -
type ProbeLog struct {
	ProbeTime time.Time
	// original value - DON'T USE ZnValue here to avoid circular dependency!
	Value    interface{}
	ValueStr string
	// valueType - get actual  valueType (*exec.ZnXXXX) as string
	ValueType string
}

// NewProbe -
func NewProbe() *Probe {
	return &Probe{
		info: map[string][]ProbeLog{},
	}
}

// AddLog - add probe data to log
func (pb *Probe) AddLog(tag string, value interface{}) {
	var valStr, valType string
	now := time.Now()
	// init probeInfo
	if _, ok := pb.info[tag]; !ok {
		pb.info[tag] = []ProbeLog{}
	}
	// TODO: add deepcopy
	rv := reflect.ValueOf(value)

	vstrMethod := rv.MethodByName("String")
	if vstrMethod.IsValid() {
		vResults := vstrMethod.Call([]reflect.Value{})
		valStr = vResults[0].String()
	}

	// add valType
	valType = rv.Type().String()

	probeLog := ProbeLog{
		ProbeTime: now,
		Value:     value,
		ValueStr:  valStr,
		ValueType: valType,
	}
	pb.info[tag] = append(pb.info[tag], probeLog)
}

// GetProbeLog -
func (pb *Probe) GetProbeLog(tag string) []ProbeLog {
	probeLog, ok := pb.info[tag]
	if ok {
		return probeLog
	}
	// return empty ProbeLog array when tag not found
	return []ProbeLog{}
}
