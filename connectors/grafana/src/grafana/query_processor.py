from common.telemetry.src.tracing.wrappers import traced
from common.telemetry.src.tracing.helpers import reword
from common.utils.helpers import TimeUtils, DataUtils



class Processor:
    @staticmethod
    @traced('map fields')
    def map(query_id: str, query_response: dict, span=None) -> dict | list:
        frames = query_response['results']['query'].get('frames', [])

        if not frames:
            return [] if query_id in ['teslamate-car-drives-info'] else {}

        frame = frames[0]
        fields = frame['schema']['fields']
        values = frame['data']['values']

        if query_id in [
            'teslamate-car-drives-info'
        ]:
            field_names = [field['name'] for field in fields]
            rows = zip(*values)
            result = [
                dict(zip(field_names, row))
                for row in rows
            ]

        else:
            result = {}

            for index, field in enumerate(fields):
                result[field['name']] = values[index][0]

        return result


    @staticmethod
    @traced('rename fields')
    def rename(query_id: str, data: dict | list, span=None) -> dict | list:
        reword_map = {
            'teslamate-last-charge-info': {
                'date': 'last_charge',
                'type': 'charge_type',
                'energy_added': 'charge_energy_added_kwh',
                'start_percent': 'charge_start_percentage',
                'end_percent': 'charge_end_percentage',
                'duration_str': 'duration',
                'duration_min': 'duration_minutes'
            },
            'teslamate-last-seen-location': {
                'time': 'last_seen'
            },
            'teslamate-car-state': {
                'since': 'last_updated'
            },
            'teslamate-car-efficiency': {
                'wh_per_km': 'average_consumption_wh_per_km',
                'driving_efficiency_pct': 'average_driving_efficiency_percentage',
                'usable_kwh': 'full_charge_usable_battery_kwh',
                'real_world_range_km': 'full_charge_usable_range_km'
            },
            'teslamate-car-drives-info': {
                'start_date_ts': 'start_time',
                'end_date_ts': 'end_time',
                'drive_id': 'id',
                'duration_str': 'duration',
                'duration_min': 'duration_minutes',
                '% Start': 'start_battery_percentage',
                '% End': 'end_battery_percentage',
                'outside_temp_c': 'outside_temperature_celsius',
                'speed_avg_km': 'average_speed_kmh',
                'speed_max_km': 'max_speed_kmh',
                'power_max': 'max_power_kw',
                'has_reduced_range': 'reduced_range',
                'efficiency': 'average_driving_efficiency_percentage',
                'consumption_kwh': 'total_consumption_kwh',
                'consumption_kwh_km': 'average_consumption_wh_per_km',
                'start_path': 'start_address_url',
                'end_path': 'end_address_url'
            }
        }

        if not query_id in reword_map:
            return data

        if query_id in [
            'teslamate-car-drives-info'
        ]:
            result = [
                {
                    reword_map[query_id].get(key, key): value
                    for key, value in item.items()
                }
                for item in data
            ]

        else:
            result = {
                reword_map[query_id].get(key, key): value
                for key, value in data.items()
            }

        return result


    @staticmethod
    @traced('drop unwanted fields')
    def drop(query_id: str, data: dict | list, span=None) -> dict | list:
        drop_map = {
            'teslamate-car-drives-info': [
                'car_id',
                'start_date'
            ]
        }

        if not query_id in drop_map:
            return data

        if query_id in [
            'teslamate-car-drives-info'
        ]:
            for item in data:
                for key in drop_map[query_id]:
                    item.pop(key, None)

        else:
            for key in drop_map[query_id]:
                data.pop(key, None)

        return data


    @staticmethod
    @traced('process response')
    def process(query_id: str, query_response: dict, span=None) -> dict:
        result = []

        if query_id.startswith('teslamate-'):
            data = Processor.map(query_id, query_response)
            data = Processor.drop(query_id, data)
            data = Processor.rename(query_id, data)

            if len(data) > 0:
                if query_id == 'teslamate-usable-battery-level':
                    data['usable_battery_kwh'] = round(data['usable_battery_kwh'], 2)

                elif query_id == 'teslamate-last-charge-info':
                    data['last_charge'] = TimeUtils.time_beautify_ms(data['last_charge'])
                    data['last_charge_since'] = TimeUtils.time_since(data['last_charge'])
                    data['charge_energy_added_percentage'] = (data['charge_end_percentage'] - data['charge_start_percentage'])
                    data['duration_minutes'] = round(data['duration_minutes'], 0)
                    data['duration_end_since_start'] = TimeUtils.time_since_minutes_only(data['duration_minutes'])

                elif query_id == 'teslamate-last-seen-location':
                    data['last_seen'] = TimeUtils.time_beautify_ms(data['last_seen'])
                    data['last_seen_since'] = TimeUtils.time_since(data['last_seen'])

                elif query_id == 'teslamate-car-state':
                    data['last_updated'] = TimeUtils.time_beautify_ms(data['last_updated'])
                    data['last_state_since'] = TimeUtils.time_since(data['last_updated'])

                if query_id in [
                    'teslamate-car-drives-info'
                ]:
                    if query_id == 'teslamate-car-drives-info':
                        for item in data:
                            try:
                                item['grafana_drive_url'] = DataUtils.get_teslamate_drive_grafana_url(
                                    drive_id = item['id'],
                                    drive_start_time = TimeUtils.time_beautify_ms(item['start_time'], convert_tz=False),
                                    drive_end_time = TimeUtils.time_beautify_ms(item['end_time'], convert_tz=False)
                                )
                            except:
                                item['grafana_drive_url'] = 'N/A'

                            for key in ['start_time', 'end_time']:
                                if item[key] is None:
                                    item[key] = 'N/A'
                                else:
                                    item[key] = TimeUtils.time_beautify_ms(item[key])
                                    item[f'{key}_ordinal'] = TimeUtils.time_beautify_ordinal(item[key])

                            for key in ['start_address_url', 'end_address_url']:
                                if item[key] is None:
                                    item[key] = 'N/A'
                                else:
                                    item[key] = DataUtils.get_maps_url(item[key])

                            try:
                                item['directions_url'] = DataUtils.get_maps_directions_url(
                                    item['start_address_url'],
                                    item['end_address_url']
                                )
                            except:
                                pass

                            for key in ['total_consumption_kwh', 'average_consumption_wh_per_km']:
                                if item[key] is None:
                                    item[key] = 'N/A'
                                else:
                                    item[key] = round(item[key], 2)

                            if item['distance_km'] is None:
                                item['distance_km'] = 'N/A'
                            else:
                                item['distance_km'] = round(item['distance_km'], 1)

                            if item['average_speed_kmh'] is None:
                                item['average_speed_kmh'] = 'N/A'
                            else:
                                item['average_speed_kmh'] = round(item['average_speed_kmh'], 0)

                            if item['average_driving_efficiency_percentage'] is None:
                                item['average_driving_efficiency_percentage'] = 'N/A'
                            else:
                                item['average_driving_efficiency_percentage'] = round((item['average_driving_efficiency_percentage'] * 100), 1)

                            try:
                                item['battery_used_percentage'] = (item['start_battery_percentage'] - item['end_battery_percentage'])
                            except:
                                pass

                            try:
                                item['duration_end_since_start'] = TimeUtils.time_since(
                                    item['start_time'],
                                    item['end_time'],
                                    instant=False
                                )
                            except:
                                pass

                    result = data

                else:
                    result.append(data)

        else:
            if query_id in ['argocd-apps']:
                for item in query_response['results']['query']['frames']:
                    labels = item['schema']['fields'][1]['labels']
                    result.append({**labels})

                result = sorted(
                    result,
                    key = lambda item: (
                        item['health_status'] == 'Healthy',
                        item['name'].lower()
                    )
                )

            elif query_id in ['longhorn-usage']:
                refined_result = {}

                for item in query_response['results']['query']['frames']:
                    labels = item['schema']['fields'][1]['labels']
                    value = item['data']['values'][1][0]

                    metric = labels['metric']
                    pvc_name = labels['pvc']
                    pvc_namespace = labels['pvc_namespace']
                    pvc_signature = (pvc_name, pvc_namespace)

                    if not pvc_signature in refined_result:
                        refined_result[pvc_signature] = {
                            'name': pvc_name,
                            'namespace': pvc_namespace
                        }

                    refined_result[pvc_signature][metric] = round(value, 1)

                result = list(refined_result.values())
                result = sorted(
                    result,
                    key = lambda item: item['pvc_usage_percentage'],
                    reverse = True
                )

        span.set_attributes(
            reword({
                'processor.result.total_items': len(result),
                'processor.result.items': result
            })
        )

        return {
            'total_items': len(result),
            'items': result
        }
