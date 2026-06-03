import os
import pytest
from unittest.mock import MagicMock, patch
from elasticsearch import NotFoundError

from app.search.elastic import (
    get_es_client,
    ensure_index,
    index_train,
    bulk_index_trains,
    update_available_seats,
    search_trains,
    search_all_trains,
    INDEX_NAME,
    INDEX_MAPPING,
)


def make_client():
    return MagicMock()


SAMPLE_TRAIN = {
    "id": 1,
    "train_number": "12345",
    "train_name": "Chennai Express",
    "source": "Chennai",
    "destination": "Mumbai",
    "departure_time": "06:00:00",
    "arrival_time": "22:00:00",
    "total_seats": 100,
    "available_seats": 80,
    "price": 1500.0,
}



def test_get_es_client_returns_client_when_ping_succeeds():
    with patch("app.search.elastic.Elasticsearch") as MockES:
        mock_instance = MagicMock()
        mock_instance.ping.return_value = True
        MockES.return_value = mock_instance

        client = get_es_client()

        assert client is mock_instance


def test_get_es_client_returns_none_when_ping_fails():
    with patch("app.search.elastic.Elasticsearch") as MockES:
        mock_instance = MagicMock()
        mock_instance.ping.return_value = False
        MockES.return_value = mock_instance

        client = get_es_client()

        assert client is None


def test_get_es_client_returns_none_when_exception_raised():
    with patch("app.search.elastic.Elasticsearch") as MockES:
        MockES.side_effect = Exception("connection refused")

        client = get_es_client()

        assert client is None


def test_get_es_client_uses_env_url():
    with patch("app.search.elastic.ELASTIC_URL", "http://custom-es:9300"), \
         patch("app.search.elastic.Elasticsearch") as MockES:
        mock_instance = MagicMock()
        mock_instance.ping.return_value = True
        MockES.return_value = mock_instance

        get_es_client()

        MockES.assert_called_once_with("http://custom-es:9300", request_timeout=3)


def test_get_es_client_uses_default_url_when_env_not_set():
    with patch("app.search.elastic.ELASTIC_URL", "http://elasticsearch:9200"), \
         patch("app.search.elastic.Elasticsearch") as MockES:
        mock_instance = MagicMock()
        mock_instance.ping.return_value = True
        MockES.return_value = mock_instance

        get_es_client()

        MockES.assert_called_once_with("http://elasticsearch:9200", request_timeout=3)



def test_ensure_index_creates_index_when_not_exists():
    client = make_client()
    client.indices.exists.return_value = False

    ensure_index(client)

    client.indices.create.assert_called_once_with(
        index=INDEX_NAME, body=INDEX_MAPPING
    )


def test_ensure_index_does_not_create_when_already_exists():
    client = make_client()
    client.indices.exists.return_value = True

    ensure_index(client)

    client.indices.create.assert_not_called()


def test_ensure_index_checks_correct_index_name():
    client = make_client()
    client.indices.exists.return_value = True

    ensure_index(client)

    client.indices.exists.assert_called_once_with(index=INDEX_NAME)
    assert INDEX_NAME == "trains"



def test_index_train_calls_client_index_with_correct_args():
    client = make_client()

    index_train(client, SAMPLE_TRAIN)

    client.index.assert_called_once_with(
        index=INDEX_NAME,
        id=SAMPLE_TRAIN["id"],
        document=SAMPLE_TRAIN,
    )


def test_index_train_uses_train_id_as_document_id():
    client = make_client()
    train = {**SAMPLE_TRAIN, "id": 42}

    index_train(client, train)

    _, kwargs = client.index.call_args
    assert kwargs["id"] == 42



def test_bulk_index_trains_does_nothing_when_list_is_empty():
    client = make_client()

    bulk_index_trains(client, [])

    client.bulk.assert_not_called()


def test_bulk_index_trains_calls_bulk_with_action_pairs():
    client = make_client()
    trains = [
        {**SAMPLE_TRAIN, "id": 1},
        {**SAMPLE_TRAIN, "id": 2, "train_number": "99999"},
    ]

    bulk_index_trains(client, trains)

    client.bulk.assert_called_once()
    _, kwargs = client.bulk.call_args
    ops = kwargs["operations"]

    assert len(ops) == 4
    assert ops[0] == {"index": {"_index": INDEX_NAME, "_id": 1}}
    assert ops[1] == trains[0]
    assert ops[2] == {"index": {"_index": INDEX_NAME, "_id": 2}}
    assert ops[3] == trains[1]


def test_bulk_index_trains_uses_refresh_true():
    client = make_client()

    bulk_index_trains(client, [SAMPLE_TRAIN])

    _, kwargs = client.bulk.call_args
    assert kwargs["refresh"] is True


def test_bulk_index_trains_single_train():
    client = make_client()

    bulk_index_trains(client, [SAMPLE_TRAIN])

    _, kwargs = client.bulk.call_args
    ops = kwargs["operations"]
    assert len(ops) == 2



