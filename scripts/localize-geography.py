# /// script
# dependencies = ["opencc-python-reimplemented==0.1.7"]
# ///
"""Pilot and apply English/Traditional Chinese names from GeoNames alternateNamesV2."""
import argparse
import collections
import gzip
import hashlib
import io
import json
from pathlib import Path
import re
import zipfile
from opencc import OpenCC

root=Path(__file__).resolve().parents[1]
source=root/'.remote-sync/geodata'
target=root/'internal/geography/data'
parser=argparse.ArgumentParser()
parser.add_argument('--phase',choices=['pilot','full'],required=True)
args=parser.parse_args()
converter=OpenCC('s2twp')
languages={'en','zh','zh-cn','zh-hans','zh-tw','zh-hant','zh-hk'}
priority={'zh-tw':0,'zh-hant':1,'zh-hk':2,'zh':3,'zh-cn':4,'zh-hans':4}
han=re.compile('[\u3400-\u9fff]')
admins={}
for line in (source/'admin1CodesASCII.txt').read_text(encoding='utf-8').splitlines():
    c=line.split('\t');admins[c[0]]=int(c[3])
city_admin={};wanted=set(admins.values())
with zipfile.ZipFile(source/'cities15000.zip') as z:
    for line in z.read('cities15000.txt').decode('utf-8').splitlines():
        c=line.split('\t');ident=int(c[0]);wanted.add(ident);city_admin[ident]=admins.get(c[8]+'.'+c[10])
archive=source/'alternateNamesV2.zip'
digest=hashlib.file_digest(archive.open('rb'),'sha256').hexdigest()
wanted_digest=hashlib.sha256(','.join(map(str,sorted(wanted))).encode()).hexdigest()
cache=source/'localized-name-index.json'
index=None
if cache.exists():
    candidate=json.loads(cache.read_text(encoding='utf-8'))
    if candidate['source_sha256']==digest and candidate.get('wanted_sha256')==wanted_digest:index=candidate
if index is None:
    records=collections.defaultdict(list)
    with zipfile.ZipFile(archive) as z:
        name=next(n for n in z.namelist() if n.endswith('alternateNamesV2.txt'))
        for line in io.TextIOWrapper(z.open(name),encoding='utf-8'):
            c=line.rstrip('\n').split('\t')
            if len(c)<8 or c[2].lower() not in languages:continue
            ident=int(c[1])
            if ident not in wanted or c[7]=='1' or len(c)>9 and c[9]:continue
            language=c[2].lower()
            if language!='en' and not han.search(c[3]):continue
            records[str(ident)].append([language,c[3],c[4]=='1',c[5]=='1',c[6]=='1'])
    index={'source_sha256':digest,'wanted_sha256':wanted_digest,'records':dict(records)}
    cache.write_text(json.dumps(index,ensure_ascii=True,separators=(',',':')),encoding='utf-8')

def select_name(ident,default):
    candidates=index['records'].get(str(ident),[])
    english=sorted((c for c in candidates if c[0]=='en'),key=lambda c:(not c[2],c[4],not c[3],c[1]!=default,c[1]))
    chinese=sorted((c for c in candidates if c[0]!='en'),key=lambda c:(priority[c[0]],not c[2],c[4],not c[3],c[1]))
    names={'en':english[0][1] if english else default}
    aliases={default,names['en']}
    provenance='original'
    if chinese:
        top=chinese[0];names['zh-TW']=converter.convert(top[1]);provenance=top[0]
    for c in candidates:
        aliases.add(c[1])
        if c[0]!='en':aliases.add(converter.convert(c[1]))
    return names,sorted(aliases),provenance

def localized(country):
    rows=json.loads(gzip.decompress((target/'cities'/(country+'.json.gz')).read_bytes()))
    for city in rows:
        names,aliases,origin=select_name(city['id'],city['name'])
        admin_names,admin_aliases,_=select_name(city_admin.get(city['id']),city['admin'])
        city['names']=names;city['adminNames']=admin_names
        city['searchNames']=sorted(set(aliases+admin_aliases+[city['asciiName']]))
        city['nameSource']=origin
    return rows

pilot_countries=['TW','JP','US','DE','FR']
pilot={code:localized(code) for code in pilot_countries}
# Fixed real-world records plus an explicit no-translation fallback case.
expected={1668341:'\u5317',1850147:'\u4eac',5128581:'\u7d04',2950159:'\u6797',2988507:'\u5df4'}
all_pilot={c['id']:c for rows in pilot.values() for c in rows}
for ident,part in expected.items():
    city=all_pilot[ident]
    assert part in city['names'].get('zh-TW',''), (ident,city['names'])
    assert city['names']['zh-TW'] in city['searchNames']
    assert city['names']['en'] in city['searchNames']
fallback=select_name(-1,'Untranslated locality')
assert fallback[0]=={'en':'Untranslated locality'}
report={'source_sha256':digest,'pilot_passed':True,'pilot_countries':pilot_countries,
        'pilot_examples':[{k:all_pilot[i][k] for k in ('id','names','nameSource')} for i in expected],
        'conversion':'OpenCC s2twp, opencc-python-reimplemented 0.1.7; build time only',
        'phase':args.phase}
(source/'localization-pilot.json').write_text(json.dumps(report,ensure_ascii=True,indent=2)+'\n',encoding='utf-8')
if args.phase=='full':
    summary=collections.Counter();coverage={}
    for file in sorted((target/'cities').glob('*.json.gz')):
        code=file.name[:2];rows=localized(code)
        coverage[code]={'total':len(rows),'translated':sum('zh-TW' in c['names'] for c in rows)}
        for city in rows:summary[city['nameSource']]+=1
        file.write_bytes(gzip.compress(json.dumps(rows,ensure_ascii=True,separators=(',',':')).encode(),mtime=0))
    report.update({'total':sum(c['total'] for c in coverage.values()),'translated':sum(c['translated'] for c in coverage.values()),
                   'coverage':coverage,'name_sources':dict(summary),'compressed_bytes':sum(p.stat().st_size for p in (target/'cities').glob('*.gz'))})
    manifest=json.loads((target/'manifest.json').read_text())
    manifest['localization']={k:report[k] for k in ('total','translated','conversion','name_sources')}
    manifest['source_sha256']['alternateNamesV2.zip']=digest
    (target/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n',encoding='utf-8')
    (source/'localization-full.json').write_text(json.dumps(report,ensure_ascii=True,indent=2)+'\n',encoding='utf-8')
print(json.dumps({k:v for k,v in report.items() if k!='coverage'},ensure_ascii=True,indent=2))
