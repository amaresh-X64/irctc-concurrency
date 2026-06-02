import os
import logging
from typing import Optional
from elasticsearch import Elasticsearch, NotFoundError

logger = logging.getLogger(__name__)

ELASTIC_URL = os.getenv("ELASTICSEARCH_URL", "http://elasticsearch:9200")
INDEX_NAME = "trains"

INDEX_MAPPING = {
    "settings": {
        "analysis": {
            "filter": {
                "city_synonyms": {
                    "type": "synonym",
                    "synonyms": [
                        "bengaluru, bengalore, bangalore",
                        "mumbai, bombay",
                        "chennai, madras",
                        "kolkata, calcutta",
                        "delhi, new delhi",
                    ]
                }
            },
            "analyzer": {
                "city_analyzer": {
                    "type": "custom",
                    "tokenizer": "standard",
                    "filter": ["lowercase", "asciifolding", "city_synonyms"]
                }
            }
        }
    },
    "mappings": {
        "properties": {
            "id":               {"type": "integer"},
            "train_number":     {"type": "keyword"},
            "train_name":       {"type": "text", "analyzer": "city_analyzer",
                                 "fields": {"keyword": {"type": "keyword"}}},
            "source":           {"type": "text", "analyzer": "city_analyzer",
                                 "fields": {"keyword": {"type": "keyword"}}},
            "destination":      {"type": "text", "analyzer": "city_analyzer",
                                 "fields": {"keyword": {"type": "keyword"}}},
            "departure_time":   {"type": "keyword"},
            "arrival_time":     {"type": "keyword"},
            "total_seats":      {"type": "integer"},
            "available_seats":  {"type": "integer"},
            "price":            {"type": "float"},
        }
    }
}


def get_es_client() -> Optional[Elasticsearch]:
    try:
        client = Elasticsearch(ELASTIC_URL, request_timeout=3)
        if client.ping():
            return client
        logger.warning("Elasticsearch ping failed — falling back to Postgres")
    except Exception as e:
        logger.warning("Elasticsearch unavailable: %s — falling back to Postgres", e)
    return None


def ensure_index(client: Elasticsearch) -> None:
    if not client.indices.exists(index=INDEX_NAME):
        client.indices.create(index=INDEX_NAME, body=INDEX_MAPPING)
        logger.info("Created Elasticsearch index '%s'", INDEX_NAME)


def index_train(client: Elasticsearch, train: dict) -> None:
    client.index(index=INDEX_NAME, id=train["id"], document=train)


def bulk_index_trains(client: Elasticsearch, trains: list[dict]) -> None:
    if not trains:
        return
    actions = []
    for t in trains:
        actions.append({"index": {"_index": INDEX_NAME, "_id": t["id"]}})
        actions.append(t)
    client.bulk(operations=actions, refresh=True)
    logger.info("Bulk-indexed %d trains into Elasticsearch", len(trains))


def update_available_seats(client: Elasticsearch, train_id: int, available_seats: int) -> None:
    try:
        client.update(
            index=INDEX_NAME,
            id=train_id,
            doc={"available_seats": available_seats},
        )
    except NotFoundError:
        logger.warning("Train %d not found in ES index for seat update", train_id)


def search_trains(
    client: Elasticsearch,
    source: str,
    destination: str,
) -> list[dict]:
    query = {
        "query": {
            "bool": {
                "must": [
                    {
                        "multi_match": {
                            "query": source,
                            "fields": ["source^3", "source.keyword^5"],
                            "fuzziness": "AUTO",
                            "operator": "or",
                        }
                    },
                    {
                        "multi_match": {
                            "query": destination,
                            "fields": ["destination^3", "destination.keyword^5"],
                            "fuzziness": "AUTO",
                            "operator": "or",
                        }
                    },
                ]
            }
        },
        "sort": [{"departure_time": {"order": "asc"}}],
        "size": 50,
    }

    resp = client.search(index=INDEX_NAME, body=query)
    return [hit["_source"] for hit in resp["hits"]["hits"]]


def search_all_trains(client: Elasticsearch) -> list[dict]:
    resp = client.search(
        index=INDEX_NAME,
        body={
            "query": {"match_all": {}},
            "sort": [{"departure_time": {"order": "asc"}}],
            "size": 200,
        },
    )
    return [hit["_source"] for hit in resp["hits"]["hits"]]