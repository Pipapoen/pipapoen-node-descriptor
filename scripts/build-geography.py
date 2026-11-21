"""Build a compact GeoNames cities15000 snapshot from reviewed source downloads."""
import collections
import datetime
import gzip
import hashlib
import json
from pathlib import Path
import zipfile

root=Path(__file__).resolve().parents[1]
source=root/'.remote-sync/geodata'
target=root/'internal/geography/data'
(target/'cities').mkdir(parents=True,exist_ok=True)
countries=[]
for line in (source/'countryInfo.txt').read_text(encoding='utf-8').splitlines():
    if not line or line.startswith('#'):continue
    c=line.split('\t')
    if len(c[0])==2:countries.append({'code':c[0],'name':c[4]})
countries.sort(key=lambda c:c['name'])
admins={}
for line in (source/'admin1CodesASCII.txt').read_text(encoding='utf-8').splitlines():
    c=line.split('\t');admins[c[0]]=c[1]
grouped=collections.defaultdict(list)
with zipfile.ZipFile(source/'cities15000.zip') as z:
    for line in z.read('cities15000.txt').decode('utf-8').splitlines():
        c=line.split('\t')
        if c[6]!='P' or len(c[8])!=2:continue
        grouped[c[8]].append({'id':int(c[0]),'name':c[1],'asciiName':c[2],
                             'admin':admins.get(c[8]+'.'+c[10],c[10]),'feature':c[7],
                             'population':int(c[14] or 0)})
known={c['code'] for c in countries}
assert set(grouped)<=known
index={str(city['id']):code for code,rows in grouped.items() for city in rows}
(target/'index.json').write_text(json.dumps(index,separators=(',',':'))+'\n',encoding='utf-8')
for country in countries:
    rows=sorted(grouped[country['code']],key=lambda c:(-c['population'],c['name'],c['id']))
    country['cities']=len(rows)
    raw=json.dumps(rows,separators=(',',':'),ensure_ascii=True).encode()
    (target/'cities'/(country['code']+'.json.gz')).write_bytes(gzip.compress(raw,mtime=0))
(target/'countries.json').write_text(json.dumps(countries,separators=(',',':'))+'\n',encoding='utf-8')
manifest={'source':'GeoNames cities15000','snapshot':'2026-11-20','license':'CC BY 4.0',
          'countries':len(countries),'cities':sum(map(len,grouped.values())),
          'coverage':'Cities with population greater than 15000 or capitals; not every settlement.',
          'url':'https://download.geonames.org/export/dump/',
          'source_sha256':{n:hashlib.sha256((source/n).read_bytes()).hexdigest() for n in ('countryInfo.txt','admin1CodesASCII.txt','cities15000.zip')}}
(target/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n',encoding='utf-8')
print(json.dumps({**{k:manifest[k] for k in ('countries','cities')},'compressed_bytes':sum(p.stat().st_size for p in (target/'cities').glob('*.gz'))}))
