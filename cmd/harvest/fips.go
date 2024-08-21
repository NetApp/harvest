//go:build boringcrypto

package main

import (
	_ "crypto/tls/fipsonly"
)
