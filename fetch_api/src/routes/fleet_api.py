import requests, os
from fastapi import APIRouter, HTTPException


FLEET_API_URL = 'https://fleet-api.prd.eu.vn.cloud.tesla.com'
FLEET_AUTH_URL = 'https://fleet-auth.prd.vn.cloud.tesla.com/oauth2/v3/token'
AUDIENCE = FLEET_API_URL

ACCESS_TOKEN_FILE = '/app/fleet-api/access_token'
REFRESH_TOKEN_FILE = '/app/fleet-api/refresh_token'


router = APIRouter()


def save_tokens(tokens: dict) -> None:
    with open(ACCESS_TOKEN_FILE, 'w') as f:
        f.write(tokens['access_token'])

    with open(REFRESH_TOKEN_FILE, 'w') as f:
        f.write(tokens['refresh_token'])


def get_access_token() -> str:
    if os.path.exists(ACCESS_TOKEN_FILE):
        return open(ACCESS_TOKEN_FILE).read().strip()

    return refresh_access_token()


def refresh_access_token() -> str:
    if not os.path.exists(REFRESH_TOKEN_FILE):
        raise HTTPException(400, 'Missing refresh_token. Re-login required.')

    refresh_token = open(REFRESH_TOKEN_FILE).read().strip()

    data = {
        'grant_type': 'refresh_token',
        'client_id': os.getenv('CLIENT_ID'),
        'refresh_token': refresh_token
    }

    resp = requests.post(FLEET_AUTH_URL, data=data)

    if resp.status_code != 200:
        raise HTTPException(resp.status_code, resp.text)

    tokens = resp.json()
    save_tokens(tokens)

    return tokens['access_token']


def tesla_get(endpoint: str) -> dict:
    access_token = get_access_token()

    headers = {
        'Authorization': f'Bearer {access_token}',
        'X-Tesla-User-Agent': 'fetch-api',
    }

    url = f'{FLEET_API_URL}{endpoint}'
    resp = requests.get(url, headers=headers)

    if resp.status_code == 401:
        new_token = refresh_access_token()
        headers['Authorization'] = f'Bearer {new_token}'
        resp = requests.get(url, headers=headers)

    if resp.status_code != 200:
        raise HTTPException(resp.status_code, resp.text)

    return resp.json()


@router.get('/callback', tags=['side-stuff'], summary='OAuth2 callback listener that updates access/refresh tokens')
def update_tokens(code: str) -> dict:
    data = {
        'grant_type': 'authorization_code',
        'client_id': os.getenv('CLIENT_ID'),
        'client_secret': os.getenv('CLIENT_SECRET'),
        'code': code,
        'redirect_uri': 'https://fleet.car.k8s.iaminyourpc.xyz/car/callback',
        'audience': AUDIENCE,
    }

    response = requests.post(FLEET_AUTH_URL, data=data)

    if response.status_code != 200:
        raise HTTPException(response.status_code, response.text)

    tokens = response.json()
    save_tokens(tokens)

    return {'status': 'ok'}


@router.get('/list/vehicles', tags=['side-stuff'], summary='List vehicles')
def list_vehicles() -> dict:
    return tesla_get('/api/1/vehicles')
