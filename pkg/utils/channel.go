package utils

// ChanSendString Sends a string to a channel (non-blocking)
func ChanSendString(channel chan string, data string) {
	select {
	case channel <- data:
		break
	default:
		break
	}
}
