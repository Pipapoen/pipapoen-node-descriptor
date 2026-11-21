"""Collect notices for Go modules actually linked into the application."""
from pathlib import Path
import shutil

root=Path(__file__).resolve().parents[1]
dest=root/'third_party'/'licenses'
dest.mkdir(parents=True,exist_ok=True)
entries=[]
for line in (root/'build/runtime-modules.txt').read_text(encoding='utf-8-sig').splitlines():
    if '|' not in line: continue
    name,folder=line.split('|',1)
    if name=='github.com/Pipapoen/pipapoen-node-descriptor': continue
    source=Path(folder)
    matches=[p for p in source.iterdir() if p.is_file() and p.name.upper() in ('LICENSE','LICENSE.MD','LICENSE.TXT','NOTICE','NOTICE.TXT','COPYING')]
    if not matches: raise RuntimeError('Missing license: '+name)
    for p in matches:
        filename=name.replace('/','_')+'_'+p.name
        shutil.copyfile(p,dest/filename)
        entries.append((name,filename))
goroot=Path((root/'build/goroot.txt').read_text(encoding='utf-8-sig').strip())
shutil.copyfile(goroot/'LICENSE',dest/'go_LICENSE')
lines=['# Third-Party Notices','','PND source is licensed under MIT. Dependencies retain their respective licenses.','',
       'The following license and notice files accompany the application runtime:','']
for name,filename in entries: lines.append(f'- [{name}](third_party/licenses/{filename})')
lines+=['- [Go runtime and standard library](third_party/licenses/go_LICENSE)','',
        'The frontend includes selected [Basecoat](https://basecoatui.com) components by Ronan Berder under the [MIT license](third_party/licenses/basecoat_LICENSE). Unused selectors are removed at build time. Component-library JavaScript is not shipped. Frontend build tools and browser testing dependencies are pinned in `pnpm-lock.yaml` and retain their own package licenses.','',
        'The precompiled component stylesheet retains the [Tailwind CSS MIT license](third_party/licenses/tailwindcss_LICENSE).','',
        'Resource suggestions derive from public DigitalOcean, Akamai/Linode, Vultr, and AWS Lightsail plan specifications collected on 2026-11-20. See the source links and normalization notes in [the catalog evaluation](docs/RESOURCE_DICTIONARY_EVALUATION.md). Catalog suggestions are snapshots.','',
        'The bundled GeoNames country, administrative-area, and cities15000 snapshot uses CC BY 4.0; see [attribution and transformations](third_party/licenses/geonames_NOTICE.txt). Optional live GeoNames lookup uses the same source attribution. Optional GeoLite2 data is supplied separately by the operator under MaxMind terms; no GeoIP database is bundled.','']
lines+=['The optional account-free updater downloads DB-IP Lite City under CC BY 4.0. See [DB-IP attribution and processing](third_party/licenses/dbip_NOTICE.txt). The database is stored in a separate data volume and retains its source attribution.','']
(root/'THIRD_PARTY_NOTICES.md').write_text('\n'.join(lines),encoding='utf-8')
print('Collected',len(entries)+1,'runtime license and notice files')
