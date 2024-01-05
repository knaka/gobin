//go:build !production

package utils

func Assert(b bool, messages ...string) {
	if b {
		return
	}
	panic(TernaryF(len(messages) > 0,
		func() string { return messages[0] },
		func() string { return "" }),
	)
}
