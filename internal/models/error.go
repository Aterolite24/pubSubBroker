package models

type BrokerError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *BrokerError) Error() string {
	return e.Message
}
