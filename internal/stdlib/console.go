package stdlib

import (
	"encoding/json"
	"fmt"
	"strings"
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
	prefix := group_prefix()
	fmt.Printf(prefix+"%s: %d\n", label, count)
}

func (c *Console) CountReset(args ...any) {
	label := "default"
	if len(args) >= 1 {
		label = args[0].(string)
	}
	counters[label] = 0
	prefix := group_prefix()
	fmt.Printf(prefix+"%s: %d\n", label, 0)
}

func (c *Console) Debug(msg string, args ...any) {
	prefix := group_prefix()
	if len(args) > 0 && strings.Contains(msg, "%") {
		fmt.Printf(prefix+msg+"\n", args...)
	} else {
		allArgs := make([]any, 0, len(args)+1)
		allArgs = append(allArgs, prefix+msg)
		allArgs = append(allArgs, args...)
		fmt.Println(allArgs...)
	}
}

type DirOptions struct {
	Colors     *string `json:"colors,omitempty"`
	Depth      *int    `json:"depth,omitempty"`
	ShowHidden *bool   `json:"showHidden,omitempty"`
}

func (c *Console) Dir(obj any, options *DirOptions) {
	prefix := group_prefix()
	if options == nil {
		fmt.Printf(prefix+"%v\n", obj)
	} else {
		c.Dirxml(obj)
	}
}

func (c *Console) Dirxml(obj any) {
	data, e := json.MarshalIndent(obj, "", " ")
	if e != nil {
		prefix := group_prefix()
		fmt.Printf(prefix+"Error serializing object: %v\n", e)
		return
	}
	prefix := group_prefix()
	fmt.Print(prefix + string(data) + "\n")
}

func (c *Console) Error(args ...any) {
	prefix := group_prefix()
	if len(args) > 0 {
		if msg, ok := args[0].(string); ok {
			args = args[1:]
			if len(args) > 0 && strings.Contains(msg, "%") {
				fmt.Printf(prefix+msg+"\n", args...)
			} else {
				allArgs := make([]any, 0, len(args)+1)
				allArgs = append(allArgs, prefix+msg)
				allArgs = append(allArgs, args...)
				fmt.Println(allArgs...)
			}
		} else {
			allArgs := make([]any, 0, len(args)+1)
			allArgs = append(allArgs, prefix)
			allArgs = append(allArgs, args...)
			fmt.Println(allArgs...)
		}
	}
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
	prefix := group_prefix()
	if len(a) > 0 && strings.Contains(message, "%") {
		format := prefix + message + "\n"
		fmt.Printf(format, a...)
	} else {
		args := make([]any, 0, len(a)+1)
		args = append(args, prefix+message)
		args = append(args, a...)
		fmt.Println(args...)
	}
}
