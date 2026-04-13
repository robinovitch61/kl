import json
import logging
import os
import random
import threading
import time
import traceback
from datetime import datetime

import psycopg2
from flask import Flask, jsonify
from flask_cors import CORS
from psycopg2 import pool


class JsonFormatter(logging.Formatter):
    def format(self, record):
        msg = record.getMessage()
        try:
            json.loads(msg)
            return msg
        except (ValueError, TypeError):
            return json.dumps({
                "timestamp": datetime.utcnow().isoformat() + "Z",
                "level": record.levelname,
                "logger": record.name,
                "message": msg,
            })

handler = logging.StreamHandler()
handler.setFormatter(JsonFormatter())
logging.root.setLevel(logging.DEBUG)
logging.root.addHandler(handler)
logger = logging.getLogger(__name__)

app = Flask(__name__)
CORS(app)

db_pool = psycopg2.pool.SimpleConnectionPool(
    1, 20,
    dbname=os.getenv("POSTGRES_DB", "flaskdb"),
    user=os.getenv("POSTGRES_USER", "flaskuser"),
    password=os.getenv("POSTGRES_PASSWORD", "flaskpassword"),
    host=os.getenv("POSTGRES_HOST", "postgres-service")
)

def get_db_connection():
    return db_pool.getconn()

def return_db_connection(conn):
    db_pool.putconn(conn)

def init_db():
    conn = get_db_connection()
    try:
        with conn.cursor() as cur:
            cur.execute("""
            CREATE TABLE IF NOT EXISTS status_hits (
                id SERIAL PRIMARY KEY,
                hits INTEGER NOT NULL
            )
            """)
            cur.execute("""
            INSERT INTO status_hits (hits)
            SELECT 0
            WHERE NOT EXISTS (SELECT 1 FROM status_hits WHERE id = 1)
            """)
        conn.commit()
        logger.info("database initialized successfully")
    except Exception as e:
        logger.error("error initializing database", exc_info=True)
        conn.rollback()
    finally:
        return_db_connection(conn)

def cleanup():
    logger.info("cleaning up resources...")
    db_pool.closeall()

def exit_after_delay():
    delay = int(os.getenv("PERIODIC_KILL_FLASK_SECONDS", "30"))
    time.sleep(delay)
    logger.info("killing flask app...")
    cleanup()
    os._exit(0)

init_db()

# start the exit timer thread
if os.getenv("PERIODIC_KILL_FLASK"):
    exit_thread = threading.Thread(target=exit_after_delay, daemon=True)
    exit_thread.start()

def _ts():
    return datetime.utcnow().isoformat() + "Z"

# --- short JSON entries (few keys, compact) ---
_SHORT_ENTRIES = [
    (logging.DEBUG,   lambda: {"ts": _ts(), "level": "debug", "msg": "cache hit", "key": "user:1042"}),
    (logging.DEBUG,   lambda: {"ts": _ts(), "level": "debug", "msg": "query ok", "ms": random.randint(1, 5)}),
    (logging.DEBUG,   lambda: {"ts": _ts(), "level": "debug", "msg": "pool stats", "active": random.randint(1, 6), "idle": random.randint(10, 18)}),
    (logging.INFO,    lambda: {"ts": _ts(), "level": "info", "msg": "GET /health", "status": 200}),
    (logging.INFO,    lambda: {"ts": _ts(), "level": "info", "msg": "heartbeat ok"}),
    (logging.INFO,    lambda: {"ts": _ts(), "level": "info", "event": "job_done", "job": "cleanup_sessions", "removed": random.randint(0, 30)}),
    (logging.INFO,    lambda: {"ts": _ts(), "level": "info", "msg": "POST /login", "status": 200, "ms": random.randint(15, 50)}),
    (logging.WARNING, lambda: {"ts": _ts(), "level": "warn", "msg": "disk usage high", "pct": random.randint(80, 94)}),
    (logging.WARNING, lambda: {"ts": _ts(), "level": "warn", "msg": "token expiring", "user_id": 2201}),
    (logging.DEBUG,   lambda: {"ts": _ts(), "level": "debug", "msg": "cache miss", "key": f"session:{random.randint(1000,9999)}"}),
]

