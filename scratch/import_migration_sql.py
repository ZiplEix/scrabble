import os
import subprocess
import sys

# Directory containing migration chunks
chunks_dir = "C:/Users/Baptiste/.gemini/antigravity/brain/bbb76eb3-3535-40b6-b770-e585a2635877/scratch/migration_chunks"
# Supabase project directory for config
workdir = "c:/Users/Baptiste/Documents/project/scrabble/frontend"

# SQL files in order of dependencies (dependencies first)
sql_files = [
    "users_chunk_1.sql",
    "games_chunk_1.sql",
    "game_players_chunk_1.sql",
    "game_moves_chunk_1.sql",
    "game_moves_chunk_2.sql",
    "game_moves_chunk_3.sql",
    "game_moves_chunk_4.sql",
    "game_moves_chunk_5.sql",
    "game_moves_chunk_6.sql",
    "game_moves_chunk_7.sql",
    "game_moves_chunk_8.sql",
    "messages_chunk_1.sql",
    "messages_chunk_2.sql",
    "game_message_reads_chunk_1.sql",
    "reports_chunk_1.sql",
    "push_subscriptions_chunk_1.sql",
    "daily_puzzles_chunk_1.sql",
    "puzzle_attempts_chunk_1.sql",
    "dictionary_definitions_chunk_1.sql",
    "user_achievements_chunk_1.sql",
    "user_friends_chunk_1.sql",
    "user_rating_history_chunk_1.sql"
]

def run_sql_file(filename):
    filepath = os.path.join(chunks_dir, filename)
    if not os.path.exists(filepath):
        print(f"Warning: File {filepath} does not exist. Skipping.")
        return True

    print(f"Importing {filename}...")
    # Run: bunx supabase db query --linked --workdir <workdir> -f <filepath>
    # Using shell=True for windows path and bunx resolver compatibility
    cmd = f'bunx supabase db query --linked --workdir "{workdir}" -f "{filepath}"'
    
    res = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    if res.returncode != 0:
        print(f"Error importing {filename}:")
        print(res.stderr)
        return False
    else:
        print(f"Successfully imported {filename}.")
        return True

def main():
    print("Starting SQL chunks import...")
    for f in sql_files:
        success = run_sql_file(f)
        if not success:
            print("Import failed. Aborting further imports.")
            sys.exit(1)
    print("All tables successfully imported!")

if __name__ == "__main__":
    main()
