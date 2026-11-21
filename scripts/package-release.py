"""Create native and Compose release archives with checksums."""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tarfile
import tempfile
import zipfile

root = Path(__file__).resolve().parents[1]
version = json.loads((root / 'package.json').read_text())['version']
output = root / 'build/release'
output.mkdir(parents=True, exist_ok=True)


def notices(folder):
    for name in ('LICENSE', 'THIRD_PARTY_NOTICES.md'):
        shutil.copyfile(root / name, folder / name)
    shutil.copytree(root / 'third_party', folder / 'third_party')
    shutil.copytree(root / 'docs', folder / 'docs')
    for name in ('internal/catalog/dictionary.json', 'internal/codec/testdata/packing.json',
                 'internal/codec/testdata/format0.json', 'internal/codec/testdata/format1.json'):
        target = folder / name
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(root / name, target)


archives = []
with tempfile.TemporaryDirectory() as directory:
    temp = Path(directory)
    for target in ('linux', 'windows'):
        folder = temp / target
        folder.mkdir()
        executable = folder / ('pnd.exe' if target == 'windows' else 'pnd')
        env = dict(os.environ, GOOS=target, GOARCH='amd64', CGO_ENABLED='0')
        subprocess.run(['go', 'build', '-trimpath', '-ldflags=-s -w', '-o', str(executable), './cmd/pnd'], cwd=root, env=env, check=True)
        notices(folder)
        archive = output / f'pnd-{version}-{target}-amd64.zip'
        with zipfile.ZipFile(archive, 'w', zipfile.ZIP_DEFLATED) as dest:
            for file in sorted(folder.rglob('*')):
                if file.is_file(): dest.write(file, file.relative_to(folder))
        archives.append(archive)
    folder = temp / 'compose'
    shutil.copytree(root / 'deploy', folder, ignore=shutil.ignore_patterns('.env', 'source.env.example', 'overlays'))
    for name in ('compose.geoip-auto.yaml', 'compose.oidc.yaml', 'compose.geoip.yaml'):
        text = (root / 'deploy/overlays' / name).read_text(encoding='utf-8').replace(f'pnd:{version}', '${PND_IMAGE:-ghcr.io/pipapoen/pipapoen-node-descriptor:' + version + '}')
        (folder / name).write_text(text, encoding='utf-8')
    notices(folder)
    archive = output / f'pnd-{version}-compose.zip'
    with zipfile.ZipFile(archive, 'w', zipfile.ZIP_DEFLATED) as dest:
        for file in sorted(folder.rglob('*')):
            if file.is_file(): dest.write(file, file.relative_to(folder))
    archives.append(archive)

(output / 'SHA256SUMS').write_text(''.join(hashlib.sha256(file.read_bytes()).hexdigest()+'  '+file.name+'\n' for file in archives), encoding='ascii')
print(f'Created {len(archives)} release archives and SHA256SUMS for {version}.')
