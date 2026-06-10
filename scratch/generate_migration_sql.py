import pg8000
import json
import os
from datetime import datetime

db_url = "postgres://postgres:0goQtfw82yGMsYCf3InCIeSg1xDZq7jkXRhadiqpxzD9DbRVX0qdATUsmbbU1302@192.168.1.20:5454/postgres?sslmode=disable"
output_dir = "C:/Users/Baptiste/.gemini/antigravity/brain/bbb76eb3-3535-40b6-b770-e585a2635877/scratch/migration_chunks"

os.makedirs(output_dir, exist_ok=True)

def parse_url(url):
    parts = url.split("://")[1].split("@")
    user_pass = parts[0].split(":")
    host_port_db = parts[1].split("/")
    host_port = host_port_db[0].split(":")
    db_name = host_port_db[1].split("?")[0]
    return {
        "user": user_pass[0],
        "password": user_pass[1],
        "host": host_port[0],
        "port": int(host_port[1]),
        "database": db_name
    }

def sql_value(val):
    if val is None:
        return "NULL"
    if isinstance(val, bool):
        return "TRUE" if val else "FALSE"
    if isinstance(val, (int, float)):
        return str(val)
    if isinstance(val, datetime):
        return f"'{val.isoformat()}'"
    if isinstance(val, (dict, list)):
        escaped = json.dumps(val).replace("'", "''")
        return f"'{escaped}'::jsonb"
    # String types
    escaped = str(val).replace("'", "''")
    return f"'{escaped}'"

def dump_table(cursor, table_name):
    print(f"Dumping table {table_name}...")
    cursor.execute(f"SELECT * FROM {table_name};")
    rows = cursor.fetchall()
    
    # Get column names
    col_names = [desc[0] for desc in cursor.description]
    
    # Special adjustments for Supabase compatibility
    has_uuid = False
    if table_name == 'users':
        if 'uuid' not in col_names:
            col_names.append('uuid')
            has_uuid = True
            
    if not rows:
        print(f"Table {table_name} is empty.")
        return
        
    chunk_size = 500
    file_idx = 1
    
    for i in range(0, len(rows), chunk_size):
        chunk = rows[i:i+chunk_size]
        sql_lines = []
        
        # Disable trigger/RLS if needed, but since we run as superuser we don't strictly have to.
        # However, to be fast and correct, we do normal inserts.
        for row in chunk:
            vals = []
            for idx, col in enumerate(cursor.description):
                val = row[idx]
                vals.append(sql_value(val))
            if table_name == 'users' and has_uuid:
                vals.append("NULL") # uuid is NULL for legacy users
                
            cols_str = ", ".join(col_names)
            vals_str = ", ".join(vals)
            sql_lines.append(f"INSERT INTO public.{table_name} ({cols_str}) VALUES ({vals_str});")
            
        filename = f"{output_dir}/{table_name}_chunk_{file_idx}.sql"
        with open(filename, "w", encoding="utf-8") as f:
            f.write("\n".join(sql_lines))
        print(f" - Wrote {len(sql_lines)} rows to {filename}")
        file_idx += 1

def main():
    creds = parse_url(db_url)
    conn = pg8000.connect(**creds)
    cursor = conn.cursor()
    
    # Ordered tables to prevent foreign key errors during truncation/insertion
    # Dependents first for deletion, dependencies first for insertion
    tables = [
        'users',
        'games',
        'game_players',
        'game_moves',
        'messages',
        'game_message_reads',
        'reports',
        'push_subscriptions',
        'daily_puzzles',
        'puzzle_attempts',
        'dictionary_definitions',
        'user_achievements',
        'user_friends',
        'user_rating_history'
    ]
    
    # Dump each table
    for t in tables:
        dump_table(cursor, t)
        
    cursor.close()
    conn.close()
    print("Done generating migration SQL!")

if __name__ == "__main__":
    main()
