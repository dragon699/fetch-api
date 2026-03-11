from fetch_api.settings import (settings, connectors)
from fastapi import APIRouter, Response


router = APIRouter()


@router.get('/.well-known/appspecific/com.tesla.3p.public-key.pem', tags=['side-stuff'], summary='Public key needed for Tesla FleetAPI')
def get_public_key() -> Response:
    with open('/app/fleet-api/public/public-key.pem', 'rb') as f:
        key = f.read()
    return Response(content=key, media_type='application/x-pem-file')
