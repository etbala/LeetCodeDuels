import json
from utils import *

"""
    Store Information as JSON (Temporary - Make Database Later)
    Store Problem Number, Name, URL, Difficulty, Tags
"""

"""
    When stored in database, have 3 tables with the following schema:

    CREATE TABLE problems (
        problem_id SERIAL PRIMARY KEY,
        name VARCHAR(127),
        url VARCHAR(127),
        difficulty VARCHAR(10)
    );

    CREATE TABLE tags (
        tag_id SERIAL PRIMARY KEY,
        tag_name VARCHAR(50)
    );

    CREATE TABLE problem_tags (
        problem_id INT REFERENCES problems(problem_id),
        tag_id INT REFERENCES tags(tag_id),
        PRIMARY KEY (problem_id, tag_id)
    );
"""

def addProblem(id, problem_num, problem_name, problem_link, problem_difficulty, problem_tags):
    # File path
    file_path = 'lc_problems.json'

    # Read existing data
    with open(file_path, 'r') as file:
        try:
            problems = json.load(file)
        except json.JSONDecodeError:  # In case the file is empty
            problems = {}

    # Add or update the problem
    problems[id] = {
        'name': problem_name,
        'num': problem_num,
        'url': problem_link,
        'difficulty': problem_difficulty,
        'tags': problem_tags
    }

    # Write data back to file
    with open(file_path, 'w') as file:
        json.dump(problems, file, indent=4)

    update_tracker('track.conf', id)

    return True

def updateProblem():
    return

def main():
    # Print out lc_problems.json
    file_path = 'lc_problems.json'

    try:
        with open(file_path, 'r') as file:
            data = json.load(file)
            print(json.dumps(data, indent=4))
    except FileNotFoundError:
        print(f"The file {file_path} was not found.")
    except json.JSONDecodeError:
        print(f"The file {file_path} does not contain valid JSON.")
    except Exception as e:
        print(f"An error occurred: {e}")


if __name__ == "__main__":
    main()