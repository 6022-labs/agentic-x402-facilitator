package helpers

import "strings"

// SendErrorKind classifies a node's answer to eth_sendRawTransaction.
type SendErrorKind int

const (
	// SendErrorUnknown may hide an accepted transaction (timeout, dropped connection): treat it as broadcast.
	SendErrorUnknown SendErrorKind = iota
	// SendErrorNonceTaken means another transaction already uses the nonce: skip past it.
	SendErrorNonceTaken
	// SendErrorRejected means the node refused the transaction outright and the nonce is still free.
	SendErrorRejected
	// SendErrorAlreadyKnown means the node already holds this very transaction.
	SendErrorAlreadyKnown
)

// ClassifySendError reads the geth-compatible error text, the only signal JSON-RPC nodes give for these cases.
func ClassifySendError(err error) SendErrorKind {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "already known"):
		return SendErrorAlreadyKnown
	case strings.Contains(message, "nonce too low"),
		strings.Contains(message, "replacement transaction underpriced"):
		return SendErrorNonceTaken
	case strings.Contains(message, "insufficient funds"),
		strings.Contains(message, "intrinsic gas too low"),
		strings.Contains(message, "exceeds block gas limit"),
		strings.Contains(message, "max fee per gas less than block base fee"),
		strings.Contains(message, "transaction underpriced"),
		strings.Contains(message, "exceeds the configured cap"):
		return SendErrorRejected
	default:
		return SendErrorUnknown
	}
}
