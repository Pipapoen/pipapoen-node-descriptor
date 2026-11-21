"""Verify that shipped Go and frontend runtime notices match installed sources."""
import subprocess
from pathlib import Path

root = Path(__file__).resolve().parents[1]
lines = subprocess.check_output(['go','list','-deps','-f','{{if .Module}}{{.Module.Path}}|{{.Module.Dir}}{{end}}','./cmd/pnd'],cwd=root,text=True)
modules = dict(line.split('|',1) for line in lines.splitlines() if '|' in line)
modules.pop('github.com/Pipapoen/pipapoen-node-descriptor',None)
notices = root/'third_party/licenses'
for name, directory in modules.items():
    files = [p for p in Path(directory).iterdir() if p.is_file() and p.name.upper() in ('LICENSE','LICENSE.MD','LICENSE.TXT','NOTICE','NOTICE.TXT','COPYING')]
    assert files, 'Missing upstream license: '+name
    for path in files:
        shipped=notices/(name.replace('/','_')+'_'+path.name)
        assert shipped.read_bytes()==path.read_bytes(), 'Update shipped notice: '+str(shipped)
assert (notices/'basecoat_LICENSE').read_bytes()==(root/'node_modules/basecoat-css/LICENSE.md').read_bytes()
for name in ('go_LICENSE','tailwindcss_LICENSE','geonames_NOTICE.txt','dbip_NOTICE.txt'):
    assert (notices/name).stat().st_size>0, 'Missing attribution: '+name
print(f'Verified notices for {len(modules)} Go modules, frontend runtime and geographic data.')
