from flask import Flask, jsonify
import psycopg2
from psycopg2 import pool
from flask_cors import CORS
import os
import logging
import json
from datetime import datetime
import random
import signal
import threading
import time

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
    time.sleep(30)  # Wait for 60 seconds
    logger.info("Exiting after 1 minute delay")
    cleanup()
    os._exit(0)

init_db()

# Start the exit timer thread
# exit_thread = threading.Thread(target=exit_after_delay, daemon=True)
# exit_thread.start()

def log_a_lot():
    while True:
        logger.info("Periodic log message")
        time.sleep(0.05)

# Start the thread that creates a lot of log messages
logging_thread = threading.Thread(target=log_a_lot, daemon=True)
logging_thread.start()

@app.route("/health")
def health():
    logger.info("Health check endpoint called")
    logger.info("FIRST Lorem ipsum \n\tdolor sit amet, consectetur adipiscing elit. \n\tSed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim \n\tad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.")
    if random.random() < 0.05:
        logger.info(generate_random_text(100_000))
    logger.info("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod \n\ttempor incididunt ut labore et dolore magna aliqua. \n\tUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.")
    logger.info("LAST Lorem ipsum \n\tdolor sit amet, consectetur adipiscing elit. \n\tSed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad \n\tminim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.")
    return jsonify({"status": "healthy"}), 200

def generate_random_text(num_words):
    words = ['hello', 'world', 'test', 'random', 'text', 'words', 'generator',
             'python', 'code', 'sample']
    return ' '.join(random.choices(words, k=num_words))

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