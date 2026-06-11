import os
from elasticsearch import Elasticsearch

TRACES_INDEX = "irctc-traces"
PERCOLATOR_INDEX = "irctc-anomaly-rules"

def get_es() -> Elasticsearch:
    url = os.getenv("ELASTICSEARCH_URL", "http://elasticsearch:9200")
    return Elasticsearch(url)