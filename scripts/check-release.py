"""Validate version consistency and the files selected for distribution."""
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def version():
    value = json.loads((ROOT / 'package.json').read_text(encoding='utf-8'))['version']
    if not re.fullmatch(r'\d{2}\.[1-9]\d*\.\d+(?:-rc\.[1-9]\d*)?', value):
        raise ValueError('Expected YY.release.patch or YY.release.patch-rc.N')
    return value


def source_files():
    folders = ('cmd', 'internal', 'web', 'demo', 'scripts', 'docs', 'deploy', 'third_party', '.github')
    for folder in folders:
        for path in (ROOT / folder).rglob('*'):
            relative = path.relative_to(ROOT)
            if relative.parts[:3] == ('demo', 'public', 'runtime'):
                continue
            if path.is_file() and 'assets' not in relative.parts and '__pycache__' not in relative.parts:
                if relative.as_posix() == 'deploy/.env':
                    continue
                yield path
    for path in ROOT.iterdir():
        if path.is_file() and path.name not in ('deployment.env',):
            yield path


def check():
    v = version()
    assert f'"{v}"' in (ROOT / 'internal/buildinfo/version.go').read_text(), 'Go version differs'
    for name in ('compose.yaml', 'deploy/overlays/compose.geoip-auto.yaml', 'deploy/compose.yaml', 'deploy/.env.example'):
        assert ':' + v in (ROOT / name).read_text(), f'Image version differs: {name}'
    patterns = [rb'-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----', rb'gh[pousr]_[A-Za-z0-9]{30,}', rb'github_pat_[A-Za-z0-9_]{30,}']
    for path in source_files():
        data = path.read_bytes()
        assert not any(re.search(pattern, data) for pattern in patterns), f'Credential pattern: {path.relative_to(ROOT)}'
        if path.suffix in ('.md', '.ts', '.yml', '.yaml', '.json', '.py', '.go'):
            data.decode('utf-8', errors='strict')
        if path.suffix == '.md':
            assert not re.search(r'[\u4e00-\u9fff]', data.decode('utf-8')), f'Non-English documentation: {path}'
    print(f'Validated version {v}, UTF-8, documentation language and common credential patterns.')


if __name__ == '__main__':
    check()
