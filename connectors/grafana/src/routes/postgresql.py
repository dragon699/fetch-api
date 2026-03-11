from connectors.grafana.src.api_processor import APIProcessor
from fastapi import APIRouter, Query


router = APIRouter()


@router.get('/car-battery', tags=['postgresql'], summary='Retrieve current car battery level from Teslamate')
def get_car_battery() -> dict:
    return APIProcessor.process_request(
        query_ds_type='postgresql',
        query_id='teslamate-usable-battery-level'
    )


@router.get('/car-last-charge', tags=['postgresql'], summary='Retrieve car last charge information from Teslamate')
def get_car_last_charge() -> dict:
    return APIProcessor.process_request(
        query_ds_type='postgresql',
        query_id='teslamate-last-charge-info'
    )


@router.get('/car-last-location', tags=['postgresql'], summary='Retrieve car last location from Teslamate')
def get_car_last_location() -> dict:
    return APIProcessor.process_request(
        query_ds_type='postgresql',
        query_id='teslamate-last-seen-location'
    )


@router.get('/car-state', tags=['postgresql'], summary='Retrieve car state from Teslamate')
def get_car_state() -> dict:
    return APIProcessor.process_request(
        query_ds_type='postgresql',
        query_id='teslamate-car-state'
    )


@router.get('/car-efficiency', tags=['postgresql'], summary='Retrieve car efficiency from Teslamate')
def get_car_average_consumption() -> dict:
    return APIProcessor.process_request(
        query_ds_type='postgresql',
        query_id='teslamate-car-efficiency'
    )


@router.get('/car-drives-history', tags=['postgresql'], summary='Retrieve car drives history from Teslamate')
def get_car_drives(number_of_drives: int = Query(5, ge=1, le=1000, description='How many recent drive records to retrieve')) -> dict:
    return APIProcessor.process_request(
        query_ds_type='postgresql',
        query_id='teslamate-car-drives-info',
        query_params={
            'number_of_drives': number_of_drives
        }
    )
