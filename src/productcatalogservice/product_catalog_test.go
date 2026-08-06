// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"database/sql"
	"testing"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/productcatalogservice/genproto"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Mock struct for sql.DB to allow testing
type mockDB struct {
	mockQuery func(context.Context, string, ...interface{}) (*sql.Rows, error)
}

func (m *mockDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return m.mockQuery(ctx, query, args...)
}

// Mock struct for sql.Rows to allow testing
type mockRows struct {
	mockNext func() bool
	mockScan func(...interface{}) error
}

func (m *mockRows) Next() bool {
	return m.mockNext()
}

func (m *mockRows) Scan(dest ...interface{}) error {
	return m.mockScan(dest...)
}

func TestListProducts(t *testing.T) {
	ctx := context.Background()
	db := &mockDB{
		mockQuery: func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
			rows := &mockRows{
				mockNext: func() bool { return true },
				mockScan: func(dest ...interface{}) error {
					id := "1"
					name := "Product 1"
					description := "Description of Product 1"
					picture := ""
					priceUsd := int64(100)
					priceNanos := int32(0)
					categories := pq.StringArray{"Category A", "Category B"}
					return rows.Scan(&id, &name, &description, &picture, &priceUsd, &priceNanos, categories)
				},
			}
			return rows, nil
		},
	}

	catalog := NewProductCatalogServer(db)
	response, err := catalog.ListProducts(ctx, &pb.Empty{})
	if err != nil {
		t.Errorf("ListProducts() error = %v", err)
		return
	}

	expectedProduct := &pb.Product{
		Id:          "1",
		Name:        "Product 1",
		Description: "Description of Product 1",
		Picture:     "",
		PriceUsd:    &pb.Money{CurrencyCode: "USD", Units: 1, Nanos: 0},
		Categories:  []string{"Category A", "Category B"},
	}
	if len(response.Products) != 1 {
		t.Errorf("ListProducts() returned %v products, want 1", len(response.Products))
		return
	}

	if !proto.Equal(expectedProduct, response.Products[0]) {
		t.Errorf("ListProducts() = %v, want %v", response.Products[0], expectedProduct)
	}
}

func TestGetProduct(t *testing.T) {
	ctx := context.Background()
	db := &mockDB{
		mockQueryRow: func(ctx context.Context, query string, args ...interface{}) *sql.Row {
			row := &mockRow{
				mockScan: func(dest ...interface{}) error {
					id := "1"
					name := "Product 1"
					description := "Description of Product 1"
					picture := ""
					priceUsd := int64(100)
					priceNanos := int32(0)
					categories := pq.StringArray{"Category A", "Category B"}
					return row.Scan(&id, &name, &description, &picture, &priceUsd, &priceNanos, categories)
				},
			}
			return row
		},
	}

	catalog := NewProductCatalogServer(db)
	response, err := catalog.GetProduct(ctx, &pb.GetProductRequest{Id: "1"})
	if err != nil {
		t.Errorf("GetProduct() error = %v", err)
		return
	}

	expectedProduct := &pb.Product{
		Id:          "1",
		Name:        "Product 1",
		Description: "Description of Product 1",
		Picture:     "",
		PriceUsd:    &pb.Money{CurrencyCode: "USD", Units: 1, Nanos: 0},
		Categories:  []string{"Category A", "Category B"},
	}
	if !proto.Equal(expectedProduct, response) {
		t.Errorf("GetProduct() = %v, want %v", response, expectedProduct)
	}
}

func TestSearchProducts(t *testing.T) {
	ctx := context.Background()
	db := &mockDB{
		mockQuery: func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
			rows := &mockRows{
				mockNext: func() bool { return true },
				mockScan: func(dest ...interface{}) error {
					id := "1"
					name := "Product 1"
					description := "Description of Product 1"
					picture := ""
					priceUsd := int64(100)
					priceNanos := int32(0)
					categories := pq.StringArray{"Category A", "Category B"}
					return rows.Scan(&id, &name, &description, &picture, &priceUsd, &priceNanos, categories)
				},
			}
			return rows, nil
		},
	}

	catalog := NewProductCatalogServer(db)
	response, err := catalog.SearchProducts(ctx, &pb.SearchProductsRequest{Query: "Product"})
	if err != nil {
		t.Errorf("SearchProducts() error = %v", err)
		return
	}

	expectedProduct := &pb.Product{
		Id:          "1",
		Name:        "Product 1",
		Description: "Description of Product 1",
		Picture:     "",
		PriceUsd:    &pb.Money{CurrencyCode: "USD", Units: 1, Nanos: 0},
		Categories:  []string{"Category A", "Category B"},
	}
	if len(response.Results) != 1 {
		t.Errorf("SearchProducts() returned %v products, want 1", len(response.Results))
		return
	}

	if !proto.Equal(expectedProduct, response.Results[0]) {
		t.Errorf("SearchProducts() = %v, want %v", response.Results[0], expectedProduct)
	}
}