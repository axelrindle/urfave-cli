package cli

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestArgumentsRootCommand(t *testing.T) {

	cmd := buildMinimalTestCommand()
	var ival int64
	var fval float64
	var fvals []float64
	cmd.Arguments = []Argument{
		&IntArg{
			Name:        "ia",
			Min:         1,
			Max:         1,
			Destination: &ival,
		},
		&FloatArg{
			Name:        "fa",
			Min:         0,
			Max:         2,
			Destination: &fval,
			Values:      &fvals,
		},
	}

	require.NoError(t, cmd.Run(context.Background(), []string{"foo", "10"}))
	require.Equal(t, int64(10), ival)

	require.NoError(t, cmd.Run(context.Background(), []string{"foo", "12", "10.1"}))
	require.Equal(t, int64(12), ival)
	require.Equal(t, []float64{10.1}, fvals)

	require.NoError(t, cmd.Run(context.Background(), []string{"foo", "13", "10.1", "11.09"}))
	require.Equal(t, int64(13), ival)
	require.Equal(t, []float64{10.1, 11.09}, fvals)

	require.Error(t, errors.New("No help topic for '12.1"), cmd.Run(context.Background(), []string{"foo", "13", "10.1", "11.09", "12.1"}))
	require.Equal(t, int64(13), ival)
	require.Equal(t, []float64{10.1, 11.09}, fvals)
}

func TestArgumentsSubcommand(t *testing.T) {

	cmd := buildMinimalTestCommand()
	var ifval int64
	var svals []string
	var tval time.Time
	cmd.Commands = []*Command{
		{
			Name: "subcmd",
			Flags: []Flag{
				&IntFlag{
					Name:        "foo",
					Value:       10,
					Destination: &ifval,
				},
			},
			Arguments: []Argument{
				&TimestampArg{
					Name:        "ta",
					Min:         1,
					Max:         1,
					Destination: &tval,
					Config: TimestampConfig{
						Layout: time.RFC3339,
					},
				},
				&StringArg{
					Name:   "sa",
					Min:    1,
					Max:    3,
					Values: &svals,
				},
			},
		},
	}

	require.Error(t, errors.New("sufficient count of arg sa not provided, given 0 expected 1"), cmd.Run(context.Background(), []string{"foo", "subcmd", "2006-01-02T15:04:05Z"}))

	require.NoError(t, cmd.Run(context.Background(), []string{"foo", "subcmd", "2006-01-02T15:04:05Z", "fubar"}))
	require.Equal(t, time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC), tval)
	require.Equal(t, []string{"fubar"}, svals)

	require.NoError(t, cmd.Run(context.Background(), []string{"foo", "subcmd", "--foo", "100", "2006-01-02T15:04:05Z", "fubar", "some"}))
	require.Equal(t, int64(100), ifval)
	require.Equal(t, time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC), tval)
	require.Equal(t, []string{"fubar", "some"}, svals)
}

func TestArgsUsage(t *testing.T) {
	arg := &IntArg{
		Name: "ia",
		Min:  0,
		Max:  1,
	}
	tests := []struct {
		name     string
		min      int
		max      int
		expected string
	}{
		{
			name:     "optional",
			min:      0,
			max:      1,
			expected: "[ia]",
		},
		{
			name:     "zero or more",
			min:      0,
			max:      2,
			expected: "[ia ...]",
		},
		{
			name:     "one",
			min:      1,
			max:      1,
			expected: "ia [ia ...]",
		},
		{
			name:     "many",
			min:      2,
			max:      1,
			expected: "ia [ia ...]",
		},
		{
			name:     "many2",
			min:      2,
			max:      0,
			expected: "ia [ia ...]",
		},
		{
			name:     "unlimited",
			min:      2,
			max:      -1,
			expected: "ia [ia ...]",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			arg.Min, arg.Max = test.min, test.max
			require.Equal(t, test.expected, arg.Usage())
		})
	}
}
