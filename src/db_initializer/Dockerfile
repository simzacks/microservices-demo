FROM python:3.14-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends gcc libpq-dev && rm -rf /var/lib/apt/lists/*
RUN pip install --no-cache-dir psycopg2-binary

COPY seed_db.py .
# copy this file from the productcatalogservice dir, docker can't copy from an dir outside of its context.
COPY products.json .
CMD ["python", "seed_db.py"]