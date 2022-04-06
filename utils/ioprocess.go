package utils

import "io"

func CustomReadAll(r io.ReadCloser) ([]byte, error) {
	b := make([]byte, 0, 512)
	for {
		if len(b) == cap(b) && cap(b) < 8192 {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}
		if len(b) >= 8192 {
			return b, err
		}
	}

}
