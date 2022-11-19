package frostutil

import (
	"math"
	"strings"
)

// Min returns whichever of x or y is lowest.
func Min[T uint | int | uint8 | uint16 | uint32 | uint64 | int8 | int16 | int32 | int64 | float32 | float64](x, y T) T {
	if x < y {
		return x
	} else {
		return y
	}
}

// Max returns whichever of x or y is highest.
func Max[T uint | int | uint8 | uint16 | uint32 | uint64 | int8 | int16 | int32 | int64 | float32 | float64](x, y T) T {
	if x > y {
		return x
	} else {
		return y
	}
}

// Abs returns the absolute value of x.
func Abs[T int | int8 | int16 | int32 | int64 | float32 | float64](x T) T {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

// Split splits a string by the separator sep, but does not split it where a separator is escaped (prefixed with a \).
// It unescapes escaped separators (removes the \ before them) and turns "\\n"s into '\n's in the output (that is, it unescapes endlines).
func Split(s string, sep rune) (out []string) {
	out = make([]string, strings.Count(s, string(sep))+1)
	amt := 0
	rs := []rune(s)
	start := 0
	bs := false
	curString := strings.Builder{}
	lastWasSep := false
	for i, r := range rs {
		lastWasSep = false
		if r == 'n' && bs {
			curString.WriteString(string(rs[start : i-1]))
			start = i + 1
			bs = false
			curString.WriteRune('\n')
		} else if r == sep {
			if bs {
				curString.WriteString(string(rs[start : i-1]))
				start = i
				bs = false
			} else {
				if curString.Len() > 0 {
					curString.WriteString(string(rs[start:i]))
					out[amt] = curString.String()
					curString.Reset()
				} else {
					out[amt] = string(rs[start:i])
				}
				lastWasSep = true
				start = i + 1
				amt++
			}
		} else if r == '\\' {
			bs = true
		} else {
			bs = false
		}
	}
	if start >= len(s) {
		if curString.Len() > 0 {
			out[amt] = curString.String()
			curString.Reset()
			amt++
		}
		if lastWasSep {
			out[amt] = ""
			amt++
		}
	} else {
		if curString.Len() > 0 {
			curString.WriteString(string(rs[start:]))
			out[amt] = curString.String()
			curString.Reset()
		} else {
			out[amt] = string(rs[start:])
		}
		amt++
	}
	out = out[:amt]
	return
}

// Joins a set of strings, placing separators (sep) between them, escaping any separators (sep) or endlines in xs (turning '\n's into "\\n"s).
// The joined/modified string is returned, and the original slice of strings is unmodified.
func Join(xs []string, sep rune) string {
	sb := strings.Builder{}
	for ix, x := range xs {
		if ix > 0 {
			// place separator before each x except the first.
			sb.WriteRune(sep)
		}
		rx := []rune(x)
		start := 0
		for ir, r := range rx {
			if r == '\n' {
				// turn the \n into \\n
				if ir > start {
					sb.WriteString(string(rx[start:ir]))
				}
				sb.WriteRune('\\')
				sb.WriteRune('n')
				start = ir + 1
			} else if r == sep {
				// put a \ before the separator
				if ir > start {
					sb.WriteString(string(rx[start:ir]))
				}
				sb.WriteRune('\\')
				sb.WriteRune(r)
				start = ir + 1
			}
		}
		if start < len(x) {
			sb.WriteString(string(rx[start:]))
		}
	}
	return sb.String()
}

// UnescapeStr unescapes "\\n" and "\<sep>" back into '\n' and '<sep>'.
func UnescapeStr(x string, sep rune) string {
	sb := strings.Builder{}
	rx := []rune(x)
	start := 0
	bs := false
	for ir, r := range rx {
		if r == '\\' {
			bs = true
		} else if r == 'n' && bs {
			if ir-1 > start {
				sb.WriteString(string(rx[start : ir-1]))
			}
			sb.WriteRune('\n')
			start = ir + 1
			bs = false
		} else if r == sep && bs {
			if ir-1 > start {
				sb.WriteString(string(rx[start : ir-1]))
			}
			start = ir
			bs = false
		} else if bs {
			bs = false
		}
	}
	if start < len(x) {
		sb.WriteString(string(rx[start:]))
	}
	return sb.String()
}

// EscapeStr escapes '\n' and '<sep>' into "\\n" and "\<sep>".
func EscapeStr(x string, sep rune) string {
	sb := strings.Builder{}
	rx := []rune(x)
	start := 0
	for ir, r := range rx {
		if r == '\n' {
			if ir > start {
				sb.WriteString(string(rx[start:ir]))
			}
			sb.WriteRune('\\')
			sb.WriteRune('n')
			start = ir + 1
		} else if r == sep {
			if ir > start {
				sb.WriteString(string(rx[start:ir]))
			}
			sb.WriteRune('\\')
			sb.WriteRune(sep)
			start = ir + 1
		}
	}
	if start < len(x) {
		sb.WriteString(string(rx[start:]))
	}
	return sb.String()
}

// SearchStringsUnsorted searches for x in an unsorted slice of strings and returns its index. If x is not in ax, it returns defaultValue.
func SearchStringsUnsorted(ax []string, x string, defaultValue int) int {
	for i, a := range ax {
		if a == x {
			return i
		}
	}
	return defaultValue
}

// returns a xor b, because Go lacks a logical xor operator
func Xor(a, b bool) (c bool) {
	if a && b {
		c = false
	} else if a || b {
		c = true
	} // there's a third possibility: when a and b are both false, in which case we return false
	return
}

// Converts degrees to radians
func DegreesToRadians(deg float64) (rad float64) {
	return (deg * math.Pi) / 180.0
}

// Converts radians to degrees
func RadiansToDegrees(rad float64) (deg float64) {
	return (rad * 180.0) / math.Pi
}

// CenterString horizontally centers a string by padding it with spaces, if its length is less than desiredLen.
// If len(str) < desiredLen, the returned string will have a length of desiredLen, and either have an equal number of spaces on both sides,
// or one additional space on the right side.
// If len(str) >= desiredLen, then str will be returned and no padding will have been added.
func CenterString(str string, desiredLen int) (ret string) {
	ret = str
	diff := desiredLen - len(str)
	if diff > 0 { // that is, if len(str) < desiredLen {
		sb := strings.Builder{}
		sb.Grow(desiredLen)
		padAmt := diff >> 1 // pad each side by half of diff, rounded down
		// pad the left side
		for i := 0; i < padAmt; i++ {
			sb.WriteRune(' ')
		}
		// add the string
		sb.WriteString(str)
		// pad the right side
		for i := 0; i < padAmt; i++ {
			sb.WriteRune(' ')
		}
		// if diff is odd, we need to add an odd number of spaces, and so we add an additional space on the end.
		if diff&1 == 1 {
			sb.WriteRune(' ')
		}
		ret = sb.String()
	}
	return
}

// Compose takes two functions f and g with a single parameter and return value of the same type, and returns a function that
// returns f(g(t)), where t is a parameter of that shared type.
func Compose[T any](f func(T) T, g func(T) T) func(T) T {
	return func(t T) T {
		return f(g(t))
	}
}

// Map returns a new slice ys created from the results of calling the function f on each element x of the slice xs.
// If you have a function that takes multiple arguments, you can use Map by using partial application, which you can do by
// writing a function (let's call it "foo") that takes your multi-argument function (let's call it "bar") along with the other arguments, and which returns a function
// that takes a single argument and calls bar with that argument along with the arguments which you passed to foo, and returns whatever bar returns.
// e.g.
//
//	 func AddThree(x, y, z int) int { return x + y + z }
//	 func PartialAddThree(y, z int) func(int) int {
//		return func(x int) int {
//			return AddThree(x, y, z)
//		}
//	 }
//	 Map(PartialAddThree(10, 100), []int {1, 2, 3, 4, 5})
//
// There's an example of partial application implemented as a generic function in Partial2to1, but it is probably harder to understand than the example above.
func Map[T, S any](f func(T) S, xs []T) (ys []S) {
	ys = make([]S, len(xs))
	for i, x := range xs {
		ys[i] = f(x)
	}
	return
}

// Partial application from two parameters to one.
// Currently this is only here for example purposes.
// There would need to be one of these functions for each combination of input and output function parameter counts for which partial application was needed, so
// it is probably more useful to write functions like this on a case by case basis in the rare event that one is needed.
// Perhaps the most likely reason one would be needed would be to be able to use the Map function with a function that takes more than one argument.
func Partial2to1[T, S, R any](f func(T, R) S, r R) func(T) S {
	return func(t T) S {
		return f(t, r)
	}
}
