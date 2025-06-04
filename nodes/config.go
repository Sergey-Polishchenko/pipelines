package nodes

// Config holds configuration parameters for nodes, such as buffer sizes and worker counts.
type Config struct {
	InBuffer int
	Buffer   int
	Workers  int
}

// DefaultConfig returns a Config with default values: InBuffer=0, Buffer=10, Workers=10.
func DefaultConfig() Config {
	return Config{
		Buffer:  10,
		Workers: 10,
	}
}
