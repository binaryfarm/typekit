package stdlib

import (
	"encoding/json"
	"fmt"
)

var counters map[string]uint64 = make(map[string]uint64)
var group_i uint

type Console struct {
}

func (c *Console) Assert(args ...any) {
	// TODO revisit and map args to expr
}

func (c *Console) Clear() {
	fmt.Printf("\033[2J\033[H")
}

func (c *Console) Count(args ...any) {
	label := "default"
	if len(args) >= 1 {
		label = args[0].(string)
	}
	var count uint64
	if v, ok := counters[label]; !ok {
		counters[label] = 1
		count = 1
	} else {
		count = v + 1
		counters[label] = count

	}
	fmt.Printf(fmt.Sprintf("%s%s", group_prefix(), "%s: %d\n"), label, count)
}

func (c *Console) CountReset(args ...any) {
	label := "default"
	if len(args) >= 1 {
		label = args[0].(string)
	}
	counters[label] = 0
	fmt.Printf(fmt.Sprintf("%s%s\n", group_prefix(), "%s: %d"), label, 0)
}

func (c *Console) Debug(msg string, args ...any) {
	fmt.Printf(fmt.Sprintf("%s%s\n", group_prefix(), msg), args...)
}

type DirOptions struct {
	Colors     *string `json:"colors,omitempty"`
	Depth      *int    `json:"depth,omitempty"`
	ShowHidden *bool   `json:"showHidden,omitempty"`
}

func (c *Console) Dir(obj any, options *DirOptions) {
	if options == nil {
		fmt.Printf(fmt.Sprintf("%s%s\n", group_prefix(), "%o"), obj)
	} else {
		c.Dirxml(obj)
	}
}
func (c *Console) Dirxml(obj any) {
	data, e := json.MarshalIndent(obj, "", " ")
	if e != nil {
		panic(e)
	}
	fmt.Printf(fmt.Sprintf("%s%s\n", group_prefix(), "%s"), string(data))
}

func (c *Console) Error(args ...any) {
	if msg, ok := args[0].(string); ok {
		args = args[1:]
		fmt.Printf(fmt.Sprintf("%s%s\n", group_prefix(), msg), args...)
	} else {

	}
	fmt.Printf("%v\n", args...)
}
func group_prefix() string {
	ret := ""
	for range group_i {
		ret = fmt.Sprintf("\t%s", ret)
	}
	return ret
}
func (c *Console) Group() {
	group_i = group_i + 1
}

func (c *Console) GroupEnd() {
	if group_i > 0 {
		group_i = group_i - 1
	}
}

func (c *Console) Log(message string, a ...any) {
	if len(a) > 0 {
		fmt.Printf(fmt.Sprintf("%s%s\n", group_prefix(), message), a...)
	} else {
		fmt.Printf(fmt.Sprintf("%s%s\n", group_prefix(), message), "")
	}
}
