from typing import Any
from datetime import datetime, timezone
from common.utils.system import read_file
import re, zoneinfo, json, hashlib



class SysUtils:
    @staticmethod
    def get_app_version(version_file: str) -> str:
        try:
            return read_file(version_file, type=None).strip()

        except:
            print(f'Unable to determine app version, {version_file} not found')
            return 'unknown'


    @staticmethod
    def get_swagger_params() -> dict:
        return {
            'swagger_ui_parameters': {
                'docExpansion': 'list',
                'defaultModelsExpandDepth': -1,
                'defaultModelExpandDepth': 2,
                'displayRequestDuration': True,
                'syntaxHighlight': {
                    'theme': 'obsidian'
                },
                'filter': True,
                'tryItOutEnabled': True,
                'persistAuthorization': True,
                'displayOperationId': False,
                'deepLinking': True,
                'requestSnippetsEnabled': True
            }
        }



class TimeUtils:
    @staticmethod
    def time_beautify_ms(milliseconds: int, target_tz: str = 'Europe/Sofia', convert_tz: bool = True) -> str:
        seconds = milliseconds / 1000
        tz = zoneinfo.ZoneInfo(target_tz)
        
        dt_utc = datetime.fromtimestamp(seconds, tz=timezone.utc)

        if convert_tz:
            dt = dt_utc.astimezone(tz)
        else:
            dt = dt_utc

        return dt.strftime('%Y-%m-%dT%H:%M:%S')


    @staticmethod
    def time_beautify_ordinal(dt_string: str, target_tz: str = 'Europe/Sofia') -> str:
        def get_ordinal_day(n: int) -> str:
            if 10 <= n % 100 <= 20:
                suffix = 'th'
            else:
                suffix = {1: 'st', 2: 'nd', 3: 'rd'}.get(n % 10, 'th')
            return f'{n}{suffix}'
        
        tz = zoneinfo.ZoneInfo(target_tz)
        dt = datetime.fromisoformat(dt_string)

        if dt.tzinfo is None:
            dt = dt.replace(tzinfo=tz)

        day = get_ordinal_day(dt.day)
        month = dt.strftime('%B')
        year = dt.year
        time = dt.strftime('%H:%M')

        return f'{year} / {day} of {month} at {time}'


    @staticmethod
    def time_now(target_tz: str = 'Europe/Sofia') -> str:
        tz = zoneinfo.ZoneInfo(target_tz)
        now = datetime.now(tz)

        return now.strftime('%Y-%m-%dT%H:%M:%S')


    @staticmethod
    def time_since(past: str, future: str | None = None, tz: str = 'Europe/Sofia', instant: bool = True) -> str:
        tz = zoneinfo.ZoneInfo(tz)
        past_dt = datetime.fromisoformat(past)

        if past_dt.tzinfo is None:
            past_dt = past_dt.replace(tzinfo=tz)
        
        if future is None:
            future_dt = datetime.now(tz)
        
        else:
            future_dt = datetime.fromisoformat(future)

            if future_dt.tzinfo is None:
                future_dt = future_dt.replace(tzinfo=tz)

        diff = future_dt - past_dt
        seconds = diff.total_seconds()

        days = int(seconds // 86400)
        hours = int((seconds % 86400) // 3600)
        minutes = int((seconds % 3600) // 60)

        if days > 0:
            if days < 7:
                return f'{days}d{hours}h'
            else:
                if instant:
                    return f'{days}d'
                else:
                    return f'{days}d+'

        elif hours > 0:
            if hours < 24:
                return f'{hours}h{minutes}m'
            else:
                if instant:
                    return f'{hours}h'
                else:
                    return f'{hours}h+'

        elif minutes > 0:
            return f'{minutes}m'

        else:
            if instant:
                return 'just now'
            else:
                return '<1m'


    @staticmethod
    def time_since_minutes_only(minutes: int) -> str:
        if minutes < 1:
            return '<1m'

        hours = minutes // 60
        mins = minutes % 60

        if hours > 0:
            if mins > 0:
                return f'{int(hours)}h{int(mins)}m'

            return f'{int(hours)}h'

        return f'{int(mins)}m'



class DataUtils:
    @staticmethod
    def get_maps_url(path: str) -> str:
        if not path.startswith('new?lat='):
            return path
        
        try:
            lat = path.split('lat=')[1].split('&')[0]
            lng = path.split('lng=')[1]

            return f'https://google.com/maps/search/?api=1&query={lat},{lng}'
        
        except:
            return path


    @staticmethod
    def get_maps_directions_url(start_url: str, end_url: str) -> str:
        def extract_coords(u: str) -> str | None:
            if 'new?lat=' in u:
                try:
                    lat = u.split('lat=')[1].split('&')[0]
                    lng = u.split('lng=')[1]
                    return f'{lat},{lng}'
                except:
                    return None

            if 'query=' in u:
                try:
                    return u.split('query=')[1]
                except:
                    return None
            
            return None

        start = extract_coords(start_url)
        end = extract_coords(end_url)

        if not start or not end:
            return 'N/A'

        return f'https://www.google.com/maps/dir/?api=1&origin={start}&destination={end}'


    @staticmethod
    def get_teslamate_drive_grafana_url(drive_id: int, drive_start_time: str, drive_end_time: str) -> str:
        grafana_url = 'https://grafana.k8s.iaminyourpc.xyz'
        grafana_dashboard_path = 'd/zm7wN6Zgz/driving-details'

        return f'{grafana_url}/{grafana_dashboard_path}?from={drive_start_time}&to={drive_end_time}&var-drive_id={drive_id}&timezone=Europe%2FSofia&orgId=1&var-temp_unit=C&var-length_unit=km&var-alternative_length_unit=m&var-preferred_range=rated&var-base_url=https:%2F%2Fcar.k8s.iaminyourpc.xyz&var-pressure_unit=bar&var-speed_unit=km%2Fh'


    @staticmethod
    def omit_volatile_data(data: Any) -> Any:
        VOLATILE_PATTERN = re.compile(
            r'^('
            r'(<\d+m)'
            r'|(\d+d\d+h(\d+m)?)'
            r'|(\d+h\d+m)'
            r'|(\d+m)'
            r'|(\d+h)'
            r'|(just now)'
            r')$',
            re.IGNORECASE
        )

        if isinstance(data, dict):
            return {
                k: DataUtils.omit_volatile_data(v)
                for k, v in data.items()
            }

        if isinstance(data, list):
            return [
                DataUtils.omit_volatile_data(item)
                for item in data
            ]

        if isinstance(data, str):
            return '<volatile>' if VOLATILE_PATTERN.match(data.strip()) else data

        return data


    @staticmethod
    def create_cache_key(connector_name: str, method: str, endpoint: str, params: dict, data: dict) -> str:
        params_hash = hashlib.md5(
            json.dumps(
                params,
                sort_keys=True
            ).encode()
        ).hexdigest()

        data_hash = hashlib.md5(
            json.dumps(
                data,
                sort_keys=True
            ).encode()
        ).hexdigest()

        return (
            f'connector:{connector_name}:'
            f'{method}:{endpoint}:'
            f'params:{params_hash}:'
            f'data:{data_hash}'
        )
