package frostutil

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Max(t *testing.T) {
	assert.Equal(t, Max(10, 5), 10)
	assert.Equal(t, Max(5, 10), 10)
	assert.Equal(t, Max(28957985, -293847), 28957985)
	assert.Equal(t, Max(0, 0), 0)
}

func Test_Min(t *testing.T) {
	assert.Equal(t, Min(10, 5), 5)
	assert.Equal(t, Min(5, 10), 5)
	assert.Equal(t, Min(28957985, -293847), -293847)
	assert.Equal(t, Min(0, 0), 0)
}

func Test_Abs(t *testing.T) {
	assert.Equal(t, Abs(42), 42)
	assert.Equal(t, Abs(-42), 42)
	assert.Equal(t, Abs(0), 0)
}

func Test_Split(t *testing.T) {
	assert.Equal(t, []string{"foo", "bar", "", "narf"}, Split(`foo,bar,,narf`, ','))
	assert.Equal(t, []string{"foo", "bar,zort\npoink", "", "narf"}, Split(`foo,bar\,zort\npoink,,narf`, ','))
}

func Test_Join(t *testing.T) {
	assert.Equal(t, `foo,bar,,narf`, Join([]string{"foo", "bar", "", "narf"}, ','))
	assert.Equal(t, `foo,bar\,zort,,narf`, Join([]string{"foo", "bar,zort", "", "narf"}, ','))
	assert.Equal(t, `foo,bar\,zort\npoink,,narf`, Join([]string{"foo", "bar,zort\npoink", "", "narf"}, ','))
}

func Test_EscapeStr(t *testing.T) {
	assert.Equal(t, "foo\\,bar\\nnarf", EscapeStr("foo,bar\nnarf", ','))
}

func Test_UnescapeStr(t *testing.T) {
	assert.Equal(t, "foo,bar\nnarf", UnescapeStr("foo\\,bar\\nnarf", ','))
}

func Test_Xor(t *testing.T) {
	tas := assert.New(t)
	tas.Equal(false, Xor(false, false))
	tas.Equal(true, Xor(true, false))
	tas.Equal(true, Xor(false, true))
	tas.Equal(false, Xor(true, true))
}

func Test_DegreesToRadians(t *testing.T) {
	tas := assert.New(t)
	testA := []float64{45, 90, 180, 270, 360}
	expectedA := []float64{math.Pi / 4.0, math.Pi / 2.0, math.Pi, math.Pi * 1.5, math.Pi * 2.0}

	for i, x := range testA {
		tas.Equal(expectedA[i], DegreesToRadians(x))
	}
}

func Test_RadiansToDegrees(t *testing.T) {
	tas := assert.New(t)
	testA := []float64{math.Pi / 4.0, math.Pi / 2.0, math.Pi, math.Pi * 1.5, math.Pi * 2.0}
	expectedA := []float64{45, 90, 180, 270, 360}

	for i, x := range testA {
		tas.Equal(expectedA[i], RadiansToDegrees(x))
	}
}

func Test_CenterString(t *testing.T) {
	tas := assert.New(t)
	tas.Equal("   123   ", CenterString("123", 9))
	tas.Equal("   123    ", CenterString("123", 10))
	tas.Equal("123 ", CenterString("123", 4))
	tas.Equal("123", CenterString("123", 2))
}

func Test_Compose(t *testing.T) {
	tas := assert.New(t)
	testA := []int{1, 2, 3, 4, 5}
	expectedA := []int{2, 8, 18, 32, 50}

	add := func(x int) int { return x + x }
	multiply := func(x int) int { return x * x }
	for i, x := range testA {
		tas.Equal(expectedA[i], Compose(add, multiply)(x))
	}
}

func AddThree(x, y, z int) int { return x + y + z }

func PartialAddThree(y, z int) func(int) int {
	return func(x int) int {
		return AddThree(x, y, z)
	}
}

func Test_Map(t *testing.T) {
	tas := assert.New(t)
	testA := []int{1, 2, 3, 4, 5}
	expectedA := []int{2, 4, 6, 8, 10}
	testB := []float64{45.0, 90.0, 180.0}
	expectedB := []float64{math.Sqrt(2) / 2.0, 1.0, 0.0}
	testC := []string{"Cat", "Dog", "Snek"}
	expectedC := []string{"cat", "dog", "snek"}
	expectedD := []int{111, 112, 113, 114, 115}

	tas.Equal(expectedA, Map(func(x int) int { return x + x }, testA))
	tas.InDeltaSlice(expectedB, Map(Compose(math.Sin, DegreesToRadians), testB), 1e-14)
	tas.Equal(expectedC, Map(strings.ToLower, testC))
	tas.Equal(expectedD, Map(PartialAddThree(10, 100), []int{1, 2, 3, 4, 5}))
}
