"""Exercise the release Compose contract in an isolated disposable project."""
import json
import os
import subprocess
import time
import urllib.request

project = 'pnd-ci-' + str(os.getpid())
os.environ['PND_PORT'] = '18080'
os.environ['PND_BIND_ADDRESS'] = '127.0.0.1'
os.environ['PND_PUBLIC_URL'] = 'http://127.0.0.1:18080'
command = ['docker', 'compose', '-f', 'deploy/compose.yaml', '-p', project]


def run(*args):
    subprocess.run(command + list(args), check=True)


def api(path, data=None):
    body = None if data is None else json.dumps(data).encode()
    request = urllib.request.Request('http://127.0.0.1:18080' + path, data=body,
                                     headers={'Content-Type': 'application/json'})
    with urllib.request.urlopen(request, timeout=5) as response:
        return json.load(response)


def ready():
    for _ in range(60):
        try:
            if api('/healthz')['status'] == 'ok':
                return
        except (OSError, ValueError):
            pass
        time.sleep(1)
    raise RuntimeError('Container readiness timed out')


try:
    run('up', '-d', '--wait', '--wait-timeout', '90')
    ready()
    meta = api('/api/meta')
    assert meta['prefix'] == 'x' and meta['domain'] == 'example.org'
    assert meta['format'] == '1'
    descriptor = {'nodeId': api('/api/node-id', {})['nodeId'], 'geo': 5128581}
    output = api('/api/encode', {'descriptor': descriptor})
    assert output['label'][1] == '1'
    assert api('/api/decode', {'input': output['hostname']}) == descriptor
    # Recreate both services to prove initialization preserves the original key.
    run('up', '-d', '--force-recreate', '--wait', '--wait-timeout', '90')
    ready()
    assert api('/api/decode', {'input': output['hostname']}) == descriptor
    print('Container startup, generic defaults, round trip and persistent key passed.')
finally:
    run('down', '-v', '--remove-orphans')
