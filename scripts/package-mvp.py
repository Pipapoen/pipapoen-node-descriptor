"""Package the reviewed Linux MVP artifacts for the initial LAN deployment."""
from pathlib import Path
import shutil
import tarfile

root=Path(__file__).resolve().parents[1]
out=root/'deploy-bundle'
out.mkdir(exist_ok=True)
for name in ('pnd-linux','codec-linux.test','server-linux.test','geography-linux.test','geoipupdate-linux.test'):
    shutil.copyfile(root/'build'/name,out/name)
shutil.copyfile(root/'LICENSE',out/'LICENSE')
shutil.copyfile(root/'THIRD_PARTY_NOTICES.md',out/'THIRD_PARTY_NOTICES.md')
shutil.copyfile(root/'compose.geoip-auto.yaml',out/'compose.geoip-auto.yaml')
shutil.copytree(root/'third_party/licenses',out/'licenses',dirs_exist_ok=True)
shutil.copytree(root/'internal/codec/testdata',out/'testdata',dirs_exist_ok=True)
compose=(root/'compose.yaml').read_text(encoding='utf-8').replace('    build: .','    build:\n      context: .\n      dockerfile: Dockerfile.runtime')
(out/'compose.yaml').write_text(compose,encoding='utf-8')
(out/'Dockerfile.runtime').write_text('''FROM scratch
COPY ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY pnd-linux /pnd
COPY LICENSE THIRD_PARTY_NOTICES.md /licenses/
COPY licenses /licenses/dependencies
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/pnd"]
''',encoding='utf-8')
with tarfile.open(out/'pnd-mvp.tar.gz','w:gz') as tar:
    for name in ('pnd-linux','codec-linux.test','server-linux.test','geography-linux.test','geoipupdate-linux.test','testdata','LICENSE','THIRD_PARTY_NOTICES.md','licenses','compose.yaml','compose.geoip-auto.yaml','Dockerfile.runtime'):
        tar.add(out/name,arcname=name)
print('Deployment archive prepared from explicit artifact list')
