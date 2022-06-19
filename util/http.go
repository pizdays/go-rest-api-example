package util

// IsHTTPCodeClientErr checks if code is classified as client error code.
func IsHTTPCodeClientErr(code int) bool {
	var (
		errCodeStart = 400
		errCodeEnd   = 499
	)

	return code > errCodeStart && code < errCodeEnd
}

// IsHTTPCodeServerErr checks if code is classified as server error code.
func IsHTTPCodeServerErr(code int) bool {
	var (
		errCodeStart = 500
		errCodeEnd   = 599
	)

	return code > errCodeStart && code < errCodeEnd
}
