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
	"strings"
	"time"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/productcatalogservice/genproto"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

// Start changes for PGSQL
    "database/sql"
	"github.com/lib/pq"
)

// Replace the old, empty productCatalog struct with this database-connected version:
type productCatalog struct {
	db *sql.DB
	catalog pb.ListProductsResponse // using this for the test feature when db is nil.
	pb.UnimplementedProductCatalogServiceServer
}

// Add this constructor method right underneath the struct:
func NewProductCatalogServer(db *sql.DB) *productCatalog {
	return &productCatalog{db: db}
}
// End definition for PGSQL

func (p *productCatalog) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (p *productCatalog) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

/*func (p *productCatalog) ListProducts(context.Context, *pb.Empty) (*pb.ListProductsResponse, error) {
	time.Sleep(extraLatency)

	return &pb.ListProductsResponse{Products: p.parseCatalog()}, nil
}

func (p *productCatalog) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	time.Sleep(extraLatency)

	catalog := p.parseCatalog()
	for _, product := range catalog {
		if req.Id == product.Id {
			return product, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "no product with ID %s", req.Id)
}

func (p *productCatalog) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	time.Sleep(extraLatency)

	var ps []*pb.Product
	for _, product := range p.parseCatalog() {
		if strings.Contains(strings.ToLower(product.Name), strings.ToLower(req.Query)) ||
			strings.Contains(strings.ToLower(product.Description), strings.ToLower(req.Query)) {
			ps = append(ps, product)
		}
	}

	return &pb.SearchProductsResponse{Results: ps}, nil
}
*/

// function used for testing
func (p *productCatalog) parseCatalog() []*pb.Product {
	if reloadCatalog || len(p.catalog.Products) == 0 {
		err := loadCatalog(&p.catalog)
		if err != nil {
			return []*pb.Product{}
		}
	}

	return p.catalog.Products
}


func (s *productCatalog) ListProducts(ctx context.Context, req *pb.Empty) (*pb.ListProductsResponse, error) {
    if s.db != nil {
		rows, err := s.db.QueryContext(ctx, "SELECT id, name, description, picture, price_usd, price_nanos, categories FROM products")
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to query products: %v", err)
		}
		defer rows.Close()

		var products []*pb.Product
		for rows.Next() {
			p := &pb.Product{PriceUsd: &pb.Money{}}
			var categories []string

			err := rows.Scan(&p.Id, &p.Name, &p.Description, &p.Picture, &p.PriceUsd.CurrencyCode, &p.PriceUsd.Nanos, pq.Array(&categories))
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to scan product row: %v", err)
			}
			// Match exact proto fields
			p.PriceUsd.CurrencyCode = "USD" 
			p.Categories = categories
			products = append(products, p)
		}

		return &pb.ListProductsResponse{Products: products}, nil
	}
	return &s.catalog, nil
}

func (s *productCatalog) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
    if s.db != nil {
		p := &pb.Product{PriceUsd: &pb.Money{CurrencyCode: "USD"}}
		var categories []string

		query := "SELECT id, name, description, picture, price_usd, price_nanos, categories FROM products WHERE id = $1"
		err := s.db.QueryRowContext(ctx, query, req.Id).Scan(
			&p.Id, &p.Name, &p.Description, &p.Picture, &p.PriceUsd.Units, &p.PriceUsd.Nanos, pq.Array(&categories),
		)

		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "product not found: %s", req.Id)
		} else if err != nil {
			return nil, status.Errorf(codes.Internal, "database lookup failed: %v", err)
		}

		p.Categories = categories
		return p, nil
	}
	//code for test
	catalog := p.parseCatalog()
	for _, product := range catalog {
		if req.Id == product.Id {
			return product, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "no product with ID %s", req.Id)

}

func (s *productCatalog) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
    if s.db != nil {
		// Simple SQL ILIKE search logic across Name or Description strings
		query := `
			SELECT id, name, description, picture, price_usd, price_nanos, categories 
			FROM products 
			WHERE name ILIKE $1 OR description ILIKE $1`
		
		rows, err := s.db.QueryContext(ctx, query, "%"+req.Query+"%")
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed searching products: %v", err)
		}
		defer rows.Close()

		var products []*pb.Product
		for rows.Next() {
			p := &pb.Product{PriceUsd: &pb.Money{CurrencyCode: "USD"}}
			var categories []string

			err := rows.Scan(&p.Id, &p.Name, &p.Description, &p.Picture, &p.PriceUsd.Units, &p.PriceUsd.Nanos, pq.Array(&categories))
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed scanning search match: %v", err)
			}
			p.Categories = categories
			products = append(products, p)
		}

		return &pb.SearchProductsResponse{Results: products}, nil
    }
	var ps []*pb.Product
	for _, product := range p.parseCatalog() {
		if strings.Contains(strings.ToLower(product.Name), strings.ToLower(req.Query)) ||
			strings.Contains(strings.ToLower(product.Description), strings.ToLower(req.Query)) {
			ps = append(ps, product)
		}
	}

	return &pb.SearchProductsResponse{Results: ps}, nil

}
