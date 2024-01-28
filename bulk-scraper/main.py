# Author: Ethan Balakumar, based on code by Bishal Sarang
import json
import time
import bs4
import requests
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.by import By
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.chrome.service import Service

import writer
from utils import *

# Setup Selenium Webdriver
CHROMEDRIVER_PATH = r"./driver/chromedriver.exe"
service = Service(CHROMEDRIVER_PATH)
options = Options()
options.add_argument("--headless=new");

# Disable warning, error and info logs, show only fatal errors
options.add_argument("--log-level=3")
driver = webdriver.Chrome(service=service, options=options)

# Get upto which problem it is already scraped from track.conf file
completed_upto = read_tracker("track.conf")

def scrape_tags(problem_num, url):
    print(f"Fetching problem num {problem_num} at {url} ")

    try:
        driver.get(url)

        # Enable Dynamic Layout & Skip Tutorial (If Applicable - should only be for first problem)
        try:
            WebDriverWait(driver, 5).until(
                EC.element_to_be_clickable((By.XPATH, "//button[contains(text(), 'Enable Dynamic Layout')]"))
            ).click()

            WebDriverWait(driver, 5).until(
                EC.presence_of_element_located((By.CLASS_NAME, "bg-sd-popover"))
            )

            for _ in range(6):
                driver.find_element(By.TAG_NAME, 'body').send_keys(Keys.ESCAPE)
                time.sleep(0.1)

            # print(" Skipped Dynamic Layout Tutorial")
        except Exception as e:
            pass
            # print(" No Dynamic Layout Tutorial Detected")

        # Get current tab page source
        html = driver.page_source
        soup = bs4.BeautifulSoup(html, "html.parser")

        # Extract Tags
        tag_elems = soup.find_all('a', class_="no-underline hover:text-current relative inline-flex items-center justify-center text-caption px-2 py-1 gap-1 rounded-full bg-fill-secondary text-text-secondary")
        tags = [tag.get_text(strip=True) for tag in tag_elems]
        
        # Update upto which the problem is downloaded
        update_tracker('track.conf', problem_num)
        print(f"Writing problem num {problem_num} with url {url} successfull")

        return tags

    except Exception as e:
        print(f"Failed Writing!!  {e} ")
        driver.quit()


def main():

    # Leetcode API URL to get json of problems on algorithms categories
    ALGORITHMS_ENDPOINT_URL = "https://leetcode.com/api/problems/algorithms/"

    # Problem URL is of format ALGORITHMS_BASE_URL + question__title_slug
    # If question__title_slug = "two-sum" then URL is https://leetcode.com/problems/two-sum
    ALGORITHMS_BASE_URL = "https://leetcode.com/problems/"

    # Load JSON from API
    algorithms_problems_json = requests.get(ALGORITHMS_ENDPOINT_URL).content
    algorithms_problems_json = json.loads(algorithms_problems_json)

    links = []
    for child in algorithms_problems_json["stat_status_pairs"]:
            # Only process free problems
            if not child["paid_only"]:
                question_title_slug = child["stat"]["question__title_slug"]
                question_article_slug = child["stat"]["question__article__slug"]
                question_title = child["stat"]["question__title"]
                frontend_question_id = child["stat"]["frontend_question_id"]
                difficulty_level = child["difficulty"]["level"]
                
                difficulty_translator = {
                    1: "Easy",
                    2: "Medium",
                    3: "Hard"
                }
                difficulty = difficulty_translator.get(difficulty_level)

                links.append((question_title_slug, difficulty, frontend_question_id, question_title, question_article_slug))

    # Sort by problem id in ascending order
    links = sorted(links, key=lambda x: (x[2]))

    try: 
        for i in range(completed_upto + 1, len(links)):
            question_title_slug, difficulty, problem_num, problem_name, description = links[i]
            
            url = ALGORITHMS_BASE_URL + question_title_slug
            
            # Scrape Tags
            tags = scrape_tags(problem_num, url)
            print(f"Scraped Question {problem_num}. {problem_name} with {difficulty} difficulty with tags {tags} at {url}\n")

            writer.addProblem(problem_num, problem_name, url, difficulty, tags)

            # Sleep for 4 secs between each problem
            time.sleep(4)

    finally:
        driver.quit()


if __name__ == "__main__":
    main()