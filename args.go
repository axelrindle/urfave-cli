package cli

import (
	"fmt"
	"time"
)

type Args interface {
	// Get returns the nth argument, or else a blank string
	Get(n int) string
	// First returns the first argument, or else a blank string
	First() string
	// Tail returns the rest of the arguments (not the first one)
	// or else an empty string slice
	Tail() []string
	// Len returns the length of the wrapped slice
	Len() int
	// Present checks if there are any arguments present
	Present() bool
	// Slice returns a copy of the internal slice
	Slice() []string
}

type stringSliceArgs struct {
	v []string
}

func (a *stringSliceArgs) Get(n int) string {
	if len(a.v) > n {
		return a.v[n]
	}
	return ""
}

func (a *stringSliceArgs) First() string {
	return a.Get(0)
}

func (a *stringSliceArgs) Tail() []string {
	if a.Len() >= 2 {
		tail := a.v[1:]
		ret := make([]string, len(tail))
		copy(ret, tail)
		return ret
	}

	return []string{}
}

func (a *stringSliceArgs) Len() int {
	return len(a.v)
}

func (a *stringSliceArgs) Present() bool {
	return a.Len() != 0
}

func (a *stringSliceArgs) Slice() []string {
	ret := make([]string, len(a.v))
	copy(ret, a.v)
	return ret
}

type Argument interface {
	Parse([]string) ([]string, error)
	Usage() string
}

type ArgumentBase[T any, C any, VC ValueCreator[T, C]] struct {
	Name        string // the name of this argument
	Value       T      // the default value of this argument
	Destination *T     // the destination point for this argument
	Values      *[]T   // all the values of this argument, only if multiple are supported
	UsageText   string // the usage text to show
	Min         int    // the min num of occurrences of this argument
	Max         int    // the max num of occurrences of this argument, set to -1 for unlimited
	Config      C      // config for this argument similar to Flag Config
}

func (a *ArgumentBase[T, C, VC]) Usage() string {
	if a.UsageText != "" {
		return a.UsageText
	}

	usageFormat := ""
	if a.Min == 0 {
		if a.Max == 1 {
			usageFormat = "[%[1]s]"
		} else {
			usageFormat = "[%[1]s ...]"
		}
	} else {
		usageFormat = "%[1]s [%[1]s ...]"
	}
	return fmt.Sprintf(usageFormat, a.Name)
}

func (a *ArgumentBase[T, C, VC]) Parse(s []string) ([]string, error) {
	tracef("calling arg%[1] parse with args %[2]", &a.Name, s)
	if a.Max == 0 {
		fmt.Printf("WARNING args %s has max 0, not parsing argument", a.Name)
		return s, nil
	}
	if a.Max != -1 && a.Min > a.Max {
		fmt.Printf("WARNING args %s has min[%d] > max[%d], not parsing argument", a.Name, a.Min, a.Max)
		return s, nil
	}

	count := 0
	var vc VC
	var t T
	value := vc.Create(a.Value, &t, a.Config)
	values := []T{}

	for _, arg := range s {
		if err := value.Set(arg); err != nil {
			return s, err
		}
		values = append(values, value.Get().(T))
		count++
		if count >= a.Max {
			break
		}
	}
	if count < a.Min {
		return s, fmt.Errorf("sufficient count of arg %s not provided, given %d expected %d", a.Name, count, a.Min)
	}

	if a.Values == nil {
		a.Values = &values
	} else {
		*a.Values = values
	}

	if a.Max == 1 && a.Destination != nil {
		*a.Destination = values[0]
	}
	return s[count:], nil
}

type FloatArg = ArgumentBase[float64, NoConfig, floatValue]
type IntArg = ArgumentBase[int64, IntegerConfig, intValue]
type StringArg = ArgumentBase[string, StringConfig, stringValue]
type StringMapArg = ArgumentBase[map[string]string, StringConfig, StringMap]
type TimestampArg = ArgumentBase[time.Time, TimestampConfig, timestampValue]
type UintArg = ArgumentBase[uint64, IntegerConfig, uintValue]
