package httpapi

import "net/http"

func (s *server) customerExtraHandler(name string) http.Handler {
	if s.config.CustomerService == nil {
		return nil
	}
	switch name {
	case "customers.volume.add":
		return http.HandlerFunc(s.addCustomerVolume)
	case "customers.validity.extend":
		return http.HandlerFunc(s.extendCustomerTime)
	default:
		return nil
	}
}
