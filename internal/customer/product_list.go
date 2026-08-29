package customer

import (
	"context"
	"database/sql"
	"errors"
)

type customerProductListStore interface {
	SearchCustomersTx(context.Context, *sql.Tx, CustomerListQuery) (CustomerPage, error)
}

func (s *Service) SearchCustomers(ctx context.Context, tx *sql.Tx, query CustomerListQuery) (CustomerPage, error) {
	if s == nil || s.store == nil || tx == nil {
		return CustomerPage{}, errors.New("customer: search capability is unavailable")
	}
	query = query.Normalize()
	if store, ok := s.store.(customerProductListStore); ok {
		return store.SearchCustomersTx(ctx, tx, query)
	}
	legacy, err := s.ListCustomers(ctx, tx)
	if err != nil {
		return CustomerPage{}, err
	}
	start := (query.Page - 1) * query.PageSize
	if start >= len(legacy) {
		return CustomerPage{Customers: []CustomerView{}, Page: query.Page, PageSize: query.PageSize, Total: int64(len(legacy))}, nil
	}
	end := start + query.PageSize
	if end > len(legacy) {
		end = len(legacy)
	}
	return CustomerPage{Customers: legacy[start:end], Page: query.Page, PageSize: query.PageSize, Total: int64(len(legacy))}, nil
}
