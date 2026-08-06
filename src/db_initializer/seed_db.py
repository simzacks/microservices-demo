import json
import os
import sys
import psycopg2

def execute_ddl_and_seed():
    # 1. Fallback / Flexible Connection Parsing
    db_url = os.getenv("DATABASE_URL")
    
    if db_url:
        print("Connecting using composite DATABASE_URL parameter...")
        connection_target = db_url
    else:
        print("Assembling connection parameters from environment values...")
        fqdn = os.getenv("DB_FQDN", "BROKEN")
        
        # Build standard connection dictionary parameters
        connection_target = f"{fqdn}"

    json_path = "products.json"
    if not os.path.exists(json_path):
        print(f"Error: {json_path} file not found.")
        sys.exit(1)

    with open(json_path, "r") as f:
        data = json.load(f)
    products = data.get("products", [])

    try:
        conn = psycopg2.connect(connection_target)
        cursor = conn.cursor()

        ddl_query = """
        CREATE TABLE IF NOT EXISTS products (
            id VARCHAR(50) PRIMARY KEY,
            name VARCHAR(100) NOT NULL,
            description TEXT,
            picture VARCHAR(255),
            price_usd BIGINT NOT NULL,
            price_nanos INT NOT NULL,
            categories TEXT[] NOT NULL
        );

        CREATE TABLE IF NOT EXISTS orders (
            order_id VARCHAR(50) PRIMARY KEY,
            customer_name VARCHAR(150) NOT NULL,
            email VARCHAR(100) NOT NULL,
            transaction_id VARCHAR(100) NOT NULL,
            shipping_street VARCHAR(255) NOT NULL,
            shipping_city VARCHAR(100) NOT NULL,
            shipping_state VARCHAR(50),
            shipping_country VARCHAR(100) NOT NULL,
            shipping_zip_code VARCHAR(20) NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS order_items (
            item_id SERIAL PRIMARY KEY,
            order_id VARCHAR(50) REFERENCES orders(order_id) ON DELETE CASCADE,
            product_id VARCHAR(50) NOT NULL,
            quantity INT NOT NULL CHECK (quantity > 0),
            cost_usd BIGINT NOT NULL,
            cost_nanos INT NOT NULL
        );

        CREATE INDEX IF NOT EXISTS idx_products_categories ON products USING gin(categories);
        CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
        """
        print("Executing DDL Schema operations...")
        cursor.execute(ddl_query)

        if products:
            insert_query = """
                INSERT INTO products (id, name, description, picture, price_usd, price_nanos, categories)
                VALUES (%s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (id) DO UPDATE SET
                    name = EXCLUDED.name,
                    description = EXCLUDED.description,
                    picture = EXCLUDED.picture,
                    price_usd = EXCLUDED.price_usd,
                    price_nanos = EXCLUDED.price_nanos,
                    categories = EXCLUDED.categories;
            """
            print(f"Seeding {len(products)} products...")
            for p in products:
                cursor.execute(
                    insert_query,
                    (p["id"], p["name"], p["description"], p["picture"], int(p["priceUsd"]["units"]), p["priceUsd"]["nanos"], p["categories"])
                )
        
        conn.commit()
        print("Database initialized successfully!")

    except Exception as e:
        print(f"Database Initialization failure: {e}")
        sys.exit(1)
    finally:
        if 'cursor' in locals(): cursor.close()
        if 'conn' in locals(): conn.close()

if __name__ == "__main__":
    execute_ddl_and_seed()
