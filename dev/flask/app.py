import json
import logging
import os
import random
import threading
import time
from datetime import datetime

import psycopg2
from flask import Flask, jsonify
from flask_cors import CORS
from psycopg2 import pool


class MixedFormatter(logging.Formatter):
    _plain = logging.Formatter("%(asctime)s %(levelname)-8s %(message)s", datefmt="%Y-%m-%dT%H:%M:%S")

    def format(self, record):
        msg = record.getMessage()
        try:
            json.loads(msg)
            return msg
        except (ValueError, TypeError):
            return self._plain.format(record)

handler = logging.StreamHandler()
handler.setFormatter(MixedFormatter())
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
        logger.info("Database initialized successfully")
    except Exception as e:
        logger.error("Error initializing database", exc_info=True)
        conn.rollback()
    finally:
        return_db_connection(conn)

def cleanup():
    logger.info("Cleaning up resources...")
    db_pool.closeall()

def exit_after_delay():
    delay = int(os.getenv("PERIODIC_KILL_FLASK_SECONDS", "30"))
    time.sleep(delay)
    logger.info("Killing flask app...")
    cleanup()
    os._exit(0)

init_db()

# Start the exit timer thread
if os.getenv("PERIODIC_KILL_FLASK"):
    exit_thread = threading.Thread(target=exit_after_delay, daemon=True)
    exit_thread.start()

_PLAIN_ENTRIES = [
    (logging.DEBUG,   "cache lookup: key=user:1042 hit=True ttl=298s"),
    (logging.DEBUG,   "db query took 3ms: SELECT * FROM users WHERE id=$1"),
    (logging.DEBUG,   "db query took 18ms: SELECT count(*) FROM orders WHERE status='pending'"),
    (logging.DEBUG,   "cache miss: key=session:abc123, fetching from db"),
    (logging.DEBUG,   "worker thread pool: 4/16 threads active"),
    (logging.DEBUG,   "config loaded: POSTGRES_HOST=postgres-service PORT=5000"),
    (logging.INFO,    "GET /api/v1/users 200 14ms"),
    (logging.INFO,    "POST /api/v1/orders 201 22ms"),
    (logging.INFO,    "GET /api/v1/products?category=electronics 200 31ms"),
    (logging.INFO,    "DELETE /api/v1/sessions/xyz 204 5ms"),
    (logging.INFO,    "user 1042 logged in from 203.0.113.47"),
    (logging.INFO,    "order #8821 created for user 1042, total=$142.50"),
    (logging.INFO,    "email queued: welcome@example.com -> user:5510"),
    (logging.INFO,    "scheduled job 'cleanup_expired_sessions' completed, removed 14 rows"),
    (logging.INFO,    "db pool: 3/20 connections in use"),
    (logging.WARNING, "slow query (1.2s): SELECT orders JOIN users ON users.id=orders.user_id"),
    (logging.WARNING, "rate limit approaching for ip 198.51.100.22: 87/100 req/min"),
    (logging.WARNING, "jwt token expiring soon for user 2201, issuing refresh"),
    (logging.WARNING, "disk usage at 81% on /var/data, threshold is 85%"),
    (logging.WARNING, "retrying db connection (attempt 2/3): connection refused"),
    (logging.ERROR,   "failed to send email to user:3309: SMTP timeout after 10s"),
    (logging.ERROR,   "unhandled exception in worker: ZeroDivisionError: division by zero"),
    (logging.ERROR,   "payment gateway returned 502, order #9001 not processed"),
    (logging.ERROR,   "db write failed: deadlock detected, rolling back transaction"),
]

_JSON_ENTRIES = [
    (logging.DEBUG,  {"event": "cache_get",  "key": "user:2048",     "hit": True,  "ttl": 120}),
    (logging.DEBUG,  {"event": "query",      "sql": "SELECT * FROM sessions WHERE token=$1", "ms": 2}),
    (logging.INFO,   {"event": "request",    "method": "GET",  "path": "/health",          "status": 200, "ms": 4}),
    (logging.INFO,   {"event": "request",    "method": "POST", "path": "/api/v1/login",    "status": 200, "ms": 37, "user_id": 1042}),
    (logging.INFO,   {"event": "request",    "method": "GET",  "path": "/api/v1/orders",   "status": 200, "ms": 55, "count": 23}),
    (logging.INFO,   {"event": "job_done",   "job": "send_digest_emails", "sent": 312, "failed": 1}),
    (logging.INFO,   {"event": "db_pool",    "active": 5, "idle": 15, "max": 20}),
    (logging.WARNING,{"event": "slow_query", "ms": 980, "sql": "UPDATE orders SET status=$1 WHERE user_id=$2"}),
    (logging.WARNING,{"event": "rate_limit", "ip": "203.0.113.99", "requests": 95, "limit": 100}),
    (logging.ERROR,  {"event": "exception",  "type": "ConnectionError", "msg": "db connection lost", "attempt": 1}),
    (logging.ERROR,  {"event": "request",    "method": "POST", "path": "/api/v1/checkout",  "status": 500, "ms": 102, "error": "payment timeout"}),
]

def _pick_log_entry():
    if random.random() < 0.35:
        level, data = random.choice(_JSON_ENTRIES)
        return level, json.dumps(data)
    else:
        return random.choice(_PLAIN_ENTRIES)


def periodic_logger():
    logs_per_second = int(os.getenv("PERIODIC_LOGGING_LOGS_PER_SECOND", "1"))
    delay = 1 / logs_per_second
    while True:
        level, msg = _pick_log_entry()
        logger.log(level, msg)
        time.sleep(delay)

# Periodically emit realistic log entries
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
    logs_per_second = int(os.getenv("PERIODIC_BIG_LOGGING_LOGS_PER_SECOND", "0.05"))
    delay = 1 / logs_per_second
    while True:
        logger.info(_make_long_text(num_bytes))
        time.sleep(delay)

# Periodically log really long random text
if os.getenv("PERIODIC_BIG_LOGGING"):
    logging_thread = threading.Thread(target=periodic_big_logger, daemon=True)
    logging_thread.start()

@app.route("/health")
def health():
    logger.debug("health check: db pool active=%d idle=%d", 3, 17)
    logger.info("GET /health 200 2ms")
    return jsonify({"status": "healthy"}), 200

@app.route("/status", methods=["GET"])
def status():
    logger.debug("Status endpoint called")
    conn = get_db_connection()
    try:
        with conn.cursor() as cur:
            cur.execute("UPDATE status_hits SET hits = hits + 1 WHERE id = 1 RETURNING hits")
            hits = cur.fetchone()[0]
        conn.commit()
        logger.info("Status hits updated successfully", extra={"hits": hits})
        return str(hits)
    except Exception as e:
        conn.rollback()
        logger.error("Error updating hits", exc_info=True)
        return jsonify({"error": "Database error"}), 500
    finally:
        return_db_connection(conn)

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=int(os.getenv("PORT", "5000")))