# --- medium JSON entries ---
_MEDIUM_ENTRIES = [
    (logging.INFO,    lambda: {
        "timestamp": _ts(), "level": "info", "service": "flask-api", "event": "request",
        "method": "GET", "path": "/api/v1/orders", "status": 200, "duration_ms": random.randint(20, 120),
        "request_id": f"req-{random.randint(10000,99999)}", "user_id": random.randint(1000, 9999),
        "response_size": random.randint(200, 5000),
    }),
    (logging.INFO,    lambda: {
        "timestamp": _ts(), "level": "info", "service": "flask-api", "event": "order_created",
        "order_id": random.randint(8000, 9999), "user_id": random.randint(1000, 5000),
        "total": round(random.uniform(10.0, 500.0), 2), "currency": "USD",
        "items": random.randint(1, 12), "payment_method": random.choice(["card", "paypal", "apple_pay"]),
    }),
    (logging.WARNING, lambda: {
        "timestamp": _ts(), "level": "warning", "service": "flask-api", "event": "slow_query",
        "duration_ms": random.randint(800, 3000),
        "query": "SELECT o.*, u.email FROM orders o JOIN users u ON u.id = o.user_id WHERE o.status = 'pending' ORDER BY o.created_at DESC LIMIT 100",
        "db": "flaskdb", "connection_id": random.randint(1, 20),
    }),
    (logging.WARNING, lambda: {
        "timestamp": _ts(), "level": "warning", "service": "flask-api", "event": "rate_limit",
        "client_ip": f"{random.randint(100,200)}.{random.randint(0,255)}.{random.randint(0,255)}.{random.randint(1,254)}",
        "requests_in_window": random.randint(85, 99), "limit": 100, "window_seconds": 60,
        "endpoint": random.choice(["/api/v1/search", "/api/v1/users", "/api/v1/export"]),
    }),
    (logging.INFO,    lambda: {
        "timestamp": _ts(), "level": "info", "service": "flask-api", "event": "email_sent",
        "to": f"user{random.randint(1000,9999)}@example.com", "template": "order_confirmation",
        "smtp_ms": random.randint(80, 400), "queue_depth": random.randint(0, 15),
    }),
    (logging.DEBUG,   lambda: {
        "timestamp": _ts(), "level": "debug", "service": "flask-api", "event": "auth_check",
        "user_id": random.randint(1000, 9999), "roles": random.choice([["admin"], ["user"], ["user", "editor"]]),
        "token_exp": random.randint(60, 3600), "ip": f"10.0.{random.randint(0,3)}.{random.randint(1,254)}",
    }),
]

