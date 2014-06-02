package dbr

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInterpolateNil(t *testing.T) {
	args := []interface{}{nil}
	
	str, err := Interpolate("SELECT * FROM x WHERE a = ?", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = NULL")
}

func TestInterpolateInts(t *testing.T) {
	args := []interface{}{
		int(1),
		int8(-2),
		int16(3),
		int32(4),
		int64(5),
		uint(6),
		uint8(7),
		uint16(8),
		uint32(9),
		uint64(10),
	}
	
	str, err := Interpolate("SELECT * FROM x WHERE a = ? AND b = ? AND c = ? AND d = ? AND e = ? AND f = ? AND g = ? AND h = ? AND i = ? AND j = ?", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 1 AND b = -2 AND c = 3 AND d = 4 AND e = 5 AND f = 6 AND g = 7 AND h = 8 AND i = 9 AND j = 10")
}

func TestInterpolateBools(t *testing.T) {
	args := []interface{}{true, false}
	
	str, err := Interpolate("SELECT * FROM x WHERE a = ? AND b = ?", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 1 AND b = 0")
}

func TestInterpolateFloats(t *testing.T) {
	args := []interface{}{float32(0.15625), float64(3.14159)}
	
	str, err := Interpolate("SELECT * FROM x WHERE a = ? AND b = ?", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 0.15625 AND b = 3.14159")
}

func TestInterpolateStrings(t *testing.T) {
	args := []interface{}{"hello", "\"hello's \\ world\" \n\r\x00\x1a"}
	
	str, err := Interpolate("SELECT * FROM x WHERE a = ? AND b = ?", args)
	assert.NoError(t, err)	
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 'hello' AND b = '\\\"hello\\'s \\\\ world\\\" \\n\\r\\x00\\x1a'")
}

func TestInterpolateSlices(t *testing.T) {
	args := []interface{}{[]int{1}, []int{1,2,3}, []uint32{5,6,7}, []string{"wat", "ok"}}
	
	str, err := Interpolate("SELECT * FROM x WHERE a = ? AND b = ? AND c = ? AND d = ?", args)
	assert.NoError(t, err)	
	assert.Equal(t, str, "SELECT * FROM x WHERE a = (1) AND b = (1,2,3) AND c = (5,6,7) AND d = ('wat','ok')")
}

type myString struct {
	Present bool
	Val string
}

func (m myString) Value() interface{} {
	if m.Present {
		return m.Val
	} else {
		return nil
	}
}

func TestIntepolatingValuers(t *testing.T) {
	args := []interface{}{myString{true, "wat"}, myString{false, "fry"}}
	
	str, err := Interpolate("SELECT * FROM x WHERE a = ? AND b = ?", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = wat AND b = NULL")
}

func TestInterpolateErrors(t *testing.T) {
	_, err := Interpolate("SELECT * FROM x WHERE a = ? AND b = ?", []interface{}{1})
	assert.Equal(t, err, ErrArgumentMismatch)
	
	_, err = Interpolate("SELECT * FROM x WHERE", []interface{}{1})
	assert.Equal(t, err, ErrArgumentMismatch)
	
	_, err = Interpolate("SELECT * FROM x WHERE a = ?", []interface{}{string([]byte{0x34, 0xFF, 0xFE})})
	assert.Equal(t, err, ErrNotUTF8)
	
	_, err = Interpolate("SELECT * FROM x WHERE a = ?", []interface{}{struct{}{}})
	assert.Equal(t, err, ErrInvalidValue)
	
	_, err = Interpolate("SELECT * FROM x WHERE a = ?", []interface{}{ []struct{}{struct{}{}, struct{}{}} })
	assert.Equal(t, err, ErrInvalidSliceValue)
}
