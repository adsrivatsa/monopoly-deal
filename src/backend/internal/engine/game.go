package engine

type Game interface {
	JSON() ([]byte, error)
	Load(data []byte) error
}
