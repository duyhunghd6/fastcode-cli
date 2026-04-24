#!/usr/bin/env python3
import sys
import argparse
from scrapling.fetchers import Fetcher
import html2text
import logging

# Configure basic logging
logging.basicConfig(level=logging.INFO, format='%(levelname)s: %(message)s')

def download_and_convert(url, output_file):
    try:
        logging.info(f"Fetching URL: {url}")
        # Fetch the page using Scrapling
        page = Fetcher.get(url)
        
        # We need the HTML content. Scrapling parse tree can return inner HTML of the document
        html_content = page.xpath('//html').get()
        
        if not html_content:
            logging.error("Could not extract HTML content from the page.")
            sys.exit(1)

        logging.info("Converting HTML to Markdown...")
        # Configure html2text
        h = html2text.HTML2Text()
        h.ignore_links = False
        h.ignore_images = False
        h.body_width = 0 # Don't wrap lines unnecessarily
        
        markdown_text = h.handle(html_content)
        
        logging.info(f"Saving to: {output_file}")
        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(markdown_text)
            
        logging.info("Successfully converted and saved.")
        
    except Exception as e:
        logging.error(f"An error occurred: {e}")
        sys.exit(1)

def main():
    parser = argparse.ArgumentParser(description="Download a URL and convert its content to Markdown.")
    parser.add_argument("url", help="The URL to download")
    parser.add_argument("output_file", help="The path to save the generated Markdown file")
    
    args = parser.parse_args()
    
    download_and_convert(args.url, args.output_file)

if __name__ == '__main__':
    main()
