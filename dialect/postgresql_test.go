package dialect

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEncodeTime(t *testing.T) {
	for _, test := range []struct {
		in   time.Time
		want string
	}{
		{
			in:   time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			want: `'2009-11-10T23:00:00Z'`,
		},
		{
			in:   time.Date(2009, time.November, 10, 15, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)),
			want: `'2009-11-10T15:00:00-08:00'`,
		},
		{
			in:   time.Date(2009, time.November, 11, 07, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60)),
			want: `'2009-11-11T07:00:00+08:00'`,
		},
		{
			in:   time.Date(2009, time.November, 10, 23, 45, 59, 123456789, time.UTC),
			want: `'2009-11-10T23:45:59.123456789Z'`,
		},
	} {
		require.Equal(t, test.want, PostgreSQL.EncodeTime(test.in))
	}
}