# --- long JSON entries (many keys, multiline errors, stack traces) ---
_LONG_ENTRIES = [
    (logging.ERROR,   lambda: {
        "timestamp": _ts(), "level": "error", "service": "flask-api", "event": "unhandled_exception",
        "request_id": f"req-{random.randint(10000,99999)}",
        "method": "POST", "path": "/api/v1/checkout", "status": 500,
        "user_id": random.randint(1000, 9999),
        "error": {
            "type": "ConnectionError",
            "message": "could not connect to payment gateway after 3 retries:\n  attempt 1: connection timed out after 5000ms (host=pay.stripe.internal:443)\n  attempt 2: connection timed out after 5000ms (host=pay.stripe.internal:443)\n  attempt 3: TLS handshake failed: certificate verify failed\n\nthe payment service may be experiencing an outage.\nplease check https://status.stripe.internal for current status.",
            "retries": 3,
            "last_error_code": "ECONNREFUSED",
        },
        "request_body": {"order_id": 9001, "amount": 142.50, "currency": "USD"},
        "duration_ms": 15230,
        "trace_id": f"trace-{random.randint(100000,999999)}",
        "span_id": f"span-{random.randint(1000,9999)}",
    }),
    (logging.ERROR,   lambda: {
        "timestamp": _ts(), "level": "error", "service": "flask-api", "event": "db_error",
        "error": {
            "type": "psycopg2.errors.DeadlockDetected",
            "message": "deadlock detected\nDetail: Process 1842 waits for ShareLock on transaction 5849321;\n  blocked by process 1847.\nProcess 1847 waits for ShareLock on transaction 5849319;\n  blocked by process 1842.\nHint: See server log for query details.\nContext: while updating tuple (42,7) in relation \"orders\"",
        },
        "query": "UPDATE orders SET status = 'shipped', updated_at = NOW() WHERE id = $1 AND status = 'paid'",
        "params": [random.randint(8000, 9999)],
        "connection_id": random.randint(1, 20),
        "transaction_id": f"tx-{random.randint(100000,999999)}",
        "duration_ms": random.randint(5000, 30000),
        "db": "flaskdb",
        "retryable": True,
    }),
    (logging.ERROR,   lambda: {
        "timestamp": _ts(), "level": "error", "service": "flask-api", "event": "request_failed",
        "request_id": f"req-{random.randint(10000,99999)}",
        "method": "POST", "path": "/api/v1/users/import",
        "status": 500, "duration_ms": random.randint(200, 2000),
        "user_id": random.randint(1000, 9999),
        "error": {
            "type": "ValidationError",
            "message": "bulk import failed: 3 records invalid out of 150 total",
            "details": [
                {"row": 42, "field": "email", "error": "invalid format: 'not-an-email'"},
                {"row": 87, "field": "phone", "error": "unsupported country code: +999"},
                {"row": 131, "field": "name", "error": "exceeds maximum length of 255 characters"},
            ],
        },
        "file_name": "users_march_2026.csv",
        "file_size_bytes": 284910,
        "records_processed": 147,
        "records_failed": 3,
        "trace_id": f"trace-{random.randint(100000,999999)}",
    }),
    (logging.ERROR,   lambda: {
        "timestamp": _ts(), "level": "error", "service": "flask-api", "event": "crash_report",
        "error": {
            "type": "RuntimeError",
            "message": "worker thread panicked unexpectedly",
            "stacktrace": (
                "Traceback (most recent call last):\n"
                "  File \"/app/workers/order_processor.py\", line 84, in process_batch\n"
                "    results = pool.map(process_single_order, batch)\n"
                "  File \"/usr/lib/python3.11/multiprocessing/pool.py\", line 367, in map\n"
                "    return self._map_async(func, iterable, mapstar, chunksize).get()\n"
                "  File \"/app/workers/order_processor.py\", line 42, in process_single_order\n"
                "    shipping = calculate_shipping(order['items'], order['destination'])\n"
                "  File \"/app/services/shipping.py\", line 118, in calculate_shipping\n"
                "    rate = get_carrier_rate(carrier, weight_kg, dims)\n"
                "  File \"/app/services/shipping.py\", line 67, in get_carrier_rate\n"
                "    response = requests.post(carrier.api_url, json=payload, timeout=10)\n"
                "  File \"/usr/lib/python3.11/site-packages/requests/api.py\", line 115, in post\n"
                "    return request('POST', url, json=json, **kwargs)\n"
                "requests.exceptions.ConnectionError: HTTPSConnectionPool(host='api.fedex.internal', port=443): "
                "Max retries exceeded with url: /rate/v1/rates/quotes "
                "(Caused by NewConnectionError('<urllib3.connection.HTTPSConnection object>: "
                "Failed to establish a new connection: [Errno -2] Name or service not known'))"
            ),
        },
        "batch_id": f"batch-{random.randint(1000,9999)}",
        "batch_size": random.randint(10, 100),
        "orders_completed": random.randint(0, 9),
        "orders_remaining": random.randint(10, 90),
        "worker_id": f"worker-{random.randint(1,8)}",
        "uptime_seconds": random.randint(100, 86400),
        "memory_mb": random.randint(128, 512),
        "trace_id": f"trace-{random.randint(100000,999999)}",
        "span_id": f"span-{random.randint(1000,9999)}",
    }),
    (logging.WARNING, lambda: {
        "timestamp": _ts(), "level": "warning", "service": "flask-api", "event": "degraded_dependency",
        "dependency": "redis",
        "host": "redis-primary.internal:6379",
        "latency_p99_ms": random.randint(200, 800),
        "latency_p50_ms": random.randint(10, 50),
        "error_rate_pct": round(random.uniform(2.0, 15.0), 1),
        "circuit_breaker": "half-open",
        "fallback": "local_cache",
        "last_error": f"CLUSTERDOWN The cluster is down\nredirected to {random.choice(['10.0.1.12', '10.0.1.13', '10.0.1.14'])}:6379\nbut the node returned READONLY You can't write against a read only replica.",
        "affected_endpoints": ["/api/v1/sessions", "/api/v1/cart", "/api/v1/recommendations"],
        "since": _ts(),
        "alert_id": f"alert-{random.randint(1000,9999)}",
    }),
    (logging.INFO,    lambda: {
        "timestamp": _ts(), "level": "info", "service": "flask-api", "event": "deployment_info",
        "version": f"2.{random.randint(10,30)}.{random.randint(0,9)}",
        "commit": f"{random.randint(0xa000000, 0xfffffff):07x}",
        "environment": "production",
        "region": random.choice(["us-east-1", "eu-west-1", "ap-southeast-1"]),
        "instance_id": f"i-{random.randint(0x10000000, 0xffffffff):08x}",
        "pod": f"flask-api-{random.choice(['a','b','c'])}{random.randint(1,5)}-{random.randint(10000,99999):05d}",
        "node": f"ip-10-0-{random.randint(0,3)}-{random.randint(1,254)}.ec2.internal",
        "cpu_cores": random.choice([2, 4, 8]),
        "memory_limit_mb": random.choice([512, 1024, 2048]),
        "config": {
            "max_connections": 20,
            "worker_threads": random.choice([4, 8, 16]),
            "log_level": "debug",
            "feature_flags": {
                "new_checkout_flow": True,
                "beta_search": False,
                "async_emails": True,
            },
        },
    }),
]