def test_update_available_seats_calls_update_with_correct_args():
    client = make_client()

    update_available_seats(client, train_id=1, available_seats=55)

    client.update.assert_called_once_with(
        index=INDEX_NAME,
        id=1,
        doc={"available_seats": 55},
    )


def test_update_available_seats_handles_not_found_gracefully():
    client = make_client()
    meta = MagicMock()
    meta.status = 404
    client.update.side_effect = NotFoundError(
        message="not found", meta=meta, body={}
    )

    update_available_seats(client, train_id=999, available_seats=1)


def test_update_available_seats_sets_zero_seats():
    client = make_client()

    update_available_seats(client, train_id=5, available_seats=0)

    _, kwargs = client.update.call_args
    assert kwargs["doc"] == {"available_seats": 0}



def make_search_response(sources: list[dict]) -> dict:
    return {"hits": {"hits": [{"_source": s} for s in sources]}}


def test_search_trains_returns_list_of_sources():
    client = make_client()
    client.search.return_value = make_search_response([SAMPLE_TRAIN])

    results = search_trains(client, source="Chennai", destination="Mumbai")

    assert results == [SAMPLE_TRAIN]


def test_search_trains_returns_empty_list_when_no_hits():
    client = make_client()
    client.search.return_value = make_search_response([])

    results = search_trains(client, source="Chennai", destination="Mumbai")

    assert results == []


def test_search_trains_passes_source_and_destination_in_query():
    client = make_client()
    client.search.return_value = make_search_response([])

    search_trains(client, source="Delhi", destination="Kolkata")

    _, kwargs = client.search.call_args
    body = kwargs["body"]
    must_clauses = body["query"]["bool"]["must"]
    assert must_clauses[0]["multi_match"]["query"] == "Delhi"
    assert must_clauses[1]["multi_match"]["query"] == "Kolkata"


def test_search_trains_uses_fuzziness_auto():
    client = make_client()
    client.search.return_value = make_search_response([])

    search_trains(client, source="Chenai", destination="Mumbay")  # intentional typos

    _, kwargs = client.search.call_args
    must_clauses = kwargs["body"]["query"]["bool"]["must"]
    assert must_clauses[0]["multi_match"]["fuzziness"] == "AUTO"
    assert must_clauses[1]["multi_match"]["fuzziness"] == "AUTO"


def test_search_trains_sorts_by_departure_time_ascending():
    client = make_client()
    client.search.return_value = make_search_response([])

    search_trains(client, source="Chennai", destination="Mumbai")

    _, kwargs = client.search.call_args
    assert kwargs["body"]["sort"] == [{"departure_time": {"order": "asc"}}]


def test_search_trains_limits_to_50_results():
    client = make_client()
    client.search.return_value = make_search_response([])

    search_trains(client, source="Chennai", destination="Mumbai")

    _, kwargs = client.search.call_args
    assert kwargs["body"]["size"] == 50


def test_search_trains_returns_multiple_results():
    client = make_client()
    train2 = {**SAMPLE_TRAIN, "id": 2, "train_number": "99999"}
    client.search.return_value = make_search_response([SAMPLE_TRAIN, train2])

    results = search_trains(client, source="Chennai", destination="Mumbai")

    assert len(results) == 2
    assert results[0]["id"] == 1
    assert results[1]["id"] == 2


def test_search_trains_uses_correct_index():
    client = make_client()
    client.search.return_value = make_search_response([])

    search_trains(client, source="Chennai", destination="Mumbai")

    _, kwargs = client.search.call_args
    assert kwargs["index"] == INDEX_NAME


def test_search_all_trains_returns_all_sources():
    client = make_client()
    train2 = {**SAMPLE_TRAIN, "id": 2}
    client.search.return_value = make_search_response([SAMPLE_TRAIN, train2])

    results = search_all_trains(client)

    assert len(results) == 2


def test_search_all_trains_uses_match_all_query():
    client = make_client()
    client.search.return_value = make_search_response([])

    search_all_trains(client)

    _, kwargs = client.search.call_args
    assert kwargs["body"]["query"] == {"match_all": {}}


def test_search_all_trains_returns_empty_list_when_no_trains():
    client = make_client()
    client.search.return_value = make_search_response([])

    results = search_all_trains(client)

    assert results == []


def test_search_all_trains_sorts_by_departure_time_ascending():
    client = make_client()
    client.search.return_value = make_search_response([])

    search_all_trains(client)

    _, kwargs = client.search.call_args
    assert kwargs["body"]["sort"] == [{"departure_time": {"order": "asc"}}]


def test_search_all_trains_limits_to_200_results():
    client = make_client()
    client.search.return_value = make_search_response([])

    search_all_trains(client)

    _, kwargs = client.search.call_args
    assert kwargs["body"]["size"] == 200


def test_search_all_trains_uses_correct_index():
    client = make_client()
    client.search.return_value = make_search_response([])

    search_all_trains(client)

    _, kwargs = client.search.call_args
    assert kwargs["index"] == INDEX_NAME