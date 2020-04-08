package main

// my parser without slices
func parseIp4(s string) uint32 {
	var ip, n uint32 = 0, 0
	var r uint = 24
	for i := 0; i < len(s); i++ {
		switch {
		case '0' <= s[i] && s[i] <= '9':
			n = n*10 + uint32(s[i]-'0')
			if n > 0xFF {
				//Debug.Printf("Bad IP (1) n=%d: %s\n", n, s)
				return 0xFFFFFFFF
			}
		case s[i] == '.':
			if r != 0 {
				ip += (n << r)
			} else {
				//Debug.Printf("Bad IP (2): %s\n", s)
				return 0xFFFFFFFF
			}
			r -= 8
			n = 0
		default:
			//Debug.Printf("Bad IP (3): %s\n", s)
			return 0xFFFFFFFF
		}
	}
	if r != 0 {
		//Debug.Printf("Bad IP (4): %s\n", s)
		return 0xFFFFFFFF
	}
	ip += n
	return ip
}

func int2Ip4(ip uint32) string {
	var (
		b [15]byte
		c int
	)
	for i := 24; i >= 0; i -= 8 {
		d := int((ip >> i) & 0x000000FF)
		for j := 100; j > 0; j /= 10 {
			t := byte((d / j) + '0')
			d %= j
			if (t > '0') || ((c != 0) && (b[c-1] != '.')) || (j == 1) {
				b[c] = t
				c++
			}
		}
		if i > 0 {
			b[c] = '.'
			c++
		}
	}
	return string(b[:c])
}
