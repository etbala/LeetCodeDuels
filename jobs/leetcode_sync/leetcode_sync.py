import os
import requests
import json
import psycopg2
from psycopg2.extras import execute_values

def get_leetcode_problems():
    """Fetches all LeetCode problems from the GraphQL API."""
    url = "https://leetcode.com/graphql/"
    query = """
    query allQuestions {
        allQuestions {
            questionId
            questionFrontendId
            title
            titleSlug
            difficulty
            isPaidOnly
            topicTags {
                name
            }
        }
    }
    """
    try:
        response = requests.post(url, json={'query': query}, timeout=15)
        response.raise_for_status()
    except requests.exceptions.RequestException as e:
        print(f"Error fetching data from LeetCode API: {e}")
        return None

    data = response.json()
    problems = data.get('data', {}).get('allQuestions', [])
    
    # Process the raw data into a more usable format
    processed_problems = []
    for problem in problems:
        processed_problems.append({
            'id': int(problem.get('questionFrontendId')),
            'name': problem.get('title'),
            'slug': problem.get('titleSlug'),
            'difficulty': problem.get('difficulty'),
            'is_paid': problem.get('isPaidOnly', False),
            'tags': [tag['name'] for tag in problem.get('topicTags', [])]
        })
    return processed_problems

def get_db_connection():
    """Establishes a connection to the PostgreSQL database."""
    try:
        conn = psycopg2.connect(os.environ.get('DB_URL'))
        return conn
    except Exception as e:
        print(f"Error connecting to the database: {e}")
        return None

def setup_database_schema(conn):
    """Creates the necessary tables/enum if they don't exist."""
    # Transaction 1: Create ENUM
    try:
        with conn.cursor() as cur:
            cur.execute("CREATE TYPE problem_difficulty AS ENUM ('Easy', 'Medium', 'Hard');")
            print("Created 'problem_difficulty' ENUM type.")
    except psycopg2.errors.DuplicateObject:
        # Type already exists
        conn.rollback()
    except Exception as e:
        print(f"Error creating ENUM: {e}")
        conn.rollback()
        raise e # Re-raise the unexpected error to stop the script

    # Transaction 2: Create Tables (this will now run in a fresh transaction)
    try:
        with conn.cursor() as cur:
            cur.execute("""
                CREATE TABLE IF NOT EXISTS problems (
                    id INTEGER PRIMARY KEY,
                    name TEXT NOT NULL,
                    slug TEXT NOT NULL UNIQUE,
                    difficulty problem_difficulty NOT NULL,
                    is_paid BOOLEAN NOT NULL
                );
            """)
            cur.execute("""
                CREATE TABLE IF NOT EXISTS tags (
                    id SERIAL PRIMARY KEY,
                    name TEXT NOT NULL UNIQUE
                );
            """)
            cur.execute("""
                CREATE TABLE IF NOT EXISTS problem_tags (
                    problem_id INTEGER REFERENCES problems(id) ON DELETE CASCADE,
                    tag_id INT REFERENCES tags(id) ON DELETE CASCADE,
                    PRIMARY KEY (problem_id, tag_id)
                );
            """)
            print("Database schema verified.")
    except Exception as e:
        print(f"Error creating tables: {e}")
        conn.rollback()
        raise e

def sync_problems_to_db():
    """Main function to sync LeetCode problems with the relational database."""
    print("Starting LeetCode problem sync...")
    conn = get_db_connection()
    if not conn:
        return

    try:
        setup_database_schema(conn)

        with conn.cursor() as cur:
            cur.execute("SELECT slug, id FROM problems")
            problem_map = dict(cur.fetchall())
            
            cur.execute("SELECT name, id FROM tags")
            tag_map = dict(cur.fetchall())

            print(f"Found {len(problem_map)} existing problems and {len(tag_map)} existing tags.")

            all_problems_from_api = get_leetcode_problems()
            if not all_problems_from_api:
                print("Could not fetch problems from LeetCode. Aborting sync.")
                return

            new_problems = [p for p in all_problems_from_api if p['slug'] not in problem_map]
            
            all_tag_names = {tag for p in all_problems_from_api for tag in p['tags']}
            new_tags = [name for name in all_tag_names if name not in tag_map]

            if new_tags:
                print(f"Found {len(new_tags)} new tags to insert.")
                insert_query = "INSERT INTO tags (name) VALUES %s RETURNING id, name;"
                inserted_tags = execute_values(cur, insert_query, [(t,) for t in new_tags], fetch=True)
                tag_map.update(dict(reversed(t) for t in inserted_tags)) # (name, id)
                print(f"Successfully inserted {len(inserted_tags)} new tags.")

            if new_problems:
                print(f"Found {len(new_problems)} new problems to insert.")
                insert_query = """
                    INSERT INTO problems (id, name, slug, difficulty, is_paid) 
                    VALUES %s RETURNING id, slug;
                """
                data_to_insert = [(p['id'], p['name'], p['slug'], p['difficulty'], p['is_paid']) for p in new_problems]
                inserted_problems = execute_values(cur, insert_query, data_to_insert, fetch=True)
                problem_map.update(dict(reversed(p) for p in inserted_problems)) # (slug, id)
                print(f"Successfully inserted {len(inserted_problems)} new problems.")
            
            cur.execute("SELECT problem_id, tag_id FROM problem_tags")
            existing_links = set(cur.fetchall())
            
            links_to_create = []
            for problem in all_problems_from_api:
                problem_id = problem_map.get(problem['slug'])
                for tag_name in problem['tags']:
                    tag_id = tag_map.get(tag_name)
                    if problem_id and tag_id and (problem_id, tag_id) not in existing_links:
                        links_to_create.append((problem_id, tag_id))

            if links_to_create:
                print(f"Creating {len(links_to_create)} new problem-tag links.")
                insert_query = "INSERT INTO problem_tags (problem_id, tag_id) VALUES %s ON CONFLICT DO NOTHING;"
                execute_values(cur, insert_query, links_to_create)
                print(f"Successfully created links.")
            else:
                print("No new problem-tag links to create.")

            conn.commit()

    except Exception as e:
        print(f"An error occurred: {e}")
        conn.rollback()
    finally:
        conn.close()
        print("Sync complete. Database connection closed.")

if __name__ == "__main__":
    os.environ['DB_URL'] = 'postgresql://neondb_owner:npg_yQ1tjhOurba3@ep-divine-breeze-adnuajh4-pooler.c-2.us-east-1.aws.neon.tech/neondb?sslmode=require'
    sync_problems_to_db()