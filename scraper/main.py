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
from utils import *
import writer

# Setup Selenium Webdriver
CHROMEDRIVER_PATH = r"./driver/chromedriver.exe"
service = Service(CHROMEDRIVER_PATH)
options = Options()
options.add_argument("--headless=new");

# Disable warning, error and info logs, show only fatal errors
options.add_argument("--log-level=3")
driver = webdriver.Chrome(service=service, options=options)

# Skip problems it is already scraped based on track.conf file
completed_upto = read_tracker("track.conf")

def reset_driver():
    global driver
    driver.quit()
    driver = webdriver.Chrome(service=service, options=options)

# Get list of all unique tags
def get_tag_list():
    url = "https://leetcode.com/problemset/"
    
    try:
        driver.get(url)

        try:
            WebDriverWait(driver, 5).until(
                EC.presence_of_element_located((By.CLASS_NAME, "text-label-1"))
            )
        except Exception as e:
            pass

        # Get current tab page source
        html = driver.page_source
        soup = bs4.BeautifulSoup(html, "html.parser")

        # Extract Tags
        tag_elems = soup.find_all('span', class_="whitespace-nowrap  group-hover:text-blue dark:group-hover:text-dark-blue text-label-1 dark:text-dark-label-1")
        tags = [tag.get_text(strip=True) for tag in tag_elems]

        return tags
    except Exception as e:
        driver.quit()


# Get all tags associated with a problem
def scrape_tags(problem_num, url):
    print(f"Fetching tags of problem {problem_num} at {url} ")

    try:
        driver.get(url)

        # Enable Dynamic Layout & Skip Tutorial (If Applicable - should only be for first problem)
        try:
            WebDriverWait(driver, 10).until(
                EC.element_to_be_clickable((By.CLASS_NAME, "font-medium items-center whitespace-nowrap focus:outline-none inline-flex bg-fill-3 dark:bg-dark-fill-3 hover:bg-fill-2 dark:hover:bg-dark-fill-2 dark:text-dark-label-2 rounded-full px-10 py-[14px] text-xl text-white"))
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

        WebDriverWait(driver, 10).until(
            EC.presence_of_all_elements_located(
                (By.CSS_SELECTOR, "a.no-underline.hover\\:text-current.relative.inline-flex.items-center.justify-center.text-caption.px-2.py-1.gap-1.rounded-full.bg-fill-secondary.text-text-secondary")
            )
        )

        # Get current tab page source
        html = driver.page_source
        soup = bs4.BeautifulSoup(html, "html.parser")

        # Extract Tags
        tag_elems = soup.find_all('a', class_="no-underline hover:text-current relative inline-flex items-center justify-center text-caption px-2 py-1 gap-1 rounded-full bg-fill-secondary text-text-secondary")
        tags = [tag.get_text(strip=True) for tag in tag_elems]
        
        return tags, None

    except Exception as e:
        driver.quit()
        return None, e


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
        failed_attempts = 0
        for i in range(completed_upto + 1, len(links)):
            question_title_slug, difficulty, problem_num, problem_name, description = links[i]
            url = ALGORITHMS_BASE_URL + question_title_slug
            
            # Scrape Tags
            tags, error = scrape_tags(problem_num, url)
            if error is None:
                print(f"Scraped problem {problem_num} '{problem_name}' with {difficulty} difficulty and tags {tags} at {url}\n")
                writer.addProblem(i, problem_num, problem_name, url, difficulty, tags)
                failed_attempts = 0
            else:
                # If Error is found, reset driver and wait 5 min, then try to scrape it again (close if failed 5 times in a row)
                print(f"Failure to scrape problem: {error}")

                failed_attempts += 1
                if failed_attempts > 5:
                    print(f"Too many failed attempts, closing program")
                    exit(1)

                i -= 1
                reset_driver()
                print(f"Sleeping 5 min...\n")
                time.sleep(300)

            # Sleep for 5 secs between each problem, or 2 minutes every 30 problems (Resets driver when waits 2 min)
            if i % 30 == 0 and i != 0:
                print(f"Sleeping 2 min...\n")
                time.sleep(120)
                reset_driver()
            else:
                print(f"Sleeping 5 sec...\n")
                time.sleep(5)

    finally:
        driver.quit()


if __name__ == "__main__":
    main()