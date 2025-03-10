package myerrors

import "errors"

var (
	ErrInvalidMetricType   = errors.New("invalid metric type")
	ErrInvalidGaugeValue   = errors.New("invalid gauge value")
	ErrInvalidCounterValue = errors.New("invalid counter value")
	ErrMetricNotFound      = errors.New("metric not found")
)
