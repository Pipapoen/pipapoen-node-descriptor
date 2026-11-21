"""Scan the allowlisted public source snapshot without reading deployment secrets."""
import importlib.util
from pathlib import Path
import shutil
import subprocess
import tempfile

spec=importlib.util.spec_from_file_location('release_checks',Path(__file__).with_name('check-release.py'))
checks=importlib.util.module_from_spec(spec)
spec.loader.exec_module(checks)
checks.check()
(checks.ROOT/'build').mkdir(exist_ok=True)
with tempfile.TemporaryDirectory(prefix='public-scan-',dir=checks.ROOT/'build') as directory:
    for path in checks.source_files():
        target=Path(directory)/path.relative_to(checks.ROOT)
        target.parent.mkdir(parents=True,exist_ok=True)
        shutil.copyfile(path,target)
    subprocess.run(['go','run','github.com/zricethezav/gitleaks/v8@v8.30.1','dir',directory,'--redact','--no-banner'],check=True)
