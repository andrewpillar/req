// Code generated by "stringer -type Token -linecomment"; DO NOT EDIT.

package token

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[EOF-1]
	_ = x[Name-2]
	_ = x[Literal-3]
	_ = x[Semi-4]
	_ = x[Comma-5]
	_ = x[Colon-6]
	_ = x[Dot-7]
	_ = x[DotDot-8]
	_ = x[Arrow-9]
	_ = x[Assign-10]
	_ = x[Ref-11]
	_ = x[Lbrace-12]
	_ = x[Rbrace-13]
	_ = x[Lbrack-14]
	_ = x[Rbrack-15]
	_ = x[If-16]
	_ = x[Else-17]
	_ = x[Match-18]
	_ = x[Range-19]
	_ = x[Yield-20]
	_ = x[Open-21]
	_ = x[Env-22]
	_ = x[Exit-23]
	_ = x[Write-24]
	_ = x[HEAD-25]
	_ = x[OPTIONS-26]
	_ = x[GET-27]
	_ = x[POST-28]
	_ = x[PUT-29]
	_ = x[PATCH-30]
	_ = x[DELETE-31]
}

const _Token_name = "eofnameliteralsemi or newline,:...->=${}[]ifelsematchrangeyieldopenenvexitwriteHEADOPTIONSGETPOSTPUTPATCHDELETE"

var _Token_index = [...]uint8{0, 3, 7, 14, 29, 30, 31, 32, 34, 36, 37, 38, 39, 40, 41, 42, 44, 48, 53, 58, 63, 67, 70, 74, 79, 83, 90, 93, 97, 100, 105, 111}

func (i Token) String() string {
	i -= 1
	if i >= Token(len(_Token_index)-1) {
		return "Token(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _Token_name[_Token_index[i]:_Token_index[i+1]]
}