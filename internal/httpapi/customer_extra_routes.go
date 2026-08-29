package httpapi

import "net/http"

func init() {
	Routes = append(Routes,
		Route{"POST", "/api/v1/customers/{id}/volume/add", "customers.volume.add", Owner, true},
		Route{"POST", "/api/v1/customers/{id}/validity/extend", "customers.validity.extend", Owner, true},
	)
}

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
