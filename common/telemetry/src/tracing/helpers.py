import json


def reword(attributes: dict) -> dict:
    reword_map = {
        'health.status.current': {True: 'healthy', False: 'unhealthy'},
        'health.status.previous': {True: 'healthy', False: 'unhealthy'},
        'health.connector.status.current': {True: 'healthy', False: 'unhealthy'},
        'health.connector.status.previous': {True: 'healthy', False: 'unhealthy'},
        'grafana.auth.status': {True: 'authenticated', False: 'unauthenticated'},
        '__redis.cache.value': 'json',
        '__processor.result.items': 'json',
        '__query.params': 'json',
        '__query.payload': 'json',
        '__query.response': 'json',
        '__request.params': 'json',
        '__request.body': 'json',
        '__response.content': 'json'
    }

    reworded = attributes.copy()
    discarded_keys = []

    for a_k, a_v in reworded.items():
        if isinstance(a_v, (list, dict)) and len(a_v) == 0:
            discarded_keys.append(a_k)
            continue

        if a_v is None:
            reworded[a_k] = 'unknown'

        elif (a_k in reword_map) and (a_v in reword_map[a_k]):
            reworded[a_k] = reword_map[a_k][a_v]

        for r_k, r_v in reword_map.items():
            if (r_k.startswith('__')) and (r_k[2:] in a_k or r_k[2:] == a_k) and (r_v == 'json'):
                if isinstance(a_v, str):
                    try:
                        reworded[a_k] = json.dumps(json.loads(a_v), indent=4)

                    except:
                        reworded[a_k] = json.dumps(a_v, indent=4)

                else:
                    reworded[a_k] = json.dumps(a_v, indent=4)

    for key in discarded_keys:
        reworded.pop(key, None)

    return reworded
