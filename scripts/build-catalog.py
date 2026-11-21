"""Normalize reviewed provider facts into advisory resource tiers; preserve the codec dictionary."""
import argparse
from decimal import Decimal
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
CATALOG = ROOT / 'internal/catalog'
CPU = [1, 2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 256]
RAM = [.5, 1, 2, 3, 4, 6, 8, 12, 16, 24, 32, 64, 128, 256, 512]
DISK = [32, 64, 128, 256, 512, 1024, 2048]
UNITS = {'bytes': 1, 'MB': 10**6, 'GB': 10**9, 'TB': 10**12, 'MiB': 2**20, 'GiB': 2**30, 'TiB': 2**40}


def positive(value):
    result = Decimal(str(value))
    if not result.is_finite() or result <= 0:
        raise ValueError('Capacity must be finite and positive')
    return result


def gib(value, unit):
    return positive(value) * UNITS[unit] / Decimal(2**30)


def tier(value, limits):
    value = positive(value)
    return next((i for i, limit in enumerate(limits) if value <= Decimal(str(limit))), len(limits))


def number_text(value):
    text = format(positive(value), 'f')
    return text.rstrip('0').rstrip('.') if '.' in text else text


def display(value, unit):
    if unit in ('MiB', 'bytes'):
        return f'{number_text(gib(value, unit))} GiB'
    return f'{number_text(value)} {unit}'


def generate():
    raw = json.loads((CATALOG / 'source-plans.json').read_text(encoding='utf-8'))
    providers = json.loads((CATALOG / 'providers.json').read_text(encoding='utf-8'))
    ids = [p['id'] for p in providers]
    assert len(ids) == len(set(ids)) and all(0 <= i <= 65535 for i in ids)
    plans = []
    seen = set()
    for row in raw['plans']:
        identity = (row['provider'], row['name'])
        assert identity not in seen, f'Duplicate plan: {identity}'
        seen.add(identity)
        assert row['provider'] in ids and row['provider'] != 65535
        assert row.get('architecture') in (None, 0, 1, 2, 3) and row['nodeType'] in (0, 1, 2, 3, 7)
        assert row['kind'] in ('plan', 'configuration', 'example')
        source = raw['sources'][row['source']]
        assert source['url'].startswith('https://')
        plan = {key: row[key] for key in ('provider', 'name', 'family', 'architecture', 'nodeType', 'kind') if key in row}
        plan.update(cpu=tier(row['vcpu'], CPU), ram=tier(gib(row['memory'], row['memoryUnit']), RAM),
                    label=row['name'], source=source['url'], checkedAt=source['checkedAt'])
        specs = [f'{number_text(row["vcpu"])} vCPU', display(row['memory'], row['memoryUnit'])]
        if row.get('disk') is not None:
            plan['storage'] = tier(gib(row['disk'], row['diskUnit']), DISK)
            specs.append(display(row['disk'], row['diskUnit']))
        plan['specs'] = ' / '.join(specs)
        plans.append(plan)
    plans.sort(key=lambda p: (p['provider'], p['family'], p['cpu'], p['ram'], p.get('storage', -1), p['name']))
    info = {'checkedAt': raw['checkedAt'], 'providers': len(providers), 'plans': len(plans),
            'families': len({(p['provider'], p['family']) for p in plans})}
    return {'plans.json': plans, 'catalog-info.json': info}


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    for name, data in generate().items():
        text = json.dumps(data, ensure_ascii=True, indent=2) + '\n'
        path = CATALOG / name
        if args.check:
            assert path.read_text(encoding='utf-8') == text, f'Regenerate {name}: python scripts/build-catalog.py'
        else:
            path.write_text(text, encoding='utf-8')
    print('Catalog facts, unit conversion, provider references and generated files validated.')
