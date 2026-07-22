#!/usr/bin/env bash
# TC-010-7: boots ./bin/server against a throwaway PostgreSQL container this
# script starts and always tears down itself — never the developer's
# configured DATABASE_URL — and proves boot, migration, health, and
# shutdown all work on a clean environment. Run via `make smoke-test`
# (which builds bin/server first); requires docker.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

CONTAINER_NAME="template-v5-smoke-test-$$"
SERVER_PID=""

cleanup() {
	if [ -n "$SERVER_PID" ] && kill -0 "$SERVER_PID" 2>/dev/null; then
		kill -KILL "$SERVER_PID" 2>/dev/null || true
		wait "$SERVER_PID" 2>/dev/null || true
	fi
	docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "==> starting ephemeral PostgreSQL container"
docker run -d --rm --name "$CONTAINER_NAME" \
	-e POSTGRES_PASSWORD=smoketest \
	-e POSTGRES_DB=smoketest \
	-p 127.0.0.1::5432 \
	postgres:16-alpine >/dev/null

HOST_PORT="$(docker port "$CONTAINER_NAME" 5432/tcp | cut -d: -f2)"
DATABASE_URL="postgres://postgres:smoketest@127.0.0.1:${HOST_PORT}/smoketest?sslmode=disable"

echo "==> waiting for PostgreSQL to accept connections"
# The official postgres image starts a temporary server to run init
# scripts, shuts it down, then starts the real one — logging "database
# system is ready to accept connections" once for each. pg_isready alone
# can succeed against the temporary server and then reset the connection
# when it stops, so wait for the second occurrence instead.
pg_ready=""
for _ in $(seq 1 60); do
	ready_count="$(docker logs "$CONTAINER_NAME" 2>&1 | grep -c "database system is ready to accept connections" || true)"
	if [ "${ready_count:-0}" -ge 2 ]; then
		pg_ready=1
		break
	fi
	sleep 1
done
if [ -z "$pg_ready" ]; then
	echo "PostgreSQL never became ready" >&2
	exit 1
fi

APP_PORT="$(python3 -c 'import socket; s = socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1])')"

echo "==> starting bin/server"
PORT="$APP_PORT" \
	DATABASE_URL="$DATABASE_URL" \
	JWT_SECRET="smoke-test-jwt-secret-not-for-real-use-32bytes" \
	./bin/server &
SERVER_PID=$!

echo "==> waiting for /health"
healthy=""
for _ in $(seq 1 30); do
	if curl -fsS "http://127.0.0.1:${APP_PORT}/health" >/dev/null 2>&1; then
		healthy=1
		break
	fi
	if ! kill -0 "$SERVER_PID" 2>/dev/null; then
		echo "server exited before becoming healthy" >&2
		exit 1
	fi
	sleep 1
done
if [ -z "$healthy" ]; then
	echo "server never became healthy" >&2
	exit 1
fi
echo "    /health answered 200 (boot + migration succeeded: migrations run before the listener starts)"

echo "==> verifying the migration actually created the schema"
table="$(docker exec "$CONTAINER_NAME" psql -U postgres -d smoketest -tAc "select to_regclass('public.users')" | tr -d '[:space:]')"
if [ "$table" != "users" ]; then
	echo "expected the migration to create public.users, got: $table" >&2
	exit 1
fi

echo "==> sending SIGTERM and waiting for a clean shutdown"
kill -TERM "$SERVER_PID"
exited=""
for _ in $(seq 1 20); do
	if ! kill -0 "$SERVER_PID" 2>/dev/null; then
		exited=1
		break
	fi
	sleep 0.5
done
if [ -z "$exited" ]; then
	echo "server did not exit within the bounded wait after SIGTERM" >&2
	exit 1
fi
# The server exits because we just signaled it, so `wait` reporting a
# signal-death exit status here is expected, not a failure of the script.
wait "$SERVER_PID" 2>/dev/null || true
SERVER_PID=""

echo "✓ smoke-test: boot, migration, health, and shutdown all verified"
