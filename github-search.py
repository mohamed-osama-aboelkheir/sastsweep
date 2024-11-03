import time
import argparse
from github import Github

def code_search(auth, query):
    urls = []
    try:
        results = auth.search_code(query)
        for result in results:
            time.sleep(1)  # Rate limiting
            url = result.repository.html_url
            if url in urls:
                continue

            urls.append(url)
            print(url, flush=True)
    except Exception as e:
        print(f"Error during search: {e}", flush=True)
        return

def main():
    parser = argparse.ArgumentParser(description='Search GitHub code with authentication')
    parser.add_argument('--token', '-t', required=True, help='GitHub authentication token')
    parser.add_argument('--query', '-q', required=True, help='Search query')
    
    args = parser.parse_args()
    
    try:
        auth = Github(args.token)
        code_search(auth, args.query)
    except Exception as e:
        print(f"Error initializing GitHub client: {e}")
        return

if __name__ == "__main__":
    main()
