FROM python:3.14-slim

WORKDIR /app

RUN apt-get update && apt-gewt install -y --no-install-recommends gcc libpq-dev && rm -rf /var/lib/apt/lists/*
RUN pip install --no-cache-dir psycopg2-binary

COPY seed_products.py .
COPY ../productcatalogservice/products.json .
CMD ["python", "seed_products.py"]