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


class JSONFormatter(logging.Formatter):
    def format(self, record):
        log_record = {
            "timestamp": datetime.utcnow().isoformat(),
            "level": record.levelname,
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName,
            "line": record.lineno
        }
        if record.exc_info:
            log_record["exception"] = self.formatException(record.exc_info)
        return json.dumps(log_record)

logger = logging.getLogger()
logger.setLevel(logging.DEBUG)
handler = logging.StreamHandler()
handler.setFormatter(JSONFormatter())
logger.addHandler(handler)

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

def generate_random_text(num_bytes, prefix=""):
    words = ['hello', 'world', 'test', 'random', 'text', 'words', 'generator',
             'python', 'code', 'sample']
    # choose words randomly, returning exactly num_bytes of text
    res = prefix
    while len(res) < num_bytes:
        res += random.choice(words) + " "
    return res[:num_bytes]


def periodic_logger():
    num_bytes = int(os.getenv("PERIODIC_LOGGING_BYTES_PER_LOG", "1000"))
    logs_per_second = int(os.getenv("PERIODIC_LOGGING_LOGS_PER_SECOND", "1"))
    delay = 1/logs_per_second
    while True:
        logger.info(generate_random_text(num_bytes))
        time.sleep(delay)

# Periodically log random text
if os.getenv("PERIODIC_LOGGING"):
    logging_thread = threading.Thread(target=periodic_logger, daemon=True)
    logging_thread.start()

def periodic_big_logger():
    num_bytes = int(os.getenv("PERIODIC_BIG_LOGGING_BYTES_PER_LOG", "500_000"))
    logs_per_second = int(os.getenv("PERIODIC_BIG_LOGGING_LOGS_PER_SECOND", "0.05"))
    delay = 1/logs_per_second
    while True:
        logger.info(generate_random_text(num_bytes, prefix="long "))
        time.sleep(delay)

# Periodically log really long random text
if os.getenv("PERIODIC_BIG_LOGGING"):
    logging_thread = threading.Thread(target=periodic_big_logger, daemon=True)
    logging_thread.start()

@app.route("/health")
def health():
    logger.info("Health check endpoint called")
    logger.info("FIRST Lorem ipsum \n\tdolor sit amet, consectetur adipiscing elit. \n\tSed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim \n\tad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.")
    logger.info("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod \n\ttempor incididunt ut labore et dolore magna aliqua. \n\tUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.")
    logger.info("LAST Lorem ipsum \n\tdolor sit amet, consectetur adipiscing elit. \n\tSed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad \n\tminim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.")
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