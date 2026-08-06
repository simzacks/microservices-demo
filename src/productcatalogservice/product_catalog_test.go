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
	"testing"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/productcatalogservice/genproto"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestListProducts(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, name, description, picture, price_usd, price_nanos, categories FROM products").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "picture", "price_usd", "price_nanos", "categories"}).
			AddRow("1", "Product 1", "Description of Product 1", "", 100, 0, pq.StringArray{"Category A", "Category B"}))
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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, name, description, picture, price_usd, price_nanos, categories FROM products WHERE id = $1").
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "picture", "price_usd", "price_nanos", "categories"}).
			AddRow("1", "Product 1", "Description of Product 1", "", 100, 0, pq.StringArray{"Category A", "Category B"}))

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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, name, description, picture, price_usd, price_nanos, categories FROM products WHERE name ILIKE $1 OR description ILIKE $1").
		WithArgs("%Product%").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "picture", "price_usd", "price_nanos", "categories"}).
			AddRow("1", "Product 1", "Description of Product 1", "", 100, 0, pq.StringArray{"Category A", "Category B"}))

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
