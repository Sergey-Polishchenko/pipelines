package utils

// CloseChannels closes all provided channels safely, signaling consumers that no more data will be sent.
func CloseChannels[T any](channels []chan<- T) {
	for _, ch := range channels {
		close(ch)
	}
}