def _pick_log_entry():
    r = random.random()
    if r < 0.40:
        level, fn = random.choice(_SHORT_ENTRIES)
    elif r < 0.75:
        level, fn = random.choice(_MEDIUM_ENTRIES)
    else:
        level, fn = random.choice(_LONG_ENTRIES)
    return level, json.dumps(fn())


def periodic_logger():
    logs_per_second = int(os.getenv("PERIODIC_LOGGING_LOGS_PER_SECOND", "1"))
    delay = 1 / logs_per_second
    while True:
        level, msg = _pick_log_entry()
        logger.log(level, msg)
        time.sleep(delay)

# periodically emit realistic log entries
if os.getenv("PERIODIC_LOGGING"):
    logging_thread = threading.Thread(target=periodic_logger, daemon=True)
    logging_thread.start()

def _make_long_text(num_bytes):
    words = ["processing", "request", "user", "order", "session", "cache",
             "database", "index", "query", "worker", "event", "payload",
             "stream", "record", "batch", "token", "config", "metric"]
    res = "long "
    while len(res) < num_bytes:
        res += random.choice(words) + " "
    return res[:num_bytes]

def periodic_big_logger():
    num_bytes = int(os.getenv("PERIODIC_BIG_LOGGING_BYTES_PER_LOG", "500_000"))
    logs_per_second = float(os.getenv("PERIODIC_BIG_LOGGING_LOGS_PER_SECOND", "0.05"))
    delay = 1 / logs_per_second
    while True:
        logger.info(_make_long_text(num_bytes))
        time.sleep(delay)

# periodically log really long random text
if os.getenv("PERIODIC_BIG_LOGGING"):
    logging_thread = threading.Thread(target=periodic_big_logger, daemon=True)
    logging_thread.start()

@app.route("/health")
def health():
    logger.debug("health check ok")
    return jsonify({"status": "healthy"}), 200

@app.route("/status", methods=["GET"])
def status():
    logger.debug("status endpoint called")
    conn = get_db_connection()
    try:
        with conn.cursor() as cur:
            cur.execute("UPDATE status_hits SET hits = hits + 1 WHERE id = 1 RETURNING hits")
            hits = cur.fetchone()[0]
        conn.commit()
        logger.info("status hits updated successfully")
        return str(hits)
    except Exception as e:
        conn.rollback()
        logger.error("error updating hits", exc_info=True)
        return jsonify({"error": "Database error"}), 500
    finally:
        return_db_connection(conn)

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=int(os.getenv("PORT", "5000")))
