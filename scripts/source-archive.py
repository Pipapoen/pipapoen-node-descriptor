"""Build an allowlisted source archive for clean container verification."""
import importlib.util
import tarfile
from pathlib import Path

spec = importlib.util.spec_from_file_location('release_checks', Path(__file__).with_name('check-release.py'))
checks = importlib.util.module_from_spec(spec)
spec.loader.exec_module(checks)
checks.check()
output = checks.ROOT / 'build/deploy-bundle/source.tar.gz'
output.parent.mkdir(parents=True,exist_ok=True)
with tarfile.open(output, 'w:gz') as archive:
    for path in sorted(checks.source_files()):
        archive.add(path, arcname=path.relative_to(checks.ROOT), recursive=False)
print('Prepared allowlisted source archive.')
