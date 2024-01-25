package st

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/Fantom-foundation/Tosca/go/ct/common"
)

func TestCallContext_NewCallContext(t *testing.T) {
	tests := map[string]struct {
		equal func(*CallContext) bool
	}{
		"AccountAddress": {func(c *CallContext) bool { want, got := (Address{}), c.AccountAddress; return want == got }},
		"OriginAddress":  {func(c *CallContext) bool { want, got := (Address{}), c.OriginAddress; return want == got }},
		"CallerAddress":  {func(c *CallContext) bool { want, got := (Address{}), c.CallerAddress; return want == got }},
		"Value":          {func(c *CallContext) bool { want, got := (NewU256()), c.Value; return want.Eq(got) }},
	}

	callContext := NewCallContext()
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if !test.equal(&callContext) {
				t.Error("Unexpected value in new context")
			}
		})
	}

}

func TestCallContext_Diff(t *testing.T) {
	tests := map[string]struct {
		change func(*CallContext)
	}{
		"Account Address": {func(c *CallContext) { c.AccountAddress[0]++ }},
		"Origin Address":  {func(c *CallContext) { c.OriginAddress[0]++ }},
		"Caller Address":  {func(c *CallContext) { c.CallerAddress[0]++ }},
		"Value":           {func(c *CallContext) { c.Value = NewU256(1) }},
	}

	callContext := NewCallContext()
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			c2 := NewCallContext()
			test.change(&c2)
			if diffs := callContext.Diff(&c2); len(diffs) == 0 {
				t.Errorf("No difference found in modified %v", name)
			}
		})
	}
}

func TestCallContext_String(t *testing.T) {
	tests := map[string]struct {
		change func(*CallContext) any
	}{
		"Account Address": {func(c *CallContext) any { c.AccountAddress[19] = 0xff; return c.AccountAddress }},
		"Origin Address":  {func(c *CallContext) any { c.OriginAddress[19] = 0xfe; return c.OriginAddress }},
		"Caller Address":  {func(c *CallContext) any { c.CallerAddress[19] = 0xfd; return c.CallerAddress }},
		"Value":           {func(c *CallContext) any { c.Value = NewU256(1); return c.Value }},
	}

	c := NewCallContext()
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := test.change(&c)
			str := c.String()
			if !strings.Contains(str, fmt.Sprintf("%v: %v", name, v)) {
				t.Errorf("Did not find %v string", name)
			}
		})
	}
}