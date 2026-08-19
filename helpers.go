package vapor

import "io"

func ReadIfNotNil(reader io.ReadCloser) ([]byte, error) {
	if reader == nil {
		return nil, nil
	}

	return io.ReadAll(reader)
}
