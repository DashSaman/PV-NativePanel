package httpapi

import "net/http"

func (s *server) customerExtraHandler(name string) http.Handler {
	if s.config.CustomerService == nil {
		return nil
	}
	switch name {
	case "customers.index", "users.index":
		return http.HandlerFunc(s.listProductCustomersAccounting)
	case "customers.create", "users.create":
		return http.HandlerFunc(s.createProductCustomer)
	case "users.update":
		return http.HandlerFunc(s.productUserUpdate)
	case "users.suspend":
		return http.HandlerFunc(s.suspendCustomer)
	case "users.resume":
		return http.HandlerFunc(s.resumeCustomer)
	case "users.revoke":
		return http.HandlerFunc(s.deleteCustomer)
	case "users.renew":
		return http.HandlerFunc(s.renewProductCustomer)
	case "users.subscription.current":
		return http.HandlerFunc(s.currentCustomerSubscription)
	case "users.subscription.rotate":
		return http.HandlerFunc(s.rotateCustomerSubscription)
	case "users.password.rotate":
		return http.HandlerFunc(s.rotateCustomerPassword)
	case "users.service.update":
		return http.HandlerFunc(s.updateCustomerService)
	case "users.volume.add", "customers.volume.add":
		return http.HandlerFunc(s.addCustomerVolume)
	case "users.validity.extend", "customers.validity.extend":
		return http.HandlerFunc(s.extendCustomerTime)
	case "users.usage.reset", "customers.usage.reset":
		return http.HandlerFunc(s.resetCustomerUsage)
	case "plans.index":
		return http.HandlerFunc(s.listProductPlans)
	case "plans.create":
		return http.HandlerFunc(s.createProductPlan)
	case "customer-groups.index":
		return http.HandlerFunc(s.listCustomerGroups)
	case "customer-groups.create":
		return http.HandlerFunc(s.createCustomerGroup)
	case "customer-tags.index":
		return http.HandlerFunc(s.listCustomerTags)
	case "customer-tags.create":
		return http.HandlerFunc(s.createCustomerTag)
	case "users.bulk.preview":
		return http.HandlerFunc(s.previewCustomerBulk)
	case "users.bulk.execute":
		return http.HandlerFunc(s.executeCustomerBulk)
	default:
		return nil
	}
}
