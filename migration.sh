#!/bin/sh
export DATABASE_URI=postgresql://grpc_test:grpc_test@simple_ecommerce_postgresql/grpc_test
python -m alembic upgrade head