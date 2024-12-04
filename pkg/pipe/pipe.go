package pipe

import "os"

func NewPipe() (*os.File, *os.File, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return r, w, nil
}